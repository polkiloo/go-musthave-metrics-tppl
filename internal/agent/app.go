package agent

import (
	"context"
	"time"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/collector"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/logger"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/sender"
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

func RunAgent(
	lc fx.Lifecycle,
	collector collector.CollectorInterface,
	senders []sender.SenderInterface,
	cfg AgentLoopConfig,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go AgentLoopSleep(collector, senders, cfg)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})
}

var ModuleAgent = fx.Module("agent",
	fx.Invoke(
		RunAgent,
	),
)

func ProvideCollector(cfg AppConfig, l logger.Logger) (collector.CollectorInterface, error) {
	return collector.NewCollector(), nil
}

var ModuleCollector = fx.Module("collector",
	fx.Provide(
		ProvideCollector,
	),
)

func ProvideSender(cfg AppConfig, l logger.Logger) ([]sender.SenderInterface, error) {
	senders := make([]sender.SenderInterface, 0, 2)
	senders = append(senders,
		sender.NewPlainSender(cfg.Host, cfg.Port, nil, l),
		sender.NewJSONSender(cfg.Host, cfg.Port, nil, l),
	)
	return senders, nil
}

var ModuleSender = fx.Module("sender",
	fx.Provide(
		ProvideSender,
	),
)

func ProvideAgentLoopConfig(cfg AppConfig) AgentLoopConfig {
	return AgentLoopConfig{
		PollInterval:   cfg.PollInterval,
		ReportInterval: cfg.ReportInterval,
		Iterations:     cfg.LoopIterations,
	}
}

var ModuleLoopConfig = fx.Module("loopconfig",
	fx.Provide(
		ProvideAgentLoopConfig,
	),
)
