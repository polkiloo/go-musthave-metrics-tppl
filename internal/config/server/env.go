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
	EnvAuditFileVarName     = "AUDIT_FILE"
	EnvAuditURLVarName      = "AUDIT_URL"
	EnvCryptoKeyVarName     = "CRYPTO_KEY"
	EnvTrustedSubnet        = "TRUSTED_SUBNET"
	EnvGRPCAddressVarName   = "GRPC_ADDRESS"
)

type ServerEnvVars struct {
	Host          string
	Port          *int
	StoreInterval *int
	FileStorage   string
	Restore       *bool
	SignKey       string
	AuditFile     string
	AuditURL      string
	CryptoKey     string
	TrustedSubnet string
	GRPCHost      string
	GRPCPort      *int
}

func getEnvVars() (ServerEnvVars, error) {
	hp, _ := commoncfg.ReadHostPortEnv(EnvAddressVarName)
	grpcHP, _ := commoncfg.ReadHostPortEnv(EnvGRPCAddressVarName)

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

	return ServerEnvVars{
		Host:          hp.Host,
		Port:          hp.Port,
		StoreInterval: interval,
		FileStorage:   os.Getenv(EnvFileStorageVarName),
		Restore:       restore,
		SignKey:       os.Getenv(EnvKeyVarName),
		AuditFile:     os.Getenv(EnvAuditFileVarName),
		AuditURL:      os.Getenv(EnvAuditURLVarName),
		CryptoKey:     os.Getenv(EnvCryptoKeyVarName),
		TrustedSubnet: os.Getenv(EnvTrustedSubnet),
		GRPCHost:      grpcHP.Host,
		GRPCPort:      grpcHP.Port,
	}, nil
}
