package main

import (
	"log"
	"os"
)

func main() {
	log.Fatal("allowed")
	os.Exit(0)
}

func nested() {
	f := func() {
		log.Fatal("disallowed") // want "log.Fatal should be called only from main function in package main"
	}
	f()
}
