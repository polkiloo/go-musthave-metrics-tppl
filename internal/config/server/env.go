package servercfg

import (
	"os"
	"strconv"

	commoncfg "github.com/polkiloo/go-musthave-metrics-tppl/internal/config/common"
)

const (
	EnvAddressVarName       = "ADDRESS"
	EnvStoreIntervalVarName = "STORE_INTERVAL"
	EnvFileStorageVarName   = "FILE_STORAGE_PATH"
	EnvRestoreVarName       = "RESTORE"
	EnvKeyVarName           = "KEY"
)

type ServerEnvVars struct {
	Host          string
	Port          *int
	StoreInterval *int
	FileStorage   string
	Restore       *bool
	SignKey       string
}

func getEnvVars() (ServerEnvVars, error) {
	hp, _ := commoncfg.ReadHostPortEnv(EnvAddressVarName)

	var interval *int
	if v := os.Getenv(EnvStoreIntervalVarName); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			interval = &i
		}
	}

	var restore *bool
	if v := os.Getenv(EnvRestoreVarName); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			restore = &b
		}
	}

	filePath := os.Getenv(EnvFileStorageVarName)

	key := os.Getenv(EnvKeyVarName)

	return ServerEnvVars{Host: hp.Host, Port: hp.Port, StoreInterval: interval, FileStorage: filePath, Restore: restore, SignKey: key}, nil
}
