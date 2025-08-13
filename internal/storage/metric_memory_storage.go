package storage

import "sync"

type MemStorageT[T comparable] struct {
	mu   sync.RWMutex
	data map[string]T
}

func NewMemStorageT[T comparable]() *MemStorageT[T] {
	return &MemStorageT[T]{data: make(map[string]T)}
}

func (m *MemStorageT[T]) Update(name string, value T) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[name] = value
}

func (m *MemStorageT[T]) Get(name string) (T, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.data[name]
	if !ok {
		var zero T
		return zero, ErrMetricNotFound
	}
	return v, nil
}

type Number interface {
	~int64 | ~float64
}

type NumMemStorage[T Number] struct {
	*MemStorageT[T]
}

func NewNumMemStorage[T Number]() *NumMemStorage[T] {
	return &NumMemStorage[T]{NewMemStorageT[T]()}
}

func (m *NumMemStorage[T]) Add(name string, delta T) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[name] += delta
}

type MemStorage struct {
	gauges   *NumMemStorage[float64]
	counters *NumMemStorage[int64]
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   NewNumMemStorage[float64](),
		counters: NewNumMemStorage[int64](),
	}
}

func (m *MemStorage) UpdateGauge(name string, value float64) {
	m.gauges.Update(name, value)
}

func (m *MemStorage) UpdateCounter(name string, delta int64) {
	m.counters.Add(name, delta)
}

func (m *MemStorage) GetGauge(name string) (float64, error) {
	return m.gauges.Get(name)
}

func (m *MemStorage) GetCounter(name string) (int64, error) {
	return m.counters.Get(name)
}

var _ MetricStorage = NewMemStorage()
