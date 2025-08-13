package main

import (
	"log"
	"time"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/agent"
	"go.uber.org/fx"
)

// ENV > FLAG > DEFAULT
func buildAppConfig(env EnvVars, fl FlagsArg) (agent.AppConfig, error) {
	cfg := agent.DefaultAppConfig

	if env.Host != "" {
		cfg.Host = env.Host
	} else if fl.Host != "" {
		cfg.Host = fl.Host
	}

	if env.Port != nil {
		cfg.Port = *env.Port
	} else if fl.Port != nil {
		cfg.Port = *fl.Port
	}

	if env.PollIntervalSec != nil {
		cfg.PollInterval = time.Duration(*env.PollIntervalSec) * time.Second
	} else if fl.PollIntervalSec != nil {
		cfg.PollInterval = time.Duration(*fl.PollIntervalSec) * time.Second
	}

	if env.ReportIntervalSec != nil {
		cfg.ReportInterval = time.Duration(*env.ReportIntervalSec) * time.Second
	} else if fl.ReportIntervalSec != nil {
		cfg.ReportInterval = time.Duration(*fl.ReportIntervalSec) * time.Second
	}

	return cfg, nil
}

func buildApp(cfg agent.AppConfig, opts ...fx.Option) *fx.App {
	base := fx.Options(
		fx.Provide(
			func() agent.AppConfig { return cfg },
			agent.ProvideCollector,
			agent.ProvideSender,
			agent.ProvideConfig,
		),
		fx.Invoke(agent.RunAgent),
	)
	if len(opts) > 0 {
		base = fx.Options(append([]fx.Option{base}, opts...)...)
	}
	return fx.New(base)
}

func main() {
	envVars := getEnvVars()
	flagArgs, err := parseFlags()
	if err != nil {
		log.Fatalf("flags: %v", err)
	}
	cfg, err := buildAppConfig(envVars, flagArgs)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	app := buildApp(cfg)
	app.Run()
}
