package proto

import (
	"fmt"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/models"
)

// MetricFromModel converts domain metric to proto message.
func MetricFromModel(m *models.Metrics) (*Metric, error) {
	if m == nil {
		return nil, fmt.Errorf("metric is nil")
	}
	pm := &Metric{Id: m.ID}
	switch m.MType {
	case models.GaugeType:
		pm.Type = Metric_GAUGE
		if m.Value != nil {
			pm.Value = *m.Value
		}
	case models.CounterType:
		pm.Type = Metric_COUNTER
		if m.Delta != nil {
			pm.Delta = *m.Delta
		}
	default:
		return nil, fmt.Errorf("unsupported metric type: %s", m.MType)
	}
	return pm, nil
}

// MetricToModel converts proto metric to domain model.
func MetricToModel(m *Metric) (*models.Metrics, error) {
	if m == nil {
		return nil, fmt.Errorf("metric is nil")
	}
	switch m.Type {
	case Metric_GAUGE:
		v := m.Value
		return models.NewGaugeMetrics(m.Id, &v)
	case Metric_COUNTER:
		d := m.Delta
		return models.NewCounterMetrics(m.Id, &d)
	default:
		return nil, fmt.Errorf("unsupported metric type: %v", m.Type)
	}
}

// MetricsFromModels converts slice of domain metrics to proto messages.
func MetricsFromModels(metrics []*models.Metrics) ([]*Metric, error) {
	res := make([]*Metric, 0, len(metrics))
	for _, m := range metrics {
		pm, err := MetricFromModel(m)
		if err != nil {
			return nil, err
		}
		res = append(res, pm)
	}
	return res, nil
}

// MetricsToModels converts slice of proto metrics to domain models.
func MetricsToModels(metrics []*Metric) ([]models.Metrics, error) {
	res := make([]models.Metrics, 0, len(metrics))
	for _, m := range metrics {
		model, err := MetricToModel(m)
		if err != nil {
			return nil, err
		}
		res = append(res, *model)
	}
	return res, nil
}
