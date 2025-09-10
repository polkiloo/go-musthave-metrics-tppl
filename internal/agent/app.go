package agent

import (
	"context"
	"time"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/collector"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/compression"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/logger"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/sender"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/sign"
	"go.uber.org/fx"
)

type AppConfig struct {
	Host           string
	Port           int
	ReportInterval time.Duration
	PollInterval   time.Duration
	LoopIterations int
	SignKey        sign.SignKey
}

const (
	DefaultAppHost           = "localhost"
	DefaultAppPort           = 8080
	DefaultAppReportInterval = 10 * time.Second
	DefaultAppPollInterval   = 2 * time.Second
	DefaultLoopIterations    = 0
)

func RunAgent(
	ctx context.Context,
	lc fx.Lifecycle,
	collector collector.CollectorInterface,
	senders []sender.SenderInterface,
	cfg AgentLoopConfig,
) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go AgentLoopSleep(ctx, collector, senders, cfg)
			return nil
		},
		OnStop: func(context.Context) error {
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

func ProvideSender(cfg AppConfig, l logger.Logger, c compression.Compressor) ([]sender.SenderInterface, error) {
	senders := make([]sender.SenderInterface, 0, 2)
	senders = append(senders,
		sender.NewPlainSender(cfg.Host, cfg.Port, nil, l, cfg.SignKey),
		sender.NewJSONSender(cfg.Host, cfg.Port, nil, l, c, cfg.SignKey),
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
