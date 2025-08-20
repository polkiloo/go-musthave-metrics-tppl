package agent

import (
	"context"
	"time"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/collector"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/sender"
)

type AgentLoopConfig struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
	Iterations     int // 0 — бесконечно
}

func AgentLoopSleep(
	ctx context.Context,
	collector collector.CollectorInterface,
	senders []sender.SenderInterface,
	cfg AgentLoopConfig,
) {
	lastReport := time.Now()

	ticker := time.NewTicker(cfg.PollInterval)
	defer ticker.Stop()

	for i := 0; cfg.Iterations == 0 || i < cfg.Iterations; i++ {
		collector.Collect()

		if time.Since(lastReport) >= cfg.ReportInterval {
			m := collector.Snapshot()
			for _, s := range senders {
				s.Send(m)
			}
			lastReport = time.Now()
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
