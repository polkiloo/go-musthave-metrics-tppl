package agent

import (
	"time"
)

func AgentLoopSleep(
	collector CollectorInterface,
	sender SenderInterface,
	pollInterval time.Duration,
	reportInterval time.Duration,
	iterations int, // 0 — бесконечно
) {
	lastReport := time.Now()
	i := 0
	for iterations == 0 || i < iterations {
		collector.Collect()
		if time.Since(lastReport) >= reportInterval {
			g, c := collector.Snapshot()
			sender.Send(g, c)
			lastReport = time.Now()
		}
		time.Sleep(pollInterval)
		i++
	}
}
