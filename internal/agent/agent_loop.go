package agent

import (
	"context"
	"sync"
	"time"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/collector"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/models"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/sender"
)

type AgentLoopConfig struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
	Iterations     int // 0 — бесконечно
	RateLimit      int
}

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
		ticker := time.NewTicker(cfg.PollInterval)
		defer ticker.Stop()
		for i := 0; cfg.Iterations == 0 || i < cfg.Iterations; i++ {
			collector.Collect()
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
		cancel()
	}()

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
			sender := s
			select {
			case tasks <- func() { sender.Send(nil) }:
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
				sender := s
				select {
				case tasks <- func() { sender.Send([]*models.Metrics{metric}) }:
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
