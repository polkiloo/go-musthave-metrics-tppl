package test

import (
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/models"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/storage"
)

type FakeStorage struct {
	gauges   map[string]float64
	counters map[string]int64
}

func NewFakeStorage() *FakeStorage {
	return &FakeStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

var _ storage.MetricStorage = (*FakeStorage)(nil)

func (f *FakeStorage) UpdateGauge(name string, value float64) { f.gauges[name] = value }
func (f *FakeStorage) UpdateCounter(name string, delta int64) { f.counters[name] += delta }
func (f *FakeStorage) GetGauge(name string) (float64, error)  { return f.gauges[name], nil }
func (f *FakeStorage) GetCounter(name string) (int64, error)  { return f.counters[name], nil }
func (f *FakeStorage) SetGauge(name string, value float64)    { f.gauges[name] = value }
func (f *FakeStorage) SetCounter(name string, value int64)    { f.counters[name] = value }
func (f *FakeStorage) AllGauges() map[string]float64          { return f.gauges }
func (f *FakeStorage) AllCounters() map[string]int64          { return f.counters }

type FakeBatchStore struct {
	*FakeStorage
	Got []models.Metrics
	Err error
}

func (f *FakeBatchStore) UpdateBatch(ms []models.Metrics) error {
	f.Got = append([]models.Metrics(nil), ms...)
	return f.Err
}

type FakeNoBatchStore struct {
	*FakeStorage
}
