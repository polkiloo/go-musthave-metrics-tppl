package proto

import (
	"testing"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/models"
)

func TestMetricFromModel(t *testing.T) {
	t.Parallel()

	gaugeVal := 10.5
	counterVal := int64(7)

	tests := []struct {
		name    string
		metric  *models.Metrics
		wantErr bool
	}{
		{name: "nil metric", metric: nil, wantErr: true},
		{name: "gauge metric", metric: &models.Metrics{ID: "g1", MType: models.GaugeType, Value: &gaugeVal}},
		{name: "counter metric", metric: &models.Metrics{ID: "c1", MType: models.CounterType, Delta: &counterVal}},
		{name: "unsupported type", metric: &models.Metrics{MType: "unknown"}, wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			res, err := MetricFromModel(tt.metric)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if res.GetId() == "" {
				t.Fatalf("id not copied")
			}
		})
	}
}

func TestMetricToModel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		metric  *Metric
		wantErr bool
	}{
		{name: "nil metric", metric: nil, wantErr: true},
		{name: "gauge metric", metric: &Metric{Id: "g", Type: Metric_GAUGE, Value: 3.14}},
		{name: "counter metric", metric: &Metric{Id: "c", Type: Metric_COUNTER, Delta: 3}},
		{name: "unsupported type", metric: &Metric{Type: -1}, wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			res, err := MetricToModel(tt.metric)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if res == nil || res.ID == "" {
				t.Fatalf("result not populated")
			}
		})
	}
}

func TestMetricsFromModels(t *testing.T) {
	t.Parallel()

	if _, err := MetricsFromModels([]*models.Metrics{{MType: "bad"}}); err == nil {
		t.Fatalf("expected error on unsupported type")
	}

	gaugeVal := 1.0
	in := []*models.Metrics{{ID: "g", MType: models.GaugeType, Value: &gaugeVal}}
	res, err := MetricsFromModels(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res) != len(in) {
		t.Fatalf("expected %d results, got %d", len(in), len(res))
	}
}

func TestMetricsToModels(t *testing.T) {
	t.Parallel()

	if _, err := MetricsToModels([]*Metric{{Type: Metric_MType(-1)}}); err == nil {
		t.Fatalf("expected error on unsupported type")
	}

	res, err := MetricsToModels([]*Metric{{Id: "c", Type: Metric_COUNTER, Delta: 5}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res) != 1 || res[0].ID != "c" {
		t.Fatalf("conversion failed")
	}
}

func TestMetricsToModelsNilSafety(t *testing.T) {
	t.Parallel()

	_, err := MetricsToModels([]*Metric{nil})
	if err == nil {
		t.Fatalf("expected non-nil error for nil metric")
	}
}
