package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/collector"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

func cpuPercentGenerator(ctx context.Context, interval time.Duration) <-chan []float64 {
	ch := make(chan []float64)
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		defer close(ch)
		for {
			perc, err := cpu.PercentWithContext(ctx, 0, true)
			if err == nil {
				select {
				case ch <- perc:
				case <-ctx.Done():
					return
				}
			}
			select {
			case <-ticker.C:
			case <-ctx.Done():
				return
			}
		}
	}()
	return ch
}

func pollGopsutil(ctx context.Context, c collector.CollectorInterface, interval time.Duration) {
	if interval <= 0 {
		interval = time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	updateMetrics := func() {
		if vm, err := mem.VirtualMemoryWithContext(ctx); err == nil {
			c.SetGauge("TotalMemory", float64(vm.Total))
			c.SetGauge("FreeMemory", float64(vm.Free))
		}
		if percents, err := cpu.PercentWithContext(ctx, 0, true); err == nil {
			for i, p := range percents {
				c.SetGauge(fmt.Sprintf("CPUutilization%d", i+1), p)
			}
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			updateMetrics()
		}
	}
}
