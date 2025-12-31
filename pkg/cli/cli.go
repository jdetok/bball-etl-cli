package cli

import "flag"

type CLIArgs struct {
	EnvFile  string
	Mode     string // run mode e.g. build, daily, etc
	Logf     string // log file, if empty create one
	Atch     string // log file, if empty create one
	Szn      string // season selector, e.g. 2024 for 2024-25 NBA/2024 WNBA
	Lg       string // league selector, nba or wnba
	DateFrom string
	DateTo   string
	Tst      string
	New      string
}

func ParseArgs() *CLIArgs {
	p := &CLIArgs{}
	// flag name, default, description
	flag.StringVar(&p.EnvFile, "envf", "",
		"specify .env file, pass '-envf skip' if environment variables already exist")
	flag.StringVar(&p.Mode, "mode", "", "etl run-mode")
	flag.StringVar(&p.Logf, "logf", "", "log file, will log to command line if empty")
	flag.StringVar(&p.Atch, "attach", "", "attach file (for email mode)")
	flag.StringVar(&p.Szn, "szn", "", "nba/wnba season e.g. 2024")
	flag.StringVar(&p.Lg, "lg", "", "nba or wnba")
	flag.StringVar(&p.DateFrom, "datefrom", "", "date specfic fetch, used with -mode custom")
	flag.StringVar(&p.DateTo, "dateto", "", "date specfic fetch, used with -mode custom")
	flag.StringVar(&p.Tst, "tst", "", "for dev/testing")
	flag.StringVar(&p.New, "new", "", "test new feature")
	flag.Parse()
	return p
}
