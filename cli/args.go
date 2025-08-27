package main

import "flag"
// TODO: add a -logf flag, if it's empty create file
// - this enables creating file in .sh, passing it and bypassing the InitLogger

type Params struct {
	Mode [2]string // run mode e.g. build, daily, etc
	Szn  [2]string // season selector, e.g. 2024 for 2024-25 NBA/2024 WNBA
	Lg   [2]string // league selector, nba or wnba
	Env  [2]string // prod, dev, test
	Logf [2]string // log file, if empty create one
}

func parseArgs() Params {
	var p = Params{
		Mode: [2]string{"mode", ""},
		Szn:  [2]string{"szn", ""},
		Lg:   [2]string{"lg", ""},
		Env:  [2]string{"env", ""},
		Logf: [2]string{"logf", ""},
	}

	// flag name, default, description
	flag.StringVar(&p.Mode[1], "mode", "", "etl run-mode")
	flag.StringVar(&p.Szn[1], "szn", "", "nba/wnba season e.g. 2024")
	flag.StringVar(&p.Lg[1], "lg", "", "nba or wnba")
	flag.StringVar(&p.Env[1], "env", "dev", "prod or dev postgres database")
	flag.StringVar(&p.Logf[1], "logf", "", "log file, will create if empty")
	flag.Parse()
	return p
}