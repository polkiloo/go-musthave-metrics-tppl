package agent

import (
	"math/rand"
	"runtime"
	"sync"

	models "github.com/polkiloo/go-musthave-metrics-tppl/internal/model"
)

type CollectorInterface interface {
	Collect()
	Snapshot() (map[string]models.Gauge, map[string]models.Counter)
}

type Collector struct {
	mu       sync.RWMutex
	Gauges   map[string]models.Gauge
	Counters map[string]models.Counter
}

type MetricGetter[T any] func(*runtime.MemStats) T

var gaugeGetters = map[string]MetricGetter[models.Gauge]{
	"Alloc":         func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.Alloc) },
	"BuckHashSys":   func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.BuckHashSys) },
	"Frees":         func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.Frees) },
	"GCCPUFraction": func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.GCCPUFraction) },
	"GCSys":         func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.GCSys) },
	"HeapAlloc":     func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.HeapAlloc) },
	"HeapIdle":      func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.HeapIdle) },
	"HeapInuse":     func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.HeapInuse) },
	"HeapObjects":   func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.HeapObjects) },
	"HeapReleased":  func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.HeapReleased) },
	"HeapSys":       func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.HeapSys) },
	"LastGC":        func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.LastGC) },
	"Lookups":       func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.Lookups) },
	"MCacheInuse":   func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.MCacheInuse) },
	"MCacheSys":     func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.MCacheSys) },
	"MSpanInuse":    func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.MSpanInuse) },
	"MSpanSys":      func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.MSpanSys) },
	"Mallocs":       func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.Mallocs) },
	"NextGC":        func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.NextGC) },
	"NumForcedGC":   func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.NumForcedGC) },
	"NumGC":         func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.NumGC) },
	"OtherSys":      func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.OtherSys) },
	"PauseTotalNs":  func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.PauseTotalNs) },
	"StackInuse":    func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.StackInuse) },
	"StackSys":      func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.StackSys) },
	"Sys":           func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.Sys) },
	"TotalAlloc":    func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.TotalAlloc) },
}

var counterGetters = map[string]MetricGetter[models.Counter]{
	"PollCount": func(_ *runtime.MemStats) models.Counter { return 1 },
}

func NewCollector() *Collector {
	return &Collector{
		Gauges:   make(map[string]models.Gauge),
		Counters: make(map[string]models.Counter),
	}
}

func (c *Collector) Collect() {
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)

	c.mu.Lock()
	defer c.mu.Unlock()

	for name, getter := range gaugeGetters {
		c.Gauges[name] = getter(&rtm)
	}
	c.Gauges["RandomValue"] = models.Gauge(rand.Float64() * 100)
	for name, getter := range counterGetters {
		c.Counters[name] += getter(&rtm)
	}
}

func (c *Collector) Snapshot() (map[string]models.Gauge, map[string]models.Counter) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	g := make(map[string]models.Gauge, len(c.Gauges))
	for k, v := range c.Gauges {
		g[k] = v
	}

	cn := make(map[string]models.Counter, len(c.Counters))
	for k, v := range c.Counters {
		cn[k] = v
	}

	return g, cn
}

var _ CollectorInterface = NewCollector()
