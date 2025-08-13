package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/agent"
)

var (
	ErrUnknownAddress           = fmt.Errorf("unknown address")
	ErrParseAddressFlags        = fmt.Errorf("error parsing address")
	ErrParseReportIntervalFlags = fmt.Errorf("error parsing report interval")
	ErrParsePollIntervalFlags   = fmt.Errorf("error parsing poll interval flags")
)

type Args struct {
	Host           string
	Port           int
	ReportInterval time.Duration
	PollInterval   time.Duration
}

const (
	defaultHost           = "localhost"
	defaultPort           = 8080
	defaultReportInterval = 10 * time.Second
	defaultPollInterval   = 2 * time.Second
)

var (
	defaultAddress = defaultHost + ":" + strconv.Itoa(defaultPort)
)

func parseFlags() (Args, error) {
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
		return Args{}, ErrUnknownAddress
	}

	args := Args{
		Host:           defaultHost,
		Port:           defaultPort,
		ReportInterval: defaultReportInterval,
		PollInterval:   defaultPollInterval,
	}

	if *reportFlag != "" {
		sec, err := strconv.Atoi(*reportFlag)
		if err != nil || sec <= 0 {
			fmt.Fprintf(os.Stderr, "incorrect value for -r (reportInterval): %v\n", *reportFlag)
			return Args{}, ErrParseReportIntervalFlags
		}
		args.ReportInterval = time.Duration(sec) * time.Second
	}
	if *pollFlag != "" {
		sec, err := strconv.Atoi(*pollFlag)
		if err != nil || sec <= 0 {
			fmt.Fprintf(os.Stderr, "incorrect value for -p (pollInterval): %v\n", *pollFlag)
			return Args{}, ErrParsePollIntervalFlags
		}
		args.PollInterval = time.Duration(sec) * time.Second
	}

	if *addressFlag != defaultAddress {

		addr := *addressFlag
		host, portStr, err := net.SplitHostPort(addr)
		if err != nil || host == "" || portStr == "" {
			return Args{}, ErrParseAddressFlags
		}
		port, err := strconv.Atoi(portStr)
		if err != nil || port <= 0 {
			return Args{}, ErrParseAddressFlags
		}
		args.Host = host
		args.Port = port
	}

	return args, nil
}

func (a Args) ToAppConfig() agent.AppConfig {
	return agent.AppConfig{
		Host:           a.Host,
		Port:           a.Port,
		ReportInterval: a.ReportInterval,
		PollInterval:   a.PollInterval,
	}
}
