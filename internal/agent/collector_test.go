package agent_test

import (
	"testing"

	models "github.com/polkiloo/go-musthave-metrics-tppl/internal/model"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/agent"

	"github.com/stretchr/testify/assert"
)

func TestCollector_CollectAndSnapshot_AllDefinedMetrics(t *testing.T) {
	c := agent.NewCollector()
	c.Collect()

	gauges, counters := c.Snapshot()

	for _, metric := range models.RuntimeMetrics {
		switch metric.Type {
		case models.GaugeType:
			value, ok := gauges[metric.Name]
			assert.True(t, ok, "missing gauge: %s", metric.Name)
			assert.IsType(t, models.Gauge(0), value)
		case models.CounterType:
			value, ok := counters[metric.Name]
			assert.True(t, ok, "missing counter: %s", metric.Name)
			assert.IsType(t, models.Counter(0), value)
		}
	}
}
