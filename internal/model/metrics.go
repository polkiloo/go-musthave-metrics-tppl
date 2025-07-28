package models

type (
	Gauge   float64
	Counter int64
)

type MetricType string

const (
	GaugeType   MetricType = "gauge"
	CounterType MetricType = "counter"
)

type Metric struct {
	Name string
	Type MetricType
}

var RuntimeMetrics = []Metric{
	{Name: "Alloc", Type: GaugeType},
	{Name: "BuckHashSys", Type: GaugeType},
	{Name: "Frees", Type: GaugeType},
	{Name: "GCCPUFraction", Type: GaugeType},
	{Name: "GCSys", Type: GaugeType},
	{Name: "HeapAlloc", Type: GaugeType},
	{Name: "HeapIdle", Type: GaugeType},
	{Name: "HeapInuse", Type: GaugeType},
	{Name: "HeapObjects", Type: GaugeType},
	{Name: "HeapReleased", Type: GaugeType},
	{Name: "HeapSys", Type: GaugeType},
	{Name: "LastGC", Type: GaugeType},
	{Name: "Lookups", Type: GaugeType},
	{Name: "MCacheInuse", Type: GaugeType},
	{Name: "MCacheSys", Type: GaugeType},
	{Name: "MSpanInuse", Type: GaugeType},
	{Name: "MSpanSys", Type: GaugeType},
	{Name: "Mallocs", Type: GaugeType},
	{Name: "NextGC", Type: GaugeType},
	{Name: "NumForcedGC", Type: GaugeType},
	{Name: "NumGC", Type: GaugeType},
	{Name: "OtherSys", Type: GaugeType},
	{Name: "PauseTotalNs", Type: GaugeType},
	{Name: "StackInuse", Type: GaugeType},
	{Name: "StackSys", Type: GaugeType},
	{Name: "Sys", Type: GaugeType},
	{Name: "TotalAlloc", Type: GaugeType},
	{Name: "RandomValue", Type: GaugeType},
	{Name: "PollCount", Type: CounterType},
}

// NOTE: Не усложняем пример, вводя иерархическую вложенность структур.
// Органичиваясь плоской моделью.
// Delta и Value объявлены через указатели,
// что бы отличать значение "0", от не заданного значения
// и соответственно не кодировать в структуру.
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}
