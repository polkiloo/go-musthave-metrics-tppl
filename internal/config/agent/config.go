package agentcfg

import (
	"time"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/agent"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/sign"
	"go.uber.org/fx"
)

// ENV > FLAG > DEFAULT
func buildAgentConfig() (agent.AppConfig, error) {
	var defaultAppConfig = agent.AppConfig{
		Host:           agent.DefaultAppHost,
		Port:           agent.DefaultAppPort,
		PollInterval:   agent.DefaultAppPollInterval,
		ReportInterval: agent.DefaultAppReportInterval,
		LoopIterations: agent.DefaultLoopIterations,
		RateLimit:      agent.DefaultRateLimit,
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

	if envVars.ReportIntervalSec != nil {
		cfg.ReportInterval = time.Duration(*envVars.ReportIntervalSec) * time.Second
	} else if flagArgs.ReportIntervalSec != nil {
		cfg.ReportInterval = time.Duration(*flagArgs.ReportIntervalSec) * time.Second
	}

	if envVars.PollIntervalSec != nil {
		cfg.PollInterval = time.Duration(*envVars.PollIntervalSec) * time.Second
	} else if flagArgs.PollIntervalSec != nil {
		cfg.PollInterval = time.Duration(*flagArgs.PollIntervalSec) * time.Second
	}

	if envVars.SignKey != nil {
		cfg.SignKey = sign.SignKey(*envVars.SignKey)
	} else if flagArgs.SignKey != "" {
		cfg.SignKey = sign.SignKey(flagArgs.SignKey)
	}

	if envVars.RateLimit != nil {
		cfg.RateLimit = *envVars.RateLimit
	} else if flagArgs.RateLimit != nil {
		cfg.RateLimit = *flagArgs.RateLimit
	}

	return cfg, nil
}

var Module = fx.Module(
	"agent-config",
	fx.Provide(buildAgentConfig),
)
