package servercfg

import (
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/server"
	"go.uber.org/fx"
)

func buildServerConfig() (server.AppConfig, error) {
	cfg := server.DefaultAppConfig

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

	return cfg, nil
}

var Module = fx.Module(
	"server-config",
	fx.Provide(
		buildServerConfig,
		func(c server.AppConfig) *server.AppConfig { return &c },
	),
)
