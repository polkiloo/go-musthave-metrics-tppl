package agent

import (
	"time"
)

type AgentLoopConfig struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
	Iterations     int // 0 — бесконечно
}

func AgentLoopSleep(
	collector CollectorInterface,
	sender SenderInterface,
	cfg AgentLoopConfig,
) {
	lastReport := time.Now()
	i := 0
	for cfg.Iterations == 0 || i < cfg.Iterations {
		collector.Collect()
		if time.Since(lastReport) >= cfg.ReportInterval {
			g, c := collector.Snapshot()
			sender.Send(g, c)
			lastReport = time.Now()
		}
		time.Sleep(cfg.PollInterval)
		i++
	}
}
