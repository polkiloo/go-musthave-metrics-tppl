package storage

import "fmt"

var (
	ErrMetricNotFound = fmt.Errorf("metric not found")
)

type MetricStorage interface {
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, delta int64)
	GetGauge(name string) (float64, error)
	GetCounter(name string) (int64, error)
	SetGauge(name string, value float64)
	SetCounter(name string, value int64)
	AllGauges() map[string]float64
	AllCounters() map[string]int64
}
