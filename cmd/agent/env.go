package main

import (
	"net"
	"os"
	"strconv"
)

const (
	EnvAddressVarName        = "ADDRESS"
	EnvReportIntervalVarName = "REPORT_INTERVAL"
	EnvPollIntervalVarName   = "POLL_INTERVAL"
)

type EnvVars struct {
	Host              string // "" если не задано
	Port              *int   // nil если не задано
	ReportIntervalSec *int   // nil если не задано
	PollIntervalSec   *int   // nil если не задано
}

func getEnvVars() EnvVars {
	var e EnvVars

	if addr := os.Getenv(EnvAddressVarName); addr != "" {
		if h, p, ok := splitHostPort(addr); ok {
			e.Host, e.Port = h, &p
		}
	}

	if v := os.Getenv(EnvReportIntervalVarName); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			e.ReportIntervalSec = &n
		}
	}
	if v := os.Getenv(EnvPollIntervalVarName); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			e.PollIntervalSec = &n
		}
	}
	return e
}

func splitHostPort(addr string) (string, int, bool) {
	h, ps, err := net.SplitHostPort(addr)
	if err != nil || h == "" || ps == "" {
		return "", 0, false
	}
	p, err := strconv.Atoi(ps)
	if err != nil || p <= 0 {
		return "", 0, false
	}
	return h, p, true
}
