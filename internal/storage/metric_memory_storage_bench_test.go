package storage

import (
	"testing"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/models"
)

func buildMetricsSet() []models.Metrics {
	gauges := make([]models.Metrics, len(models.GaugeNames))
	for i, name := range models.GaugeNames {
		value := float64(i)
		gauges[i] = models.Metrics{
			ID:    name,
			MType: models.GaugeType,
			Value: &value,
		}
	}

	counters := make([]models.Metrics, len(models.CounterNames))
	for i, name := range models.CounterNames {
		value := int64(i)
		counters[i] = models.Metrics{
			ID:    name,
			MType: models.CounterType,
			Delta: &value,
		}
	}

	metrics := append(gauges, counters...)
	return metrics
}

func BenchmarkMemStorageUpdateBatch(b *testing.B) {
	storage := NewMemStorage()
	metrics := buildMetricsSet()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := storage.UpdateBatch(metrics); err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

func BenchmarkMemStorageSnapshot(b *testing.B) {
	storage := NewMemStorage()
	metrics := buildMetricsSet()
	if err := storage.UpdateBatch(metrics); err != nil {
		b.Fatalf("unexpected error: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = storage.Snapshot()
	}
}
