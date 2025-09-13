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
	gen := cpuPercentGenerator(ctx, interval)
	for {
		select {
		case percents, ok := <-gen:
			if !ok {
				return
			}
			if vm, err := mem.VirtualMemoryWithContext(ctx); err == nil {
				c.SetGauge("TotalMemory", float64(vm.Total))
				c.SetGauge("FreeMemory", float64(vm.Free))
			}
			for i, p := range percents {
				c.SetGauge(fmt.Sprintf("CPUutilization%d", i+1), p)
			}
		case <-ctx.Done():
			return
		}
	}
}
