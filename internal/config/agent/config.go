package agentcfg

import (
	"fmt"
	"os"
	"time"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/agent"
	commoncfg "github.com/polkiloo/go-musthave-metrics-tppl/internal/config/common"
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
		CryptoKeyPath:  agent.DefaultCryptoKeyPath,
		GRPCHost:       agent.DefaultGRPCHost,
		GRPCPort:       agent.DefaultGRPCPort,
	}
	cfg := defaultAppConfig

	envVars, _ := getEnvVars()
	flagArgs, _ := parseFlags()

	configPath := os.Getenv("CONFIG")
	if configPath == "" {
		configPath = flagArgs.ConfigPath
	}

	var fileCfg agentFileConfig
	if err := commoncfg.LoadConfigFile(configPath, &fileCfg); err != nil {
		return cfg, err
	}

	if fileCfg.Address != nil {
		hp, err := commoncfg.ParseAddressFlag(*fileCfg.Address, true)
		if err != nil {
			return cfg, fmt.Errorf("config address: %w", err)
		}
		if hp.Host != "" {
			cfg.Host = hp.Host
		}
		if hp.Port != nil {
			cfg.Port = *hp.Port
		}
	}

	if fileCfg.GRPCAddress != nil {
		hp, err := commoncfg.ParseAddressFlag(*fileCfg.GRPCAddress, true)
		if err != nil {
			return cfg, fmt.Errorf("config grpc_address: %w", err)
		}
		if hp.Host != "" {
			cfg.GRPCHost = hp.Host
		}
		if hp.Port != nil {
			cfg.GRPCPort = *hp.Port
		}
	}

	if fileCfg.ReportInterval != nil {
		d, err := parseDuration(*fileCfg.ReportInterval)
		if err != nil {
			return cfg, fmt.Errorf("config report_interval: %w", err)
		}
		cfg.ReportInterval = d
	}

	if fileCfg.PollInterval != nil {
		d, err := parseDuration(*fileCfg.PollInterval)
		if err != nil {
			return cfg, fmt.Errorf("config poll_interval: %w", err)
		}
		cfg.PollInterval = d
	}

	if fileCfg.Key != nil {
		cfg.SignKey = sign.SignKey(*fileCfg.Key)
	}

	if fileCfg.RateLimit != nil {
		cfg.RateLimit = *fileCfg.RateLimit
	}

	if fileCfg.CryptoKey != nil {
		cfg.CryptoKeyPath = *fileCfg.CryptoKey
	}

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

	if envVars.GRPCHost != "" {
		cfg.GRPCHost = envVars.GRPCHost
	} else if flagArgs.GRPCAddress.Host != "" {
		cfg.GRPCHost = flagArgs.GRPCAddress.Host
	}

	if envVars.GRPCPort != nil {
		cfg.GRPCPort = *envVars.GRPCPort
	} else if flagArgs.GRPCAddress.Port != nil {
		cfg.GRPCPort = *flagArgs.GRPCAddress.Port
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

	if envVars.CryptoKeyPath != nil {
		cfg.CryptoKeyPath = *envVars.CryptoKeyPath
	} else if flagArgs.CryptoKey != "" {
		cfg.CryptoKeyPath = flagArgs.CryptoKey
	}
	return cfg, nil
}

var Module = fx.Module(
	"agent-config",
	fx.Provide(buildAgentConfig),
)
