package agentcfg

import (
	"os"
	"strconv"

	commoncfg "github.com/polkiloo/go-musthave-metrics-tppl/internal/config/common"
)

const (
	EnvAddressVarName        = "ADDRESS"
	EnvReportIntervalVarName = "REPORT_INTERVAL"
	EnvPollIntervalVarName   = "POLL_INTERVAL"
)

type AgentEnvVars struct {
	Host              string
	Port              *int
	ReportIntervalSec *int
	PollIntervalSec   *int
}

func getEnvVars() (AgentEnvVars, error) {
	var e AgentEnvVars

	hp, _ := commoncfg.ReadHostPortEnv(EnvAddressVarName)

	e.Host = hp.Host
	e.Port = hp.Port

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
	return e, nil
}
