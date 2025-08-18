package service

import (
	"fmt"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/models"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/storage"
)

var (
	ErrMetricNotFound = fmt.Errorf("metric not found")
)

type MetricServiceInterface interface {
	ProcessUpdate(*models.Metrics) error
	ProcessGetValue(name string, metricType models.MetricType) (*models.Metrics, error)
}

type MetricService struct {
	store storage.MetricStorage
}

func NewMetricService() *MetricService {
	return &MetricService{store: storage.NewMemStorage()}
}

func (s *MetricService) ProcessUpdate(m *models.Metrics) error {
	if m == nil {
		return ErrMetricNotFound
	}

	switch m.MType {
	case models.GaugeType:
		s.store.UpdateGauge(m.ID, *m.Value)
	case models.CounterType:
		s.store.UpdateCounter(m.ID, *m.Delta)
	}
	return nil
}

func (s *MetricService) ProcessGetValue(metricName string, metricType models.MetricType) (*models.Metrics, error) {
	var m *models.Metrics

	switch {
	case models.IsGauge(metricType):
		v, err := s.store.GetGauge(metricName)
		if err == storage.ErrMetricNotFound {
			return nil, ErrMetricNotFound
		}
		if err != nil {
			return nil, err
		}
		m, err = models.NewGaugeMetrics(metricName, &v)
		if err != nil {
			return nil, err
		}

	case models.IsCounter(metricType):
		v, err := s.store.GetCounter(metricName)
		if err == storage.ErrMetricNotFound {
			return nil, ErrMetricNotFound
		}
		if err != nil {
			return nil, err
		}
		m, err = models.NewCounterMetrics(metricName, &v)
		if err != nil {
			return nil, err
		}

	default:
		return nil, models.ErrMetricUnknownName
	}

	return m, nil
}

var _ MetricServiceInterface = NewMetricService()
