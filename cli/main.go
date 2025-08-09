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
	Mode [2]string // run mode e.g. build, daily, etc
	Szn  [2]string // season selector, e.g. 2024 for 2024-25 NBA/2024 WNBA
	Lg   [2]string // league selector, nba or wnba
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

	// DEFINE BEHAVIOR BASED ON MODE ARGUMENT
	switch p.Mode[1] {
	case "": // no mode passed,
		e.Msg = "a mode must be specified"
		fmt.Println(e.NewErr())
		os.Exit(1)
	case "build":
		// build etl: all seasons 1970 through current
		fmt.Println("build mode selected:", p.Mode)
	case "daily":
		// daily etl: etl for previous day's games
		// might be worth switching on league from here, then can handle which
		// league to run in the bash script rather than in the application
		fmt.Println("daily mode selected:", p.Mode)
	case "custom": // "custom" run - a season MUST be specified, lg defaults to both
		fmt.Println("custom mode selected:", p.Mode)
		if p.Szn[1] == "" {
			e.Msg = "a season (-szn) must be specified in custom mode"
			fmt.Println(e.NewErr())
			os.Exit(1)
		}
		switch p.Lg[1] {
		case "":
			// RUN FOR BOTH NBA AND WNBA
			fmt.Println("no league argument:", p.Lg)
		case "nba":
			// NBA ONLY
			fmt.Println("nba as league argumenet:", p.Lg)
		case "wnba":
			// WNBA ONLY
			fmt.Println("wnba as league argumenet:", p.Lg)
		}
	default:
		e.Msg = fmt.Sprintf(
			"invalid mode: '%s' is not an option", p.Mode)
		fmt.Println(e.NewErr())
		os.Exit(1)
	}
}
