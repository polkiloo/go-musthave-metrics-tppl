package service

import (
	models "github.com/polkiloo/go-musthave-metrics-tppl/internal/model"
)

type FakeMetricService struct {
	Err    error
	MType  models.MetricType
	MName  string
	MValue string
}

func (f *FakeMetricService) ProcessUpdate(metricType models.MetricType, name, rawValue string) error {
	f.MType = metricType
	f.MName = name
	f.MValue = rawValue
	return f.Err
}

func (f *FakeMetricService) ProcessGetValue(metricType models.MetricType, name string) (string, error) {
	f.MType = metricType
	f.MName = name
	return f.MValue, f.Err
}

var _ MetricServiceInterface = &FakeMetricService{}
