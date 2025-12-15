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
	EnvKeyVarName            = "KEY"
	EnvRateLimitVarName      = "RATE_LIMIT"
	EnvCryptoKeyPathVarName  = "CRYPTO_KEY"
)

type AgentEnvVars struct {
	Host              string
	Port              *int
	ReportIntervalSec *int
	PollIntervalSec   *int
	SignKey           *string
	RateLimit         *int
	CryptoKeyPath     *string
}

func getEnvVars() (AgentEnvVars, error) {
	var e AgentEnvVars

	hp, _ := commoncfg.ReadHostPortEnv(EnvAddressVarName)

	e.Host = hp.Host
	e.Port = hp.Port

	if v, ok := os.LookupEnv(EnvReportIntervalVarName); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			e.ReportIntervalSec = &n
		}
	}
	if v, ok := os.LookupEnv(EnvPollIntervalVarName); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			e.PollIntervalSec = &n
		}
	}
	if v, ok := os.LookupEnv(EnvKeyVarName); ok {
		e.SignKey = &v
	}
	if v, ok := os.LookupEnv(EnvCryptoKeyPathVarName); ok && v != "" {
		e.CryptoKeyPath = &v
	}
	if v, ok := os.LookupEnv(EnvRateLimitVarName); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			e.RateLimit = &n
		}
	}
	return e, nil
}
