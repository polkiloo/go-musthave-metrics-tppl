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

	for _, metric := range models.RuntimeMetrics {
		switch metric.Name {
		case "Alloc":
			c.Gauges[metric.Name] = models.Gauge(rtm.Alloc)
		case "BuckHashSys":
			c.Gauges[metric.Name] = models.Gauge(rtm.BuckHashSys)
		case "Frees":
			c.Gauges[metric.Name] = models.Gauge(rtm.Frees)
		case "GCCPUFraction":
			c.Gauges[metric.Name] = models.Gauge(rtm.GCCPUFraction)
		case "GCSys":
			c.Gauges[metric.Name] = models.Gauge(rtm.GCSys)
		case "HeapAlloc":
			c.Gauges[metric.Name] = models.Gauge(rtm.HeapAlloc)
		case "HeapIdle":
			c.Gauges[metric.Name] = models.Gauge(rtm.HeapIdle)
		case "HeapInuse":
			c.Gauges[metric.Name] = models.Gauge(rtm.HeapInuse)
		case "HeapObjects":
			c.Gauges[metric.Name] = models.Gauge(rtm.HeapObjects)
		case "HeapReleased":
			c.Gauges[metric.Name] = models.Gauge(rtm.HeapReleased)
		case "HeapSys":
			c.Gauges[metric.Name] = models.Gauge(rtm.HeapSys)
		case "LastGC":
			c.Gauges[metric.Name] = models.Gauge(rtm.LastGC)
		case "Lookups":
			c.Gauges[metric.Name] = models.Gauge(rtm.Lookups)
		case "MCacheInuse":
			c.Gauges[metric.Name] = models.Gauge(rtm.MCacheInuse)
		case "MCacheSys":
			c.Gauges[metric.Name] = models.Gauge(rtm.MCacheSys)
		case "MSpanInuse":
			c.Gauges[metric.Name] = models.Gauge(rtm.MSpanInuse)
		case "MSpanSys":
			c.Gauges[metric.Name] = models.Gauge(rtm.MSpanSys)
		case "Mallocs":
			c.Gauges[metric.Name] = models.Gauge(rtm.Mallocs)
		case "NextGC":
			c.Gauges[metric.Name] = models.Gauge(rtm.NextGC)
		case "NumForcedGC":
			c.Gauges[metric.Name] = models.Gauge(rtm.NumForcedGC)
		case "NumGC":
			c.Gauges[metric.Name] = models.Gauge(rtm.NumGC)
		case "OtherSys":
			c.Gauges[metric.Name] = models.Gauge(rtm.OtherSys)
		case "PauseTotalNs":
			c.Gauges[metric.Name] = models.Gauge(rtm.PauseTotalNs)
		case "StackInuse":
			c.Gauges[metric.Name] = models.Gauge(rtm.StackInuse)
		case "StackSys":
			c.Gauges[metric.Name] = models.Gauge(rtm.StackSys)
		case "Sys":
			c.Gauges[metric.Name] = models.Gauge(rtm.Sys)
		case "TotalAlloc":
			c.Gauges[metric.Name] = models.Gauge(rtm.TotalAlloc)
		case "RandomValue":
			c.Gauges[metric.Name] = models.Gauge(rand.Float64() * 100)
		case "PollCount":
			c.Counters[metric.Name]++
		}
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
