package agent

import (
	"context"
	"sync"
	"time"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/collector"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/models"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/sender"
)

// AgentLoopConfig groups parameters controlling the agent execution loop.
type AgentLoopConfig struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
	Iterations     int // 0 — бесконечно
	RateLimit      int
}

// AgentLoopSleep collects metrics on a schedule and sends them via the provided senders.
func AgentLoopSleep(
	ctx context.Context,
	collector collector.CollectorInterface,
	senders []sender.SenderInterface,
	cfg AgentLoopConfig,
) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		pollGopsutil(ctx, collector, cfg.PollInterval)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(cfg.ReportInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				metrics := collector.Snapshot()
				sendMetrics(ctx, senders, metrics, cfg.RateLimit)
			case <-ctx.Done():
				return
			}
		}
	}()

	ticker := time.NewTicker(cfg.PollInterval)
	defer ticker.Stop()

	for i := 0; cfg.Iterations == 0 || i < cfg.Iterations; i++ {
		select {
		case <-ctx.Done():
			cancel()
			wg.Wait()
			return
		default:
		}

		collector.Collect()

		if cfg.Iterations != 0 && i+1 >= cfg.Iterations {
			break
		}

		select {
		case <-ctx.Done():
			cancel()
			wg.Wait()
			return
		case <-ticker.C:
		}
	}

	cancel()
	wg.Wait()
}

func sendMetrics(ctx context.Context, senders []sender.SenderInterface, metrics []*models.Metrics, limit int) {
	if limit <= 0 {
		limit = 1
	}
	tasks := make(chan func())
	var wg sync.WaitGroup
	for i := 0; i < limit; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range tasks {
				task()
			}
		}()
	}
	if len(metrics) == 0 {
		for _, s := range senders {
			sdr := s

			send := func(ms []*models.Metrics) {
				if cs, ok := sdr.(sender.ContextualSender); ok {
					cs.SendWithContext(ctx, ms)
					return
				}
				sdr.Send(ms)
			}
			select {
			case tasks <- func() { send(nil) }:
			case <-ctx.Done():
				close(tasks)
				wg.Wait()
				return
			}
		}
	} else {
		for _, m := range metrics {
			for _, s := range senders {
				metric := m
				sdr := s
				send := func(ms []*models.Metrics) {
					if cs, ok := sdr.(sender.ContextualSender); ok {
						cs.SendWithContext(ctx, ms)
						return
					}
					sdr.Send(ms)
				}
				select {
				case tasks <- func() { send([]*models.Metrics{metric}) }:
				case <-ctx.Done():
					close(tasks)
					wg.Wait()
					return
				}
			}
		}
	}
	close(tasks)
	wg.Wait()
}
