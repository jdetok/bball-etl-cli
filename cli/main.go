package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/jdetok/bball-etl-cli/etl"
	"github.com/jdetok/bball-etl-cli/pkg/conn"
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
	DBConf *pgdb.DBConfig
	// EmailConf EmailCnf
	Cnf       *etl.Conf
	CmplMsg   string
	StartTime time.Time
}

// func EmailLog(logf string) error {
// 	m := maild.MakeMail(
// 		[]string{"jdekock17@gmail.com"},
// 		"Go bball ETL log attached",
// 		"the Go bball ETL process ran. The log is attached.",
// 	)
// 	return m.SendMIMEEmail(logf)
// }

func main() {
	// parse flags
	var p Params = parseArgs() // get args passed - exit if 1 will be at least 2 if arg was passed
	runArgs := os.Args
	if len(runArgs) == 1 {
		fmt.Println("fatal: an argument must be passed")
		os.Exit(1)
	}

	app := &App{
		StartTime: time.Now(),
		Cnf:       &etl.Conf{},
		DBConf:    pgdb.NewDBConf(PG_OPEN, PG_IDLE, PG_LIFE*time.Minute),
	}

	// if a logf val is passed, set logger as existing file or stdout
	// otherwise, create new file
	var tmpLg io.Writer
	var lgErr error
	logf := p.Logf[1]
	if logf == "" {
		tmpLg, lgErr = logd.GetLogWriter(true, logf, "")
	} else {
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
	if envF == "" || envF == "skip" {
		pgEnv, envErr = conn.Load(PG_HOSTN, PG_PORTN, PG_USERN, PG_PASSN, PG_DATAN)
	} else {
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

	// RUN APPROPRIATE ETL PROCESS BASED ON FLAGS
	runMode := p.Mode[1]
	switch runMode {
	case "email": // send log file in an email
		if logf == "" {
			app.Cnf.Lg.Fatalf("must pass a log file when run in email mode")
		}
		if _, err := os.Stat(logf); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				app.Cnf.Lg.Fatalf("fatal: file to attach not found at %s: %v", logf, err)
			}
			app.Cnf.Lg.Fatalf("fatal: error occured finding os.Stat(%s): %v", logf, err)
		}
		if err := maild.EmailLog(logf, GM_USERN, GM_PASSN, GM_HOSTN, GM_PORTN); err != nil {
			app.Cnf.Lg.Fatalf("error emailing log: %v", err)
		}

	// daily etl: etl for previous day's games
	case "daily", "dly", "d", "":
		// RUN NIGHTLY ETL
		if err = etl.RunNightlyETL(app.Cnf); err != nil {
			app.Cnf.Lg.Fatalf("error with %v daily etl", etl.Yesterday(time.Now()))
		}
		app.CmplMsg = fmt.Sprintf( // assign in switch
			"++ daily etl for %v complete | total rows affected: %d\n",
			etl.Yesterday(time.Now()), app.Cnf.RowCnt,
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
		app.CmplMsg = fmt.Sprintf(
			"++ etl for seasons between %s and %s | total rows affected: %d\n",
			st, en, app.Cnf.RowCnt,
		)

		// "custom" run - a season MUST be specified, lg defaults to both
	case "custom":
		// exit if no season passed
		if p.Szn[1] == "" {
			app.Cnf.Lg.Fatalf("a season (-szn) must be specified in custom mode")
		}
		// switch on lg to determine whether to do both leagues or just one
		switch p.Lg[1] {
		case "":
			// RUN FOR BOTH NBA AND WNBA
			if err := etl.GLogSeasonETL(app.Cnf, p.Szn[1]); err != nil {
				app.Cnf.Lg.Fatalf("error running etl for %s season", p.Szn[1])
			}
			app.CmplMsg = fmt.Sprintf(
				"++ etl for %s nba/wnba seasons | total rows affected: %d\n",
				p.Szn[1], app.Cnf.RowCnt,
			)
		case "nba", "wnba":
			// TODO: specific season fetch
			if err := etl.LgSznGlogs(app.Cnf, p.Lg[1], p.Szn[1]); err != nil {
				app.Cnf.Lg.Fatalf("error running etl for %s %s season",
					p.Szn[1], p.Lg[1])
			}
			app.CmplMsg = fmt.Sprintf(
				"++ etl for %s %s seasons | total rows affected: %d\n",
				p.Szn[1], p.Lg[1], app.Cnf.RowCnt,
			)
		}
		// EMAIL MODE: RUN AT END OF SH

	// NO ARGS PASSED - ERROR OUT
	default:
		app.Cnf.Lg.Fatalf("invalid mode: '%s' is not an option", p.Mode[1])
	}

	// write errors to the log
	if len(app.Cnf.Errs) > 0 {
		for _, e := range app.Cnf.Errs {
			app.Cnf.Lg.Errorf(e)
		}
	}

	// complete log
	app.Cnf.Lg.Infof("etl complete\n++ start time: %v\n++ complete time: %v\n++ duration: %v\n%s",
		app.StartTime, time.Now(), time.Since(app.StartTime), app.CmplMsg)
}
