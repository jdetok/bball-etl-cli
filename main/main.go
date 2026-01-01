package main

import (
	"fmt"
	"os"
	"time"

	"github.com/jdetok/bball-etl-cli/pkg/cli"
	"github.com/jdetok/bball-etl-cli/pkg/cnf"
	"github.com/jdetok/bball-etl-cli/pkg/logd"
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

func main() {
	meta := &cli.RunMode{
		StartTime:    time.Now(),
		Cnf:          &cnf.Conf{},
		Args:         cli.ParseArgs(),
		EmailUserVar: GM_USERN,
		EmailPassVar: GM_PASSN,
		EmailHostVar: GM_HOSTN,
		EmailPortVar: GM_PORTN,
	}

	// SETUP LOGGING
	tmpLg, err := getLogWriter(meta.Args.Logf.Value)
	if err != nil {
		fmt.Println(fmt.Errorf("** fatal: failed to get io.Writer for logging: %v", err))
		os.Exit(1)
	}
	meta.Cnf.Lg = logd.NewLogd(tmpLg)

	meta.Cnf.Lg.Infof("logger configured successfully")

	// load environment variables for database connection
	pgEnv, err := loadDBEnv(meta.Args.Envf.Value)
	if err != nil {
		meta.Cnf.Lg.Fatalf("failed to load env vars for DB connection: %v", err)
	}

	// setup database connection
	db, err := pgdb.NewPGConn(pgEnv, pgdb.NewDBConf(PG_OPEN, PG_IDLE, PG_LIFE*time.Minute))
	if err != nil {
		meta.Cnf.Lg.Fatalf("failed to establish postgres connection: %v", err)
	}
	meta.Cnf.DB = db
	meta.Cnf.RowCnt = 0

	meta.Cnf.Lg.Infof("database connected successfully")

	// build run modes map
	modes := cli.BuildRunModes(meta)

	meta.Cnf.Lg.Infof("run mode: %s", meta.Args.Mode.Value)

	// retrieve appropriate etl function from runmode map
	fn, ok := modes[*meta.Args.Mode.Dest]
	if !ok {
		meta.Cnf.Lg.Fatalf("invalid mode: '%s' does not exist", meta.Args.Mode.Name)
	}

	// execute etl function from runmode map
	if err := fn(meta); err != nil {
		meta.Cnf.Lg.Fatalf("%s mode failed: %v", meta.Args.Mode.Value, err)
	}

	// complete log
	meta.Cnf.Lg.Infof(
		"ETL COMPLETE: %d TOTAL ROWS AFFECTED\n++ STARTED AT: %v\n++ COMPLETED AT: %v\n++ RUNTIME: %v\n",
		meta.Cnf.RowCnt, meta.StartTime, time.Now(), time.Since(meta.StartTime))
}
