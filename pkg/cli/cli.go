package cli

import (
	"flag"
	"time"

	"github.com/jdetok/bball-etl-cli/pkg/etl"
)

type Arg struct {
	Name        string
	Desc        string
	Default     string
	DefaultBool bool
	Value       string
	ValueBool   bool
	Dest        *string
	DestBool    *bool
}

type CLIArgs struct {
	Envf     Arg
	Mode     Arg // run mode e.g. build, daily, etc
	Logf     Arg // log file, if empty create one
	Atch     Arg // log file, if empty create one
	Szn      Arg // season selector, e.g. 2024 for 2024-25 NBA/2024 WNBA
	Lg       Arg // league selector, nba or wnba
	DateFrom Arg
	DateTo   Arg
	StartYr  Arg
	EndYr    Arg
	Tst      Arg
	New      Arg

	ArgVals  []*Arg
	ArgFlags []*Arg
}

func ParseArgs() *CLIArgs {
	p := &CLIArgs{
		Envf: Arg{
			Name:    "envf",
			Desc:    "specify .env file, pass '-envf skip' if environment variables already exist",
			Default: ".env",
		},
		Mode: Arg{
			Name:    "mode",
			Desc:    "etl run mode",
			Default: "daily",
		},
		Logf: Arg{
			Name:    "logf",
			Desc:    "log file, will log to command line if empty",
			Default: "cli",
		},
		Atch: Arg{
			Name: "attach",
			Desc: "attach file (for email mode)",
		},
		Szn: Arg{
			Name: "szn",
			Desc: "nba/wnba season e.g. 2024",
		},
		Lg: Arg{
			Name: "lg",
			Desc: "specify nba or wnba (both by default)",
		},
		DateFrom: Arg{
			Name: "dateFrom",
			Desc: "start date for custom daily mode",
		},
		DateTo: Arg{
			Name: "dateTo",
			Desc: "end date for custom daily mode",
		},
		StartYr: Arg{
			Name:    "startYear",
			Desc:    "attach file (for email mode)",
			Default: "1970",
		},
		EndYr: Arg{
			Name:    "endYear",
			Desc:    "attach file (for email mode)",
			Default: etl.CurrentSzns(time.Now())[0][:4],
			// Default: time.Now().Format("2006"),
		},
		Tst: Arg{
			Name:        "tst",
			Desc:        "for dev use - only run tst func",
			DefaultBool: false,
		},
		New: Arg{
			Name:        "new",
			Desc:        "for dev use - only run new func",
			DefaultBool: false,
		},
	}

	p.ArgVals = []*Arg{
		&p.Envf,
		&p.Mode,
		&p.Logf,
		&p.Atch,
		&p.Szn,
		&p.Lg,
		&p.DateFrom,
		&p.DateTo,
		&p.StartYr,
		&p.EndYr,
	}

	p.ArgFlags = []*Arg{
		&p.Tst,
		&p.New,
	}

	for i := range p.ArgVals {
		arg := p.ArgVals[i]
		arg.Dest = &arg.Value
		flag.StringVar(arg.Dest, arg.Name, arg.Default, arg.Desc)
	}

	for i := range p.ArgFlags {
		arg := p.ArgFlags[i]
		arg.DestBool = &arg.ValueBool
		flag.BoolVar(arg.DestBool, arg.Name, arg.DefaultBool, arg.Desc)
	}

	flag.Parse()
	return p
}
