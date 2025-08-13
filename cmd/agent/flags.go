package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/agent"
)

var (
	ErrUnknownArgs = fmt.Errorf("unknown args")
)

type FlagsArg struct {
	Host              string // "" если не задано
	Port              *int   // nil если не задано
	ReportIntervalSec *int   // nil если не задано
	PollIntervalSec   *int   // nil если не задано
}

var (
	defaultAddress = agent.DefaultAppHost + ":" + strconv.Itoa(agent.DefaultAppPort)
)

func parseFlags() (FlagsArg, error) {
	addressFlag := flag.String("a", defaultAddress, "HTTP endpoint (e.g, localhost:8080)")
	reportFlag := flag.String("r", "", "reportInterval in seconds (default 10 seconds)")
	pollFlag := flag.String("p", "", "pollInterval in seconds (default 2 seconds)")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [-a address] [-r reportIntervalSeconds] [-p pollIntervalSeconds]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if len(flag.Args()) > 0 {
		fmt.Fprintf(os.Stderr, "unknown args: %v\n", flag.Args())
		return FlagsArg{}, ErrUnknownArgs
	}

	args := FlagsArg{}

	if *reportFlag != "" {
		sec, err := strconv.Atoi(*reportFlag)
		if err != nil || sec <= 0 {
			fmt.Fprintf(os.Stderr, "incorrect value for -r (reportInterval): %v\n", *reportFlag)
		}
		args.ReportIntervalSec = &sec
	}
	if *pollFlag != "" {
		sec, err := strconv.Atoi(*pollFlag)
		if err != nil || sec <= 0 {
			fmt.Fprintf(os.Stderr, "incorrect value for -p (pollInterval): %v\n", *pollFlag)
		}
		args.PollIntervalSec = &sec
	}

	if *addressFlag != defaultAddress {

		addr := *addressFlag
		host, portStr, err := net.SplitHostPort(addr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "incorrect value for -a (HTTP addres): %v\n", *addressFlag)
		}

		port, err := strconv.Atoi(portStr)
		if err == nil && port > 0 {
			args.Port = &port
		}

		args.Host = host

	}

	return args, nil
}
