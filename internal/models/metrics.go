package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

type (
	Gauge   float64
	Counter int64
)

type MetricType string

const (
	GaugeType   MetricType = "gauge"
	CounterType MetricType = "counter"
)

var MetricTypes = []MetricType{
	GaugeType,
	CounterType,
}

func (t MetricType) IsValid() bool {
	switch t {
	case GaugeType, CounterType:
		return true
	default:
		return false
	}
}

func ParseMetricType(s string) (MetricType, error) {
	mt := MetricType(s)
	if mt.IsValid() {
		return mt, nil
	}
	return "", ErrMetricInvalidType
}

var GaugeNames = []string{
	"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys",
	"HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased",
	"HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys",
	"MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC",
	"NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys",
	"Sys", "TotalAlloc", "RandomValue",
}

var CounterNames = []string{
	"PollCount",
}
var (
	GaugeSet   = make(map[string]struct{}, len(GaugeNames))
	CounterSet = make(map[string]struct{}, len(CounterNames))
)

var (
	ErrMetricUnknownName      = errors.New("unknown metric name")
	ErrMetricInvalidType      = errors.New("invalid metric type")
	ErrMetricInvalidValueType = errors.New("invalid value type for metric")
	ErrMetricMissingValue     = errors.New("missing value for metric")
	ErrMetricAmbiguousValue   = errors.New("both gauge and counter values are set")
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

// NOTE: Не усложняем пример, вводя иерархическую вложенность структур.
// Органичиваясь плоской моделью.
// Delta и Value объявлены через указатели,
// что бы отличать значение "0", от не заданного значения
// и соответственно не кодировать в структуру.
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
func IsGauge(t MetricType) bool {
	return t == GaugeType
}
func IsCounter(t MetricType) bool {
	return t == CounterType
}

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
