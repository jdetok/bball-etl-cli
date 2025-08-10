package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jdetok/bball-etl-cli/etl"
	"github.com/jdetok/golib/errd"
	"github.com/jdetok/golib/logd"
	"github.com/jdetok/golib/pgresd"
)

/*
PROJECT INTENT:
  - create a command line interface for the bball-etl-go package to eliminate
    the need for separate directories for the nightly/build runs of the etl
*/

/* TODO:
- exit program if no arguments are passed (args slices == 1)
- define a flag for mode:
	- build: runs full etl all games 1970 through current
	- daily: runs etl for games from previous day

- LATER:
	- define an additional flag for league
	- eventually, define flags for different endpoint endpoints
*/

type Params struct {
	Mode [2]string // run mode e.g. build, daily, etc
	Szn  [2]string // season selector, e.g. 2024 for 2024-25 NBA/2024 WNBA
	Lg   [2]string // league selector, nba or wnba
}

func parseArgs() Params {
	var p = Params{
		Mode: [2]string{"mode", ""},
		Szn:  [2]string{"szn", ""},
		Lg:   [2]string{"lg", ""},
	}
	flag.StringVar(&p.Mode[1], "mode", "", "etl run-mode")
	flag.StringVar(&p.Szn[1], "szn", "", "nba/wnba season e.g. 2024")
	flag.StringVar(&p.Lg[1], "lg", "", "nba or wnba")
	flag.Parse()
	return p
}

func main() {
	// configs
	e := errd.InitErr()
	var sTime time.Time = time.Now()
	var cnf etl.Conf
	var compMsg string // complete message - different based on args

	// get args passed - exit if 1 will be at least 2 if arg was passed
	runArgs := os.Args
	if len(runArgs) == 1 {
		e.Msg = "an argument must be passed"
		fmt.Println(e.NewErr())
		os.Exit(1)
	}

	// init database
	pg := pgresd.GetEnvPG()
	pg.MakeConnStr()
	db, err := pg.Conn()
	if err != nil {
		fmt.Println(e.BuildErr(err))
		os.Exit(1)
	}
	db.SetMaxOpenConns(40)
	db.SetMaxIdleConns(40)
	cnf.DB = db
	cnf.RowCnt = 0

	// parse through params
	var p Params = parseArgs()
	switch p.Mode[1] {
	case "": // no mode passed,
		e.Msg = "a mode must be specified"
		fmt.Println(e.NewErr())
		os.Exit(1)
	case "build":
		// build etl: all seasons 1970 through current
		// initialize logger with build log
		l, err := logd.InitLogger("z_log_b", "build_etl")
		if err != nil {
			e.Msg = "error initializing logger"
			log.Fatal(e.BuildErr(err))
		}
		cnf.L = l // assign to cnf

		// SET START AND END SEASONS
		var st string = "1970"
		var en string = time.Now().Format("2006") // current year

		// RUN ETL
		if err = etl.RunSeasonETL(cnf, st, en); err != nil {
			e.Msg = fmt.Sprintf(
				"error running season etl: start year: %s | end year: %s", st, en)
			cnf.L.WriteLog(e.Msg)
			log.Fatal(e.BuildErr(err))
		}
		compMsg = fmt.Sprintf(
			"\n---- etl for seasons between %s and %s | total rows affected: %d",
			st, en, cnf.RowCnt,
		)

	case "daily":
		// daily etl: etl for previous day's games
		// initialize logger with nightly log
		l, err := logd.InitLogger("z_log_n", "nightly_etl")
		if err != nil {
			e.Msg = "error initializing logger"
			log.Fatal(e.BuildErr(err))
		}
		cnf.L = l // assign to cnf

		// RUN NIGHTLY ETL
		if err = etl.RunNightlyETL(cnf); err != nil {
			e.Msg = fmt.Sprintf(
				"error with %v nightly etl", etl.Yesterday(time.Now()))
			cnf.L.WriteLog(e.Msg)
			log.Fatal(e.BuildErr(err))
		}
		compMsg = fmt.Sprintf( // assign in switch
			"\n---- nightly etl for %v complete | total rows affected: %d",
			etl.Yesterday(time.Now()), cnf.RowCnt,
		)

	case "custom": // "custom" run - a season MUST be specified, lg defaults to both
		fmt.Println("custom mode selected:", p.Mode)
		if p.Szn[1] == "" {
			e.Msg = "a season (-szn) must be specified in custom mode"
			fmt.Println(e.NewErr())
			os.Exit(1)
		}
		switch p.Lg[1] {
		case "":
			// RUN FOR BOTH NBA AND WNBA
			fmt.Println("no league argument:", p.Lg)
		case "nba":
			// NBA ONLY
			fmt.Println("nba as league argumenet:", p.Lg)
		case "wnba":
			// WNBA ONLY
			fmt.Println("wnba as league argumenet:", p.Lg)
		}
	default:
		e.Msg = fmt.Sprintf(
			"invalid mode: '%s' is not an option", p.Mode)
		fmt.Println(e.NewErr())
		os.Exit(1)
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
