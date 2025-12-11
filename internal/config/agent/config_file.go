package agentcfg

import (
	"fmt"
	"strconv"
	"time"
)

type agentFileConfig struct {
	Address        *string `json:"address"`
	ReportInterval *string `json:"report_interval"`
	PollInterval   *string `json:"poll_interval"`
	Key            *string `json:"key"`
	RateLimit      *int    `json:"rate_limit"`
	CryptoKey      *string `json:"crypto_key"`
}

func parseDuration(raw string) (time.Duration, error) {
	if d, err := time.ParseDuration(raw); err == nil {
		return d, nil
	}
	if v, err := strconv.Atoi(raw); err == nil {
		return time.Duration(v) * time.Second, nil
	}
	return 0, fmt.Errorf("invalid duration: %s", raw)
}
