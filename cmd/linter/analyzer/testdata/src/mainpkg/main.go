package mainpkg

import (
	"log"
	"os"
)

func main() {
	log.Fatal("blocked") // want "log.Fatal should be called only from main function in package main"
	os.Exit(0)           // want "os.Exit should be called only from main function in package main"
}

func helper() {
	panic("not allowed") // want "avoid using panic"
}

func other() {
	log.Fatal("stop") // want "log.Fatal should be called only from main function in package main"
}

func another() {
	os.Exit(1) // want "os.Exit should be called only from main function in package main"
}
