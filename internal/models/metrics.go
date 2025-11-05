package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

type (
	// Gauge represents a floating-point metric value.
	Gauge float64

	// Counter represents an integer metric value that is typically aggregated.
	Counter int64
)

// MetricType describes the domain-specific type of a metric (gauge or counter).
type MetricType string

const (
	// GaugeType indicates gauge metrics (floating-point values).
	GaugeType MetricType = "gauge"
	// CounterType indicates counter metrics (integer values).
	CounterType MetricType = "counter"
)

// MetricTypes enumerates all supported metric types.
var MetricTypes = []MetricType{
	GaugeType,
	CounterType,
}

// IsValid reports whether the metric type is supported by the service.
func (t MetricType) IsValid() bool {
	switch t {
	case GaugeType, CounterType:
		return true
	default:
		return false
	}
}

// ParseMetricType converts a string into a MetricType and returns an error for unsupported values.
func ParseMetricType(s string) (MetricType, error) {
	mt := MetricType(s)
	if mt.IsValid() {
		return mt, nil
	}
	return "", ErrMetricInvalidType
}

// GaugeNames lists Go runtime metrics that are collected as gauges.
var GaugeNames = []string{
	"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys",
	"HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased",
	"HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys",
	"MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC",
	"NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys",
	"Sys", "TotalAlloc", "RandomValue",
}

// CounterNames lists metrics that are collected as counters.
var CounterNames = []string{
	"PollCount",
}

var (
	// GaugeSet provide quick membership checks for known runtime metrics.
	GaugeSet = make(map[string]struct{}, len(GaugeNames))
	// CounterSet provide quick membership checks for known runtime metrics.
	CounterSet = make(map[string]struct{}, len(CounterNames))
)

var (
	// ErrMetricUnknownName indicates that a metric with the provided name is not supported.
	ErrMetricUnknownName = errors.New("unknown metric name")
	// ErrMetricInvalidType indicates that the metric type is not recognised.
	ErrMetricInvalidType = errors.New("invalid metric type")
	// ErrMetricInvalidValueType reports that the value could not be parsed into the metric type.
	ErrMetricInvalidValueType = errors.New("invalid value type for metric")
	// ErrMetricMissingValue is returned when a required metric value is absent.
	ErrMetricMissingValue = errors.New("missing value for metric")
	// ErrMetricAmbiguousValue reports that both gauge and counter values were provided simultaneously.
	ErrMetricAmbiguousValue = errors.New("both gauge and counter values are set")
	// ErrMetricNameTypeMismatch indicates a mismatch between metric name and type.
	ErrMetricNameTypeMismatch = errors.New("metric name does not match the metric type")
)

func initSets() {
	GaugeSet = make(map[string]struct{}, len(GaugeNames))
	for _, n := range GaugeNames {
		GaugeSet[n] = struct{}{}
	}
	CounterSet = make(map[string]struct{}, len(CounterNames))
	for _, n := range CounterNames {
		CounterSet[n] = struct{}{}
	}
}

func init() {
	initSets()
}

// Metrics represents a single metric payload exchanged between client and server.
// Delta и Value объявлены через указатели, чтобы отличать значение "0" от не заданного значения и
// соответственно не кодировать его в структуру.
type Metrics struct {
	ID    string     `json:"id"`
	MType MetricType `json:"type"`
	Delta *int64     `json:"delta,omitempty"`
	Value *float64   `json:"value,omitempty"`
	Hash  string     `json:"hash,omitempty"`
}

// func IsGaugeName(name string) bool {
// _, ok := GaugeSet[name]
// return ok
// }

// func IsCounterName(name string) bool {
// _, ok := CounterSet[name]
// return ok
// }

// IsGauge reports whether the metric type is GaugeType.
func IsGauge(t MetricType) bool {
	return t == GaugeType
}

// IsCounter reports whether the metric type is CounterType.
func IsCounter(t MetricType) bool {
	return t == CounterType
}

// NewGaugeMetrics constructs a gauge metric with the provided name and value.
func NewGaugeMetrics(name string, value *float64) (*Metrics, error) {
	// if !IsGaugeName(name) {
	// 	if IsCounterName(name) {
	// 		return nil, ErrMetricNameTypeMismatch
	// 	}
	// 	return nil, fmt.Errorf("%w: %q", ErrMetricUnknownName, name)
	// }

	return &Metrics{
		ID:    name,
		MType: GaugeType,
		Value: value,
	}, nil
}

