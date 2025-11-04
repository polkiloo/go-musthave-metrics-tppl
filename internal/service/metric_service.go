package service

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/models"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/storage"
)

var (
	ErrMetricNotFound = fmt.Errorf("metric not found")
)

type MetricServiceInterface interface {
	ProcessUpdate(*models.Metrics) error
	ProcessUpdates([]models.Metrics) error
	ProcessGetValue(name string, metricType models.MetricType) (*models.Metrics, error)
	SaveFile(path string) error
	LoadFile(path string) error
}

type MetricService struct {
	store storage.MetricStorage
}

func NewMetricService(store storage.MetricStorage) *MetricService {
	return &MetricService{store: store}
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

type batchUpdater interface {
	UpdateBatch([]models.Metrics) error
}

var processUpdateFn = (*MetricService).ProcessUpdate

func (s *MetricService) ProcessUpdates(metrics []models.Metrics) error {
	if len(metrics) == 0 {
		return nil
	}
	if bu, ok := s.store.(batchUpdater); ok {
		return bu.UpdateBatch(metrics)
	}
	for i := 0; i < len(metrics); i++ {
		if err := processUpdateFn(s, &metrics[i]); err != nil {
			return err
		}
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

func (s *MetricService) SaveFile(path string) error {
	if path == "" {
		return nil
	}
	metrics := make([]models.Metrics, 0)
	if ms, ok := s.store.(*storage.MemStorage); ok {
		metrics = append(metrics, ms.Snapshot()...)
	}
	b, err := json.Marshal(metrics)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o666)
}

func (s *MetricService) LoadFile(path string) error {
	if path == "" {
		return nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var metrics []models.Metrics
	if err := json.Unmarshal(b, &metrics); err != nil {
		return err
	}
	if ms, ok := s.store.(*storage.MemStorage); ok {
		for _, m := range metrics {
			switch m.MType {
			case models.GaugeType:
				if m.Value != nil {
					ms.UpdateGauge(m.ID, *m.Value)
				}
			case models.CounterType:
				if m.Delta != nil {
					ms.SetCounter(m.ID, *m.Delta)
				}
			}
		}
	}
	return nil
}

var _ MetricServiceInterface = NewMetricService(nil)
