package service

import (
	"fmt"
	"strconv"

	models "github.com/polkiloo/go-musthave-metrics-tppl/internal/model"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/storage"
)

var (
	ErrUnknownMetricType = fmt.Errorf("unknown metric type")
	ErrMetricNotFound    = fmt.Errorf("metric not found")
)

type MetricServiceInterface interface {
	ProcessUpdate(metricType models.MetricType, name, rawValue string) error
	ProcessGetValue(metricType models.MetricType, name string) (string, error)
}

type MetricService struct {
	store storage.MetricStorage
}

func NewMetricService() *MetricService {
	return &MetricService{store: storage.NewMemStorage()}
}

func (s *MetricService) ProcessUpdate(metricType models.MetricType, name, rawValue string) error {
	switch metricType {
	case models.GaugeType:
		v, err := strconv.ParseFloat(rawValue, 64)
		if err != nil {
			return fmt.Errorf("invalid gauge value: %w", err)
		}
		s.store.UpdateGauge(name, v)
	case models.CounterType:
		d, err := strconv.ParseInt(rawValue, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid counter value: %w", err)
		}
		s.store.UpdateCounter(name, d)
	default:
		return ErrUnknownMetricType
	}
	return nil
}

func (s *MetricService) ProcessGetValue(metricType models.MetricType, name string) (string, error) {
	switch metricType {
	case models.GaugeType:
		v, err := s.store.GetGauge(name)
		if err == storage.ErrMetricNotFound {
			return "", ErrMetricNotFound
		}
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	case models.CounterType:
		v, err := s.store.GetCounter(name)
		if err == storage.ErrMetricNotFound {
			return "", ErrMetricNotFound
		}
		return strconv.FormatInt(v, 10), nil
	default:
		return "", ErrUnknownMetricType
	}
}

var _ MetricServiceInterface = NewMetricService()
