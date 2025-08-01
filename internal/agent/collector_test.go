package agent_test

import (
	"testing"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/agent"
	models "github.com/polkiloo/go-musthave-metrics-tppl/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollector_Collect_KnownGauges(t *testing.T) {
	c := agent.NewCollector()
	c.Collect()
	expected := []string{
		"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc", "HeapIdle",
		"HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "LastGC", "Lookups", "MCacheInuse",
		"MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC", "NumGC",
		"OtherSys", "PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc",
	}
	for _, name := range expected {
		_, ok := c.Gauges[name]
		assert.True(t, ok, "missing gauge: %s", name)
	}
}

func TestCollector_Collect_PollCountIncrementsViaMap(t *testing.T) {
	c := agent.NewCollector()
	c.Collect()
	require.Equal(t, models.Counter(1), c.Counters["PollCount"])
	c.Collect()
	require.Equal(t, models.Counter(2), c.Counters["PollCount"])
}

func TestCollector_Snapshot_ReturnsCopies(t *testing.T) {
	c := agent.NewCollector()

	c.Gauges["Alloc"] = 123.45
	c.Gauges["RandomValue"] = 42
	c.Counters["PollCount"] = 3

	g, cn := c.Snapshot()

	assert.Equal(t, c.Gauges, g, "Gauges map values should match")
	assert.Equal(t, c.Counters, cn, "Counters map values should match")

	g["Alloc"] = 0
	cn["PollCount"] = 100
	assert.NotEqual(t, c.Gauges["Alloc"], g["Alloc"], "Snapshot must return a copy, not reference")
	assert.NotEqual(t, c.Counters["PollCount"], cn["PollCount"], "Snapshot must return a copy, not reference")
}

func TestCollector_Snapshot_Empty(t *testing.T) {
	c := agent.NewCollector()
	g, cn := c.Snapshot()
	assert.Empty(t, g)
	assert.Empty(t, cn)
}
