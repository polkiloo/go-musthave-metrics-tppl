package service

import (
	"fmt"
	"strconv"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/storage"
)

type MetricService struct {
	store storage.MetricStorage
}

func NewMetricService() *MetricService {
	return &MetricService{store: storage.NewMemStorage()}
}

func (s *MetricService) ProcessUpdate(metricType, name, rawValue string) error {
	switch metricType {
	case "gauge":
		v, err := strconv.ParseFloat(rawValue, 64)
		if err != nil {
			return fmt.Errorf("invalid gauge value: %w", err)
		}
		s.store.UpdateGauge(name, v)
	case "counter":
		d, err := strconv.ParseInt(rawValue, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid counter value: %w", err)
		}
		s.store.UpdateCounter(name, d)
	default:
		return fmt.Errorf("unknown metric type %q", metricType)
	}
	return nil
}
