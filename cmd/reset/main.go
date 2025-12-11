package main

import (
	"flag"
	"log"
	"os"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/resetgen"
)

var run = resetgen.Run
var fatalf = log.Fatalf

func main() {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	dirFlag := fs.String("dir", ".", "root directory to scan")
	fs.Parse(os.Args[1:])

	if err := run(*dirFlag); err != nil {
		fatalf("reset generation failed: %v", err)
	}
}
