package servercfg

import (
	"fmt"
	"strconv"
	"time"
)

type serverFileConfig struct {
	Address       *string `json:"address"`
	Restore       *bool   `json:"restore"`
	StoreInterval *string `json:"store_interval"`
	StoreFile     *string `json:"store_file"`
	DatabaseDSN   *string `json:"database_dsn"`
	Key           *string `json:"key"`
	AuditFile     *string `json:"audit_file"`
	AuditURL      *string `json:"audit_url"`
	CryptoKey     *string `json:"crypto_key"`
}

func parseDurationSeconds(raw string) (int, error) {
	if raw == "" {
		return 0, fmt.Errorf("empty duration")
	}
	if d, err := time.ParseDuration(raw); err == nil {
		return int(d / time.Second), nil
	}
	if v, err := strconv.Atoi(raw); err == nil {
		return v, nil
	}
	return 0, fmt.Errorf("invalid duration: %s", raw)
}
