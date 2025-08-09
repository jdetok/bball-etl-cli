package main

import (
	"flag"
	"fmt"
)

// commment for testing
func main() {
	// args := os.Args
	var word string

	// wordPtr := flag.String("word", "foo", "some string")
	flag.StringVar(&word, "word", "foo", "some string")
	flag.Parse()
	fmt.Println("word arg: ", word)

	// fmt.Println(args)
}
