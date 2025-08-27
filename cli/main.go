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
	// "flag"
	"fmt"
	"os"
	"time"

	"github.com/jdetok/bball-etl-cli/etl"
	"github.com/jdetok/golib/errd"
	"github.com/jdetok/golib/logd"
	"github.com/jdetok/golib/maild"
	"github.com/jdetok/golib/pgresd"
)

func EmailLog(l logd.Logger) error {
	m := maild.MakeMail(
		[]string{"jdekock17@gmail.com"},
		"Go bball ETL log attached",
		"the Go bball ETL process ran. The log is attached.",
	)
	l.WriteLog(fmt.Sprintf("attempting to email %s to %s", l.LogF, m.MlTo[0]))
	return m.SendMIMEEmail(l.LogF)
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

	// parse flags
	var p Params = parseArgs()

	// init database based on -dev flag
	var pg pgresd.PostGres
	switch p.Env[1] {
	case "dev":
		pg = pgresd.GetEnvFilePG("./.envdev")
	case "test":
		pg = pgresd.GetEnvFilePG("./.envtst")
	case "prod":
		pg = pgresd.GetEnvPG() // reads .env
	}
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

	fmt.Println("DEBUG db connected")
	// RUN APPROPRIATE ETL PROCESS BASED ON FLAGS
	switch p.Mode[1] {
	case "": // no mode passed,
		e.Msg = "a mode must be specified"
		fmt.Println(e.NewErr())
		os.Exit(1)

	// daily etl: etl for previous day's games
	case "daily": 
		switch p.Logf[1] { // init logger based on if user passed -logf flag
		case "": // no flag
			// initialize logger and create log file
			l, err := logd.InitLogger("z_log_d", "dly_etl")
			if err != nil {
				e.Msg = "error initializing logger"
				fmt.Println(e.BuildErr(err))
				os.Exit(1)
			}
			cnf.L = l // assign to cnf
		default: // passed a logf
			// initialize logger and create log file
			// pass dir and file in same string
			l, err := logd.InitLogf(p.Logf[1])
			if err != nil {
				e.Msg = "error initializing logger"
				fmt.Println(e.BuildErr(err))
				os.Exit(1)
			}
			cnf.L = l // assign to cnf
		}
	
		// RUN NIGHTLY ETL
		if err = etl.RunNightlyETL(cnf); err != nil {
			e.Msg = fmt.Sprintf(
				"error with %v daily etl", etl.Yesterday(time.Now()))
			cnf.L.WriteLog(e.Msg)
			fmt.Println(e.BuildErr(err))
			os.Exit(1)
		}
		compMsg = fmt.Sprintf( // assign in switch
			"\n---- daily etl for %v complete | total rows affected: %d",
			etl.Yesterday(time.Now()), cnf.RowCnt,
		)

		// build etl: all seasons 1970 through current
	case "build":
		l, err := logd.InitLogger("z_log_bld", "bld_etl")
		if err != nil {
			e.Msg = "error initializing logger"
			fmt.Println(e.BuildErr(err))
			os.Exit(1)
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
			fmt.Println(e.BuildErr(err))
			os.Exit(1)
		}
		compMsg = fmt.Sprintf(
			"\n---- etl for seasons between %s and %s | total rows affected: %d",
			st, en, cnf.RowCnt,
		)

		// "custom" run - a season MUST be specified, lg defaults to both
	case "custom":
		// exit if no season passed
		if p.Szn[1] == "" {
			e.Msg = "a season (-szn) must be specified in custom mode"
			fmt.Println(e.NewErr())
			os.Exit(1)
		}
		// switch on lg to determine whether to do both leagues or just one
		switch p.Lg[1] {
		case "":
			l, err := logd.InitLogger("z_log",
				fmt.Sprintf("szn_etl_%s", p.Szn[1]))
			if err != nil {
				e.Msg = "error initializing logger"
				fmt.Println(e.BuildErr(err))
				os.Exit(1)
			}
			cnf.L = l // assign to cnf

			// RUN FOR BOTH NBA AND WNBA
			if err := etl.GLogSeasonETL(&cnf, p.Szn[1]); err != nil {
				e.Msg = fmt.Sprintf("error running etl for %s season", p.Szn[1])
				fmt.Println(e.BuildErr(err))
				os.Exit(1)
			}
			compMsg = fmt.Sprintf(
				"\n---- etl for %s nba/wnba seasons | total rows affected: %d",
				p.Szn[1], cnf.RowCnt,
			)
		case "nba", "wnba":
			l, err := logd.InitLogger("z_log",
				fmt.Sprintf("szn_etl_%s_%s", p.Lg[1], p.Szn[1]))
			if err != nil {
				e.Msg = "error initializing logger"
				fmt.Println(e.BuildErr(err))
				os.Exit(1)
			}
			cnf.L = l // assign to cnf
			// TODO: specific season fetch
			if err := etl.LgSznGlogs(&cnf, p.Lg[1], p.Szn[1]); err != nil {
				e.Msg = fmt.Sprintf("error running etl for %s %s season",
					p.Szn[1], p.Lg[1])
				fmt.Println(e.BuildErr(err))
				os.Exit(1)
			}
			compMsg = fmt.Sprintf(
				"\n---- etl for %s %s seasons | total rows affected: %d",
				p.Szn[1], p.Lg[1], cnf.RowCnt,
			)
		}

		// NO ARGS PASSED - ERROR OUT
	default:
		e.Msg = fmt.Sprintf(
			"invalid mode: '%s' is not an option", p.Mode[1])
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

	// email log file to myself
	EmailLog(cnf.L)
	if err != nil {
		e.Msg = "error emailing log"
		cnf.L.WriteLog(e.Msg)
		fmt.Println(e.BuildErr(err))
		os.Exit(1)
	}

	cnf.L.WriteLog("email sent - exiting bball-etl-cli")
}
