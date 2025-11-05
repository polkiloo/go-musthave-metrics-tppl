package storage

import (
	"sync"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/models"
)

// MemStorageT stores metrics of any comparable type in memory with concurrency safety.
type MemStorageT[T comparable] struct {
	mu   sync.RWMutex
	data map[string]T
}

// NewMemStorageT creates a new thread-safe storage for the provided type.
func NewMemStorageT[T comparable]() *MemStorageT[T] {
	return &MemStorageT[T]{data: make(map[string]T)}
}

// Update sets the value of the metric with the given name.
func (m *MemStorageT[T]) Update(name string, value T) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[name] = value
}

// Get retrieves the value for the metric with the given name.
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

// Number constrains numeric types supported by the storage.
type Number interface {
	~int64 | ~float64
}

// NumMemStorage wraps MemStorageT for numeric types and adds atomic addition.
type NumMemStorage[T Number] struct {
	*MemStorageT[T]
}

// NewNumMemStorage constructs numeric storage capable of accumulation operations.
func NewNumMemStorage[T Number]() *NumMemStorage[T] {
	return &NumMemStorage[T]{NewMemStorageT[T]()}
}

// Add increments the existing value of a metric by delta.
func (m *NumMemStorage[T]) Add(name string, delta T) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[name] += delta
}

// MemStorage keeps gauge and counter metrics in memory.
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

// UpdateGauge stores the latest gauge value.
func (m *MemStorage) UpdateGauge(name string, value float64) {
	m.gauges.Update(name, value)
}

// UpdateCounter increments the counter by the provided delta.
func (m *MemStorage) UpdateCounter(name string, delta int64) {
	m.counters.Add(name, delta)
}

// GetGauge retrieves a gauge value.
func (m *MemStorage) GetGauge(name string) (float64, error) {
	return m.gauges.Get(name)
}

// GetCounter retrieves a counter value.
func (m *MemStorage) GetCounter(name string) (int64, error) {
	return m.counters.Get(name)
}

// SetGauge overwrites a gauge without additional processing.
func (m *MemStorage) SetGauge(name string, value float64) {
	m.gauges.Update(name, value)
}

// SetCounter overwrites a counter without additional processing.
func (m *MemStorage) SetCounter(name string, value int64) {
	m.counters.Update(name, value)
}

// AllGauges returns a snapshot of all gauges.
func (m *MemStorage) AllGauges() map[string]float64 {
	m.gauges.mu.RLock()
	defer m.gauges.mu.RUnlock()
	res := make(map[string]float64, len(m.gauges.data))
	for k, v := range m.gauges.data {
		res[k] = v
	}
	return res
}

// AllCounters returns a snapshot of all counters.
func (m *MemStorage) AllCounters() map[string]int64 {
	m.counters.mu.RLock()
	defer m.counters.mu.RUnlock()
	res := make(map[string]int64, len(m.counters.data))
	for k, v := range m.counters.data {
		res[k] = v
	}
	return res
}

// Snapshot returns all metrics as model instances suitable for serialisation.
func (m *MemStorage) Snapshot() []models.Metrics {
	metrics := make([]models.Metrics, 0, len(m.gauges.data)+len(m.counters.data))
	gaugeValues := make([]float64, 0, len(m.gauges.data))
	counterValues := make([]int64, 0, len(m.counters.data))

	m.gauges.mu.RLock()
	for name, value := range m.gauges.data {
		gaugeValues = append(gaugeValues, value)
		v := &gaugeValues[len(gaugeValues)-1]
		metrics = append(metrics, models.Metrics{
			ID:    name,
			MType: models.GaugeType,
			Value: v,
		})
	}
	m.gauges.mu.RUnlock()

	m.counters.mu.RLock()
	for name, value := range m.counters.data {
		counterValues = append(counterValues, value)
		v := &counterValues[len(counterValues)-1]
		metrics = append(metrics, models.Metrics{
			ID:    name,
			MType: models.CounterType,
			Delta: v,
		})
	}
	m.counters.mu.RUnlock()

	return metrics
}

// UpdateBatch applies a batch of metric updates in a single pass.
func (m *MemStorage) UpdateBatch(metrics []models.Metrics) error {
	for i := range metrics {
		mt := &metrics[i]

		if mt.MType == models.GaugeType {
			if mt.Value != nil {
				m.UpdateGauge(mt.ID, *mt.Value)
			}
			continue
		}
		if mt.MType == models.CounterType {
			if mt.Delta != nil {
				m.UpdateCounter(mt.ID, *mt.Delta)
			}
			continue
		}
	}
	return nil
}

var _ MetricStorage = NewMemStorage()
