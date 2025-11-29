package collector

import (
	"math/rand"
	"runtime"
	"sync"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/models"
)

// CollectorInterface defines behaviour required from metric collectors.
type CollectorInterface interface {
	Collect()
	Snapshot() []*models.Metrics
	SetGauge(name string, value float64)
}

// Collector gathers runtime metrics and stores them in-memory.
type Collector struct {
	mu      sync.RWMutex
	metrics map[string]*models.Metrics
}

// NewCollector constructs a Collector ready to gather runtime metrics.
func NewCollector() *Collector {
	return &Collector{
		metrics: make(map[string]*models.Metrics),
	}
}

// MetricGetter extracts a metric value from runtime.MemStats.
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

// Collect gathers the latest runtime metrics and stores them internally.
func (c *Collector) Collect() {
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)

	c.mu.Lock()
	defer c.mu.Unlock()

	for name, get := range gaugeGetters {
		val := get(&rtm)
		f := float64(val)
		if m, ok := c.metrics[name]; ok && m.MType == models.GaugeType && m.Value != nil {
			*m.Value = f
			continue
		}
		if m, err := models.NewGaugeMetrics(name, &f); err == nil {
			c.metrics[name] = m
		}
	}

	{
		val := rand.Float64() * 100
		if m, ok := c.metrics["RandomValue"]; ok && m.MType == models.GaugeType && m.Value != nil {
			*m.Value = val
		} else {
			v := val
			if m, err := models.NewGaugeMetrics("RandomValue", &v); err == nil {
				c.metrics["RandomValue"] = m
			}
		}
	}

	for name, get := range counterGetters {
		inc := get(&rtm)
		i := int64(inc)
		if m, ok := c.metrics[name]; ok && m.MType == models.CounterType && m.Delta != nil {
			*m.Delta += i
			continue
		}
		if m, err := models.NewCounterMetrics(name, &i); err == nil {
			c.metrics[name] = m
		}
	}
}

// Snapshot returns deep copies of all collected metrics.
func (c *Collector) Snapshot() []*models.Metrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	out := make([]*models.Metrics, 0, len(c.metrics))
	for _, m := range c.metrics {
		out = append(out, cloneMetrics(m))
	}
	return out
}

// SetGauge sets a specific gauge metric to the provided value.
func (c *Collector) SetGauge(name string, value float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if m, ok := c.metrics[name]; ok && m.MType == models.GaugeType && m.Value != nil {
		*m.Value = value
		return
	}
	if m, err := models.NewGaugeMetrics(name, &value); err == nil {
		c.metrics[name] = m
	}
}

func cloneMetrics(m *models.Metrics) *models.Metrics {
	if m == nil {
		return nil
	}
	var (
		vp *float64
		dp *int64
	)
	if m.Value != nil {
		v := *m.Value
		vp = &v
	}
	if m.Delta != nil {
		d := *m.Delta
		dp = &d
	}
	return &models.Metrics{
		ID:    m.ID,
		MType: m.MType,
		Delta: dp,
		Value: vp,
		Hash:  m.Hash,
	}
}

var _ CollectorInterface = NewCollector()
