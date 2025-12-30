package cli

import "flag"

type Params struct {
	EnvFile [2]string
	Mode    [2]string // run mode e.g. build, daily, etc
	Logf    [2]string // log file, if empty create one
	Atch    [2]string // log file, if empty create one
	Szn     [2]string // season selector, e.g. 2024 for 2024-25 NBA/2024 WNBA
	Lg      [2]string // league selector, nba or wnba
}

func ParseArgs() Params {
	var p = Params{
		EnvFile: [2]string{"envf", ""},
		Mode:    [2]string{"mode", ""},
		Szn:     [2]string{"szn", ""},
		Lg:      [2]string{"lg", ""},
		Logf:    [2]string{"logf", ""},
		Atch:    [2]string{"attach", ""},
	}

	// flag name, default, description
	flag.StringVar(&p.EnvFile[1], "envf", "",
		"specify .env file, pass '-envf skip' if environment variables already exist")
	flag.StringVar(&p.Mode[1], "mode", "", "etl run-mode")
	flag.StringVar(&p.Logf[1], "logf", "", "log file, will log to command line if empty")
	flag.StringVar(&p.Atch[1], "attach", "", "attach file (for email mode)")
	flag.StringVar(&p.Szn[1], "szn", "", "nba/wnba season e.g. 2024")
	flag.StringVar(&p.Lg[1], "lg", "", "nba or wnba")

	flag.Parse()
	return p
}
