package main

import (
	"flag"
	"fmt"
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

func main() {
	// args := os.Args
	var word string

	// wordPtr := flag.String("word", "foo", "some string")
	flag.StringVar(&word, "word", "foo", "some string")
	flag.Parse()
	fmt.Println("word arg: ", word)

	// fmt.Println(args)
}
