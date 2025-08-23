package test

import (
	"sync"
	"sync/atomic"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/models"
)

type FakeCollector struct {
	mu        sync.RWMutex
	items     []*models.Metrics
	Collected int32
}

func NewFakeCollector(ms ...*models.Metrics) *FakeCollector {
	f := &FakeCollector{}
	f.SetMetrics(ms...)
	return f
}

func (m *FakeCollector) SetMetrics(ms ...*models.Metrics) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items = make([]*models.Metrics, len(ms))
	for i, src := range ms {
		m.items[i] = cloneMetrics(src)
	}
}

func cloneMetrics(src *models.Metrics) *models.Metrics {
	if src == nil {
		return nil
	}
	var vp *float64
	var dp *int64
	if src.Value != nil {
		v := *src.Value
		vp = &v
	}
	if src.Delta != nil {
		d := *src.Delta
		dp = &d
	}
	return &models.Metrics{
		ID:    src.ID,
		MType: src.MType,
		Delta: dp,
		Value: vp,
		Hash:  src.Hash,
	}
}
func (m *FakeCollector) Collect() { atomic.AddInt32(&m.Collected, 1) }

func (m *FakeCollector) Snapshot() (metrics []*models.Metrics) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metrics = make([]*models.Metrics, len(m.items))
	for i, src := range m.items {
		metrics[i] = cloneMetrics(src)
	}
	return
}
