package storage

import "sync"

type MemStorage struct {
	mu       sync.RWMutex
	gauges   map[string]float64
	counters map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (m *MemStorage) UpdateGauge(name string, value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.gauges[name] = value
}

func (m *MemStorage) UpdateCounter(name string, delta int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counters[name] += delta
}

func (m *MemStorage) GetGauge(name string) (float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.gauges[name]
	if !ok {
		return 0, ErrMetricNotFound
	}
	return v, nil
}

func (m *MemStorage) GetCounter(name string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.counters[name]
	if !ok {
		return 0, ErrMetricNotFound
	}
	return v, nil
}

var _ MetricStorage = NewMemStorage()
