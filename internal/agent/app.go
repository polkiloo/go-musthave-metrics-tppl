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

// AppConfig contains runtime configuration for the agent application.
// generate:reset
type AppConfig struct {
	Host           string
	Port           int
	ReportInterval time.Duration
	PollInterval   time.Duration
	LoopIterations int
	SignKey        sign.SignKey
	RateLimit      int
}

const (
	// DefaultAppHost is the default server host the agent connects to.
	DefaultAppHost = "localhost"
	// DefaultAppPort is the default server port the agent connects to.
	DefaultAppPort = 8080
	// DefaultAppReportInterval defines how often the agent reports metrics by default.
	DefaultAppReportInterval = 10 * time.Second
	// DefaultAppPollInterval defines how often the agent polls metrics by default.
	DefaultAppPollInterval = 2 * time.Second
	// DefaultLoopIterations controls the number of collection cycles (0 = infinite).
	DefaultLoopIterations = 0
	// DefaultRateLimit controls concurrency for metric sending.
	DefaultRateLimit = 1
)

// RunAgent launches the agent loop when the fx application starts.
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

// ModuleAgent registers the agent runtime with the fx container.
var ModuleAgent = fx.Module("agent",
	fx.Invoke(
		RunAgent,
	),
)

// ProvideCollector constructs the metrics collector used by the agent.
func ProvideCollector(cfg AppConfig, l logger.Logger) (collector.CollectorInterface, error) {
	return collector.NewCollector(), nil
}

// ModuleCollector provides the collector dependency via fx.
var ModuleCollector = fx.Module("collector",
	fx.Provide(
		ProvideCollector,
	),
)

// ProvideSender constructs both plain-text and JSON senders for the agent.
func ProvideSender(cfg AppConfig, l logger.Logger, c compression.Compressor) ([]sender.SenderInterface, error) {
	senders := make([]sender.SenderInterface, 0, 2)
	senders = append(senders,
		sender.NewPlainSender(cfg.Host, cfg.Port, nil, l, cfg.SignKey),
		sender.NewJSONSender(cfg.Host, cfg.Port, nil, l, c, cfg.SignKey),
	)
	return senders, nil
}

// ModuleSender provides the sender dependencies via fx.
var ModuleSender = fx.Module("sender",
	fx.Provide(
		ProvideSender,
	),
)

// ProvideAgentLoopConfig derives the loop configuration from the agent config.
func ProvideAgentLoopConfig(cfg AppConfig) AgentLoopConfig {
	return AgentLoopConfig{
		PollInterval:   cfg.PollInterval,
		ReportInterval: cfg.ReportInterval,
		Iterations:     cfg.LoopIterations,
		RateLimit:      cfg.RateLimit,
	}
}

// ModuleLoopConfig provides AgentLoopConfig to the fx container.
var ModuleLoopConfig = fx.Module("loopconfig",
	fx.Provide(
		ProvideAgentLoopConfig,
	),
)
