package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/jdetok/golib/errd"
)

/*
PROJECT INTENT:
  - create a command line interface for the bball-etl-go package to eliminate
    the need for separate directories for the nightly/build runs of the etl
*/

/* TODO:
- exit program if no arguments are passed (args slices == 1)
- define a flag for mode:
	- build: runs full etl all games 1970 through current
	- daily: runs etl for games from previous day

- LATER:
	- define an additional flag for league
	- eventually, define flags for different endpoint endpoints
*/

type Params struct {
	Mode [2]string
	Szn  [2]string
	Lg   [2]string
}

func parseArgs() Params {
	var p = Params{
		Mode: [2]string{"mode", ""},
		Szn:  [2]string{"szn", ""},
		Lg:   [2]string{"lg", ""},
	}
	flag.StringVar(&p.Mode[1], "mode", "", "etl run-mode")
	flag.StringVar(&p.Szn[1], "szn", "", "nba/wnba season e.g. 2024")
	flag.StringVar(&p.Lg[1], "lg", "", "nba or wnba")
	flag.Parse()
	return p
}

func main() {
	e := errd.InitErr()
	runArgs := os.Args

	// one argument by default (name of program) - exit if nothing passed
	if len(runArgs) == 1 {
		e.Msg = "an argument must be passed"
		fmt.Println(e.NewErr())
		os.Exit(1)
	}

	var p Params = parseArgs()
	fmt.Println(p.Mode)
	fmt.Println(p.Szn)
	fmt.Println(p.Lg)
}
