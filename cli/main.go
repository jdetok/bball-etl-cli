/*
BBALL ETL COMMAND LINE INTERFACE
PROJECT INTENT:
  - create a command line interface for the bball-etl-go package to eliminate
    the need for separate directories for the nightly/build runs of the etl

- exit program if no arguments are passed (args slices == 1)
- mode flag:
	- build: runs full etl all games 1970 through current
	- daily: runs etl for games from previous day
	- custom (not yet build): pass a season and league (optional) to run the etl
		for a specific season

- TODO:
	- dev / prod as an argument
	- custom etl
	- eventually, define flags for different endpoints
*/

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/jdetok/bball-etl-cli/etl"
	"github.com/jdetok/golib/errd"
	"github.com/jdetok/golib/pgresd"
)

func main() {
	e := errd.InitErr()
	var sTime time.Time = time.Now()
	var cnf etl.Conf
	var compMsg string // complete message - different based on args

	// get args passed - exit if 1 will be at least 2 if arg was passed
	runArgs := os.Args
	if len(runArgs) == 1 {
		e.Msg = "an argument must be passed"
		e.Fatal(e.NewErr())
	}

	// parse flags
	args := parseArgs()

	// init database based on -env flag
	var pg pgresd.PostGres
	db, envErr := args.SetEnv(&pg)
	if envErr != nil {
		e.Msg = "failed to setup the environment"
		e.Fatal(envErr)
	}

	// set etl.cnf database values
	cnf.DB = db
	cnf.RowCnt = 0

	// RUN APPROPRIATE ETL PROCESS BASED ON FLAGS
	switch args.Mode[1] {
	case "": // no mode passed,
		e.Msg = "a mode must be specified"
		e.Fatal(e.NewErr())

	// daily etl: etl for previous day's games
	case "daily":
		dlyCmpl, dlyErr := args.SetupDailyETL(&cnf, &e)
		if dlyErr != nil {
			e.Msg = dlyCmpl
			e.Fatal(dlyErr)
		}

		// build etl: all seasons 1970 through current
	case "build":
		bldCmpl, bldErr := args.SetupBuildETL(&cnf, "1970", "current")
		if bldErr != nil {
			e.Msg = bldCmpl
			e.Fatal(bldErr)
		}
		fmt.Println(bldCmpl)

		// "custom" run - a season MUST be specified, lg defaults to both
	case "custom":
		// exit if no season passed
		if args.Szn[1] == "" {
			e.Msg = "a season (-szn) must be specified in custom mode"
			e.Fatal(e.NewErr())
		}
		// switch on lg to determine whether to do both leagues or just one
		switch args.Lg[1] {
		case "":
			bCmpl, bErr := args.CustomBothLgETL(&cnf)
			if bErr != nil {
				e.Msg = bCmpl
				e.Fatal(bErr)
			}
		case "nba", "wnba":
			lgCmpl, lgErr := args.CustomLgETL(&cnf)
			if lgErr != nil {
				e.Msg = lgCmpl
				e.Fatal(lgErr)
			}
		}
	// EMAIL MODE: RUN AT END OF SH. must pas a log file
	case "email":
		switch args.Logf[1] {
		case "":
			e.Msg = "must pass a log file when run in email mode"
			e.Fatal(e.NewErr())
		default:
			if err := EmailLog(args.Logf[1]); err != nil {
				e.Msg = "error emailing log"
				e.Fatal(err)
			}
		}

	// NO ARGS PASSED - ERROR OUT
	default:
		e.Msg = fmt.Sprintf(
			"invalid mode: '%s' is not an option", args.Mode[1])
		e.Fatal(e.NewErr())
	}

	// write errors to the log
	if len(cnf.Errs) > 0 {
		cnf.L.WriteLog(fmt.Sprintln("ERRORS:"))
		for _, e := range cnf.Errs {
			cnf.L.WriteLog(fmt.Sprintln(e))
		}
	}

	// complete log
	cnf.L.WriteLog(
		fmt.Sprint(
			"process complete",
			fmt.Sprintf(
				"\n ---- start time: %v", sTime),
			fmt.Sprintf(
				"\n ---- cmplt time: %v", time.Now()),
			fmt.Sprintf(
				"\n ---- duration: %v", time.Since(sTime)),
			compMsg, // assigned in switch based on passed mode
		),
	)
}
