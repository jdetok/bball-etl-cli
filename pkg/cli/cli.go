package cli

import (
	"flag"
	"time"
)

type CLIArgs struct {
	EnvFile  string
	Mode     string // run mode e.g. build, daily, etc
	Logf     string // log file, if empty create one
	Atch     string // log file, if empty create one
	Szn      string // season selector, e.g. 2024 for 2024-25 NBA/2024 WNBA
	Lg       string // league selector, nba or wnba
	DateFrom string
	DateTo   string
	StartYr  string
	EndYr    string
	Tst      bool
	New      bool
}

func ParseArgs() *CLIArgs {
	p := &CLIArgs{}

	// env flag - determines whether an env file is read before loading env vars
	flag.StringVar(&p.EnvFile, "envf", ".env",
		"specify .env file, pass '-envf skip' if environment variables already exist")

	// main run modes flag
	flag.StringVar(&p.Mode, "mode", "daily", "etl run-mode")

	flag.StringVar(&p.Logf, "logf", "cli", "log file, will log to command line if empty")
	flag.StringVar(&p.Atch, "attach", "", "attach file (for email mode)")

	// seasons for custom mode
	flag.StringVar(&p.Szn, "szn", "", "nba/wnba season e.g. 2024")
	flag.StringVar(&p.Lg, "lg", "", "nba or wnba")

	// from/to dates for custom daily etl mode
	flag.StringVar(&p.DateFrom, "datefrom", "", "date specfic fetch, used with -mode custom")
	flag.StringVar(&p.DateTo, "dateto", "", "date specfic fetch, used with -mode custom")

	// build mode start/end years
	flag.StringVar(&p.StartYr, "startYear", "1970", "build mode start year")
	flag.StringVar(&p.EndYr, "endYear", time.Now().Format("2006"), "build mode end year")

	// tst/new flags - for development
	flag.BoolVar(&p.Tst, "tst", false, "for dev/testing")
	flag.BoolVar(&p.New, "new", false, "test new feature")

	flag.Parse()
	return p
}
