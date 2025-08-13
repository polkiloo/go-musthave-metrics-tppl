package agent

import (
	"context"
	"time"

	"go.uber.org/fx"
)

type AppConfig struct {
	Host           string
	Port           int
	ReportInterval time.Duration
	PollInterval   time.Duration
	LoopIterations int
}

const (
	DefaultAppHost           = "localhost"
	DefaultAppPort           = 8080
	DefaultAppReportInterval = 10 * time.Second
	DefaultAppPollInterval   = 2 * time.Second
	DefaultLoopIterations    = 0
)

var DefaultAppConfig = AppConfig{
	Host:           DefaultAppHost,
	Port:           DefaultAppPort,
	PollInterval:   DefaultAppPollInterval,
	ReportInterval: DefaultAppReportInterval,
	LoopIterations: DefaultLoopIterations,
}

func ProvideCollector() CollectorInterface {
	return NewCollector()
}

func ProvideSender(args AppConfig) SenderInterface {
	return NewSender("http://"+args.Host, args.Port)
}

func ProvideConfig(args AppConfig) AgentLoopConfig {
	return AgentLoopConfig{
		PollInterval:   args.PollInterval,
		ReportInterval: args.ReportInterval,
		Iterations:     args.LoopIterations,
	}
}

func RunAgent(
	lc fx.Lifecycle,
	collector CollectorInterface,
	sender SenderInterface,
	cfg AgentLoopConfig,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go AgentLoopSleep(collector, sender, cfg)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})
}
