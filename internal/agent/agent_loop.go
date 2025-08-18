package agent

import (
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
	collector collector.CollectorInterface,
	senders []sender.SenderInterface,
	cfg AgentLoopConfig,
) {
	lastReport := time.Now()
	i := 0
	for cfg.Iterations == 0 || i < cfg.Iterations {
		collector.Collect()
		if time.Since(lastReport) >= cfg.ReportInterval {
			m := collector.Snapshot()
			for _, sender := range senders {
				sender.Send(m)
			}

			lastReport = time.Now()
		}
		time.Sleep(cfg.PollInterval)
		i++
	}
}
