package main

import (
	"database/sql"
	"flag"
	"fmt"
	"time"

	"github.com/jdetok/bball-etl-cli/etl"
	"github.com/jdetok/golib/errd"
	"github.com/jdetok/golib/logd"
	"github.com/jdetok/golib/pgresd"
)

type Args struct {
	Mode [2]string // run mode e.g. build, daily, etc
	Szn  [2]string // season selector, e.g. 2024 for 2024-25 NBA/2024 WNBA
	Lg   [2]string // league selector, nba or wnba
	Env  [2]string // prod, dev, test
	Logf [2]string // log file, if empty create one
	CC   [2]string // CONCURRENCY
}

// parse flag args
func parseArgs() Args {
	var args = Args{
		Mode: [2]string{"mode", ""},
		Szn:  [2]string{"szn", ""},
		Lg:   [2]string{"lg", ""},
		Env:  [2]string{"env", ""},
		Logf: [2]string{"logf", ""},
		CC:   [2]string{"cc", ""},
	}

	// flag name, default, description
	flag.StringVar(&args.Mode[1], "mode", "", "etl run-mode")
	flag.StringVar(&args.Szn[1], "szn", "", "nba/wnba season e.g. 2024")
	flag.StringVar(&args.Lg[1], "lg", "", "nba or wnba")
	flag.StringVar(&args.Env[1], "env", "dev", "prod or dev postgres database")
	flag.StringVar(&args.Logf[1], "logf", "", "log file, will create if empty")

	var ccFlag bool
	flag.BoolVar(&ccFlag, "cc", false, "enable concurrency")

	flag.Parse()

	if ccFlag {
		args.CC[1] = "true"
	}
	return args
}

// switch environment variables based on the -env flag
func (args *Args) SetEnv(pg *pgresd.PostGres) (*sql.DB, error) {
	// specify which .env file is used
	switch args.Env[1] {
	case "dev":
		*pg = pgresd.GetEnvFilePG("./.envdev")
	case "test":
		*pg = pgresd.GetEnvFilePG("./.envtst")
	case "prod":
		*pg = pgresd.GetEnvPG() // reads .env
	}

	// make postgres connection string from variables in .env
	pg.MakeConnStr()

	// attempt to create pg connection
	db, err := pg.Conn()
	if err != nil {
		return nil, err
	}

	// set limits on open/idle connections
	db.SetMaxOpenConns(40)
	db.SetMaxIdleConns(40)
	return db, nil
}

// logger setup logic for -mode daily
func (args *Args) SetLogger(cnf *etl.Conf, e *errd.Err) (*logd.Logger, error) {
	var l logd.Logger
	var err error
	switch args.Logf[1] { // init logger based on if user passed -logf flag
	case "": // no flag
		// initialize logger and create log file
		l, err = logd.InitLogger("z_log_d", "dly_etl")
		if err != nil {
			return nil, err
		}
	default: // passed a logf
		// initialize logger and create log file | pass dir and file in same string
		l, err = logd.InitLogf(args.Logf[1])
		if err != nil {
			return nil, err
		}
	}
	return &l, nil
}

// logic for -mode daily
func (args *Args) SetupDailyETL(cnf *etl.Conf, e *errd.Err) (string, error) {
	l, lErr := args.SetLogger(cnf, e)
	if lErr != nil {
		e.Msg = "failed to setup logger"
		e.Fatal(lErr)
	}

	// assign logger to cnf
	cnf.L = *l

	// RUN NIGHTLY ETL
	if err := etl.RunNightlyETL(*cnf); err != nil {
		return fmt.Sprintf("error with %v daily etl", etl.Yesterday(time.Now())), err
	}
	return fmt.Sprintf("\n---- daily etl for %v complete | total rows affected: %d",
		etl.Yesterday(time.Now()), cnf.RowCnt), nil
}

// logic for -mode build
// TODO: ADD A CONCURRENT OPTION
func (args *Args) SetupBuildETL(cnf *etl.Conf, st, en string) (string, error) {
	// setup logger, associate with cnf.L
	l, err := logd.InitLogger("z_log_bld", "bld_etl")
	if err != nil {
		return "error initializing logger\n** %v", err
	}
	cnf.L = l

	// current year as string if passed "current"
	if en == "current" {
		en = time.Now().Format("2006")
	}

	if args.CC[1] == "true" {
		// CALL NEW CONCURRENT FUNCTION
		return "concurrent on", nil
	}

	// pass seasons to RunSeasonETL
	if err := etl.RunSeasonETL(*cnf, st, en); err != nil {
		return fmt.Sprintf("error running etl for %s season", args.Szn[1]), err
	}
	return fmt.Sprintf("\n---- etl for %s nba/wnba seasons | total rows affected: %d",
		args.Szn[1], cnf.RowCnt), nil
}

// custom season etl for both leagues
func (args *Args) CustomBothLgETL(cnf *etl.Conf) (string, error) {
	l, err := logd.InitLogger("z_log",
		fmt.Sprintf("szn_etl_%s", args.Szn[1]))
	if err != nil {
		return "error initializing logger", nil

	}
	cnf.L = l // assign to cnf

	// RUN FOR BOTH NBA AND WNBA
	if err := etl.GLogSeasonETL(cnf, args.Szn[1]); err != nil {
		return fmt.Sprintf("error running etl for %s season", args.Szn[1]), nil
	}
	return fmt.Sprintf(
		"\n---- etl for %s nba/wnba seasons | total rows affected: %d",
		args.Szn[1], cnf.RowCnt,
	), nil
}

// custom ETL for specific season
func (args *Args) CustomLgETL(cnf *etl.Conf) (string, error) {
	l, err := logd.InitLogger("z_log",
		fmt.Sprintf("szn_etl_%s_%s", args.Lg[1], args.Szn[1]))
	if err != nil {
		return "error initializing logger", nil

	}
	cnf.L = l // assign to cnf
	// TODO: specific season fetch
	if err := etl.LgSznGlogs(cnf, args.Lg[1], args.Szn[1]); err != nil {
		return fmt.Sprintf("error running etl for %s %s season",
			args.Szn[1], args.Lg[1]), nil

	}
	return fmt.Sprintf(
		"\n---- etl for %s %s seasons | total rows affected: %d",
		args.Szn[1], args.Lg[1], cnf.RowCnt,
	), nil
}
