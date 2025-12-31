package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/jdetok/bball-etl-cli/pkg/cli"
	"github.com/jdetok/bball-etl-cli/pkg/cnf"
	"github.com/jdetok/bball-etl-cli/pkg/conn"
	"github.com/jdetok/bball-etl-cli/pkg/etl"
	"github.com/jdetok/bball-etl-cli/pkg/logd"
	"github.com/jdetok/bball-etl-cli/pkg/maild"
	"github.com/jdetok/bball-etl-cli/pkg/pgdb"
)

const (
	ENV_FILE = ".env"
	PG_OPEN  = 80
	PG_IDLE  = 30
	PG_LIFE  = 30
	PG_HOSTN = "PG_HOST"
	PG_PORTN = "PG_PORT"
	PG_USERN = "PG_USER_FETCH"
	PG_PASSN = "PG_PASS_FETCH"
	PG_DATAN = "PG_DB"
	GM_HOSTN = "GMAIL_HOST"
	GM_PORTN = "GMAIL_PORT"
	GM_USERN = "GMAIL_SNDR"
	GM_PASSN = "GMAIL_PASS"
)

type App struct {
	DBConf    *pgdb.DBConfig
	Cnf       *cnf.Conf
	CmplMsg   string
	StartTime time.Time
}

func main() {
	// parse flags
	var p cli.Params = cli.ParseArgs() // get args passed - exit if 1 will be at least 2 if arg was passed
	runArgs := os.Args
	if len(runArgs) == 1 {
		fmt.Println("fatal: an argument must be passed")
		os.Exit(1)
	}

	app := &App{
		StartTime: time.Now(),
		Cnf:       &cnf.Conf{},
		DBConf:    pgdb.NewDBConf(PG_OPEN, PG_IDLE, PG_LIFE*time.Minute),
	}

	if p.Tst[1] != "" {
		sch, err := etl.GetCurrentLgSchedules()
		if err != nil {
			fmt.Println("failed to get schedules:", err)
			os.Exit(1)
		}
		fmt.Printf("NBA SCHEDULE:\n%v\n\n", sch.N)
		fmt.Printf("WNBA SCHEDULE:\n%v\n\n", sch.W)

		fmt.Println("MAPS:")
		fmt.Println(sch.LgMap["nba"])

		os.Exit(0)
	}

	// if a logf val is passed, set logger as existing file or stdout
	// otherwise, create new file
	var tmpLg io.Writer
	var lgErr error
	logf := p.Logf[1]
	switch logf {
	case "", "cli":
		tmpLg, lgErr = logd.GetLogWriter(true, logf, "")
	default:
		tmpLg, lgErr = logd.GetLogWriter(false, logf, "010206_150405")
	}
	if lgErr != nil {
		fmt.Println("** fatal: failed to get io.Writer for logging")
		os.Exit(1)
	}
	app.Cnf.Lg = logd.NewLogd(tmpLg)

	var pgEnv *conn.DBEnv
	var envErr error
	var envF string = p.EnvFile[1]
	switch envF {
	case "", "skip": // don't load a .env file (env vars already exist)
		pgEnv, envErr = conn.Load(PG_HOSTN, PG_PORTN, PG_USERN, PG_PASSN, PG_DATAN)
	default:
		pgEnv, envErr = conn.LoadFromDotEnv(envF, PG_HOSTN, PG_PORTN, PG_USERN, PG_PASSN, PG_DATAN)
	}

	if envErr != nil {
		app.Cnf.Lg.Fatalf("failed to load environment variables: %v", envErr)
	}

	db, err := pgdb.NewPGConn(pgEnv, app.DBConf)
	if err != nil {
		app.Cnf.Lg.Fatalf("failed to establish postgres connection: %v", err)
	}
	app.Cnf.DB = db
	app.Cnf.RowCnt = 0

	if p.New[1] != "" {
		if err := etl.DailyGamelogs(app.Cnf); err != nil {
			app.Cnf.Lg.Fatalf("new dailygamelogs failed: %v", err)
		}
		os.Exit(0)
	}

	// RUN APPROPRIATE ETL PROCESS BASED ON FLAGS
	runMode := p.Mode[1]
	if runMode == "email" {
		atch := p.Atch[1]
		if atch == "" {
			app.Cnf.Lg.Fatalf("must pass an attachment in email mode")
		}
		if _, err := os.Stat(atch); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				app.Cnf.Lg.Fatalf("fatal: file to attach not found at %s: %v", atch, err)
			}
			app.Cnf.Lg.Fatalf("fatal: error occured finding os.Stat(%s): %v", atch, err)
		}
		if err := maild.EmailLog(atch, GM_USERN, GM_PASSN, GM_HOSTN, GM_PORTN); err != nil {
			app.Cnf.Lg.Fatalf("error emailing log: %v", err)
		}
	}
	switch runMode {
	case "daily", "dly", "d", "": // daily etl: etl for previous day's games
		// RUN NIGHTLY ETL
		yday := etl.Yesterday(time.Now())
		if err = etl.RunNightlyETL(app.Cnf, yday, yday); err != nil {
			app.Cnf.Lg.Fatalf("error with %v daily etl: %v", yday, err)
		}
		app.CmplMsg = fmt.Sprintf( // assign in switch
			"++ daily etl for %v complete \n++ total rows affected: %d\n",
			yday, app.Cnf.RowCnt,
		)
		// build etl: all seasons 1970 through current
	case "build", "bld", "b":
		// SET START AND END SEASONS
		var st string = "1970"                    // make this an arg
		var en string = time.Now().Format("2006") // current year

		// RUN ETL
		if err = etl.RunSeasonETL(app.Cnf, st, en); err != nil {
			app.Cnf.Lg.Fatalf("error running season etl: start year: %s | end year: %s", st, en)
		}
		app.Cnf.Lg.Infof("++ etl for seasons between %s and %s | total rows affected: %d\n",
			st, en, app.Cnf.RowCnt)

		// "custom" run - a season MUST be specified, lg defaults to both
	case "custom":
		// exit if no season passed
		szn := p.Szn[1]
		lg := p.Lg[1]
		df := p.DateFrom[1]
		dt := p.DateTo[1]

		if szn == "" && df == "" && dt == "" {
			app.Cnf.Lg.Fatalf("a season or date must be specified in custom mode")
		}

		// custom dates
		if df != "" || dt != "" {
			if dt == "" && df != "" {
				dt = df
			} else if df == "" && dt != "" {
				df = dt
			}
			if err = etl.RunNightlyETL(app.Cnf, df, dt); err != nil {
				app.Cnf.Lg.Fatalf("error with %s-%s daily etl: %v", df, dt, err)
			}
			app.CmplMsg = fmt.Sprintf( // assign in switch
				"++ daily etl for %s-%s complete \n++ total rows affected: %d\n",
				df, dt, app.Cnf.RowCnt,
			)
			break
		}

		// switch on lg to determine whether to do both leagues or just one
		switch lg {
		case "":
			// RUN FOR BOTH NBA AND WNBA
			if err := etl.GLogSeasonETL(app.Cnf, szn); err != nil {
				app.Cnf.Lg.Fatalf("error running etl for %s season", szn)
			}
			app.Cnf.Lg.Infof("++ COMPLETED {SZN: %s} nba/wnba seasons | total rows affected: %d\n",
				szn, app.Cnf.RowCnt,
			)
		case "nba", "wnba":
			// TODO: specific season fetch
			if err := etl.LgSznGlogs(app.Cnf, lg, szn); err != nil {
				app.Cnf.Lg.Fatalf("error running etl for %s %s season",
					szn, lg)
			}
		}
		app.Cnf.Lg.Infof("COMPLETED {SZN: %s} {LG: %s} ETL\n++ total rows affected: %d\n",
			szn, lg, app.Cnf.RowCnt,
		)

	// NO ARGS PASSED - ERROR OUT
	default:
		app.Cnf.Lg.Fatalf("invalid mode: '%s' is not an option", runMode)
	}

	// write errors to the log
	if len(app.Cnf.Errs) > 0 {
		for _, e := range app.Cnf.Errs {
			app.Cnf.Lg.Errorf(e)
		}
	}

	// complete log
	app.Cnf.Lg.Infof("ETL COMPLETE\n++ STARTED AT: %v\n++ COMPLETED AT: %v\n++ RUNTIME: %v\n%s",
		app.StartTime, time.Now(), time.Since(app.StartTime), app.CmplMsg)
}