// NewCounterMetrics constructs a counter metric with the provided name and delta.
func NewCounterMetrics(name string, value *int64) (*Metrics, error) {
	// if !IsCounterName(name) {
	// 	if IsGaugeName(name) {
	// 		return nil, ErrMetricNameTypeMismatch
	// 	}
	// 	return nil, fmt.Errorf("%w: %q", ErrMetricUnknownName, name)
	// }

	return &Metrics{
		ID:    name,
		MType: CounterType,
		Delta: value,
	}, nil
}

// NewMetrics parses the string value according to the metric type and returns a Metrics instance.
func NewMetrics(name string, value string, t MetricType) (*Metrics, error) {
	if IsGauge(t) {
		var p *float64
		if value != "" {
			v, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return nil, fmt.Errorf("%w: parse float64 for %q: %v", ErrMetricInvalidValueType, name, err)
			}
			p = &v
		}
		return NewGaugeMetrics(name, p)
	}

	if IsCounter(t) {
		var p *int64
		if value != "" {
			v, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("%w: parse int64 for %q: %v", ErrMetricInvalidValueType, name, err)
			}
			p = &v
		}
		return NewCounterMetrics(name, p)
	}

	return nil, fmt.Errorf("%w: %q", ErrMetricUnknownName, name)
}

// MarshalJSON serialises the metric to JSON while avoiding nil pointer panics.
func (m *Metrics) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}

	// switch MetricType(m.MType) {
	// case GaugeType:
	// 	if !IsGaugeName(m.ID) {
	// 		return nil, fmt.Errorf("%w: %q", ErrMetricUnknownName, m.ID)
	// 	}
	// 	if m.Value == nil {
	// 		return nil, ErrMetricMissingValue
	// 	}
	// 	if m.Delta != nil {
	// 		return nil, ErrMetricAmbiguousValue
	// 	}
	// case CounterType:
	// 	if !IsCounterName(m.ID) {
	// 		return nil, fmt.Errorf("%w: %q", ErrMetricUnknownName, m.ID)
	// 	}
	// 	if m.Delta == nil {
	// 		return nil, ErrMetricMissingValue
	// 	}
	// 	if m.Value != nil {
	// 		return nil, ErrMetricAmbiguousValue
	// 	}
	// default:
	// 	return nil, ErrMetricInvalidType
	// }

	type alias Metrics
	return json.Marshal(alias(*m))
}

// UnmarshalJSON deserialises metric JSON payloads into the Metrics struct.
func (m *Metrics) UnmarshalJSON(data []byte) error {
	type alias struct {
		ID    string     `json:"id"`
		MType MetricType `json:"type"`
		Delta *int64     `json:"delta,omitempty"`
		Value *float64   `json:"value,omitempty"`
		Hash  string     `json:"hash,omitempty"`
	}

	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}

	// switch MetricType(a.MType) {
	// case GaugeType:
	// 	if a.Value == nil {
	// 		return ErrMetricMissingValue
	// 	}
	// 	if a.Delta != nil {
	// 		return ErrMetricAmbiguousValue
	// 	}
	// 	if !IsGauge(a.MType) {
	// 		if IsCounter(a.MType) {
	// 			return ErrMetricNameTypeMismatch
	// 		}
	// 		return fmt.Errorf("%w: %q", ErrMetricInvalidType, a.ID)
	// 	}
	// case CounterType:
	// 	if a.Delta == nil {
	// 		return ErrMetricMissingValue
	// 	}
	// 	if a.Value != nil {
	// 		return ErrMetricAmbiguousValue
	// 	}
	// 	if !IsCounter(a.MType) {
	// 		if IsGauge(a.MType) {
	// 			return ErrMetricNameTypeMismatch
	// 		}
	// 		return fmt.Errorf("%w: %q", ErrMetricInvalidType, a.ID)
	// 	}
	// default:
	// 	return ErrMetricInvalidType
	// }

	m.ID = a.ID
	m.MType = a.MType
	m.Delta = a.Delta
	m.Value = a.Value
	return nil
}
