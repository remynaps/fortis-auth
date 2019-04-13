package main

import (
	"flag"
	"os"
)

func main() {
	textPtr := flag.String("create_client", "", "Create a token")
	lsPtr := flag.String("ls", "", "List current tokens")

	flag.Parse()

	if *textPtr == "" {
		flag.PrintDefaults()
		os.Exit(1)
	} else {

	}
	if *lsPtr == "" {
		flag.PrintDefaults()
		os.Exit(1)
	} else {

	}
}
