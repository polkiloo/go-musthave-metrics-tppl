package servercfg

import (
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/server"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/sign"
	"go.uber.org/fx"
)

func buildServerConfig() (server.AppConfig, error) {
	var defaultAppConfig = server.AppConfig{
		Host:            server.DefaultAppHost,
		Port:            server.DefaultAppPort,
		StoreInterval:   server.DefaultStoreInterval,
		FileStoragePath: server.DefaultFileStoragePath,
		Restore:         server.DefaultRestore,
	}

	cfg := defaultAppConfig

	envVars, _ := getEnvVars()
	flagArgs, _ := parseFlags()

	if envVars.Host != "" {
		cfg.Host = envVars.Host
	} else if flagArgs.addressFlag.Host != "" {
		cfg.Host = flagArgs.addressFlag.Host
	}

	if envVars.Port != nil {
		cfg.Port = *envVars.Port
	} else if flagArgs.addressFlag.Port != nil {
		cfg.Port = *flagArgs.addressFlag.Port
	}

	if envVars.StoreInterval != nil {
		cfg.StoreInterval = *envVars.StoreInterval
	} else if flagArgs.storeInterval != nil {
		cfg.StoreInterval = *flagArgs.storeInterval
	}

	if envVars.FileStorage != "" {
		cfg.FileStoragePath = envVars.FileStorage
	} else if flagArgs.fileStorage != "" {
		cfg.FileStoragePath = flagArgs.fileStorage
	}

	if envVars.Restore != nil {
		cfg.Restore = *envVars.Restore
	} else if flagArgs.restore != nil {
		cfg.Restore = *flagArgs.restore
	}

	if envVars.SignKey != "" {
		cfg.SignKey = sign.SignKey(envVars.SignKey)
	} else if flagArgs.SignKey != "" {
		cfg.SignKey = sign.SignKey(flagArgs.SignKey)
	}

	return cfg, nil
}

var Module = fx.Module(
	"server-config",
	fx.Provide(
		buildServerConfig,
		func(c server.AppConfig) *server.AppConfig { return &c },
		func(c server.AppConfig) sign.SignKey { return sign.SignKey(c.SignKey) },
	),
)
