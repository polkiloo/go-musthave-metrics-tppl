package main

import (
	"flag"
	"fmt"
	"io"
)

func parseFlags(args []string) (string, error) {
	fs := flag.NewFlagSet("server", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	addr := "localhost:8080"
	fs.StringVar(&addr, "a", addr, "HTTP server address (host:port)")

	if err := fs.Parse(args); err != nil {
		return "", err
	}
	if fs.NArg() > 0 {
		return "", fmt.Errorf("unknown arguments: %v", fs.Args())
	}
	return addr, nil
}
