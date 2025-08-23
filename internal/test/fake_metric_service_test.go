package test

import "github.com/polkiloo/go-musthave-metrics-tppl/internal/service"

var _ service.MetricServiceInterface = &FakeMetricService{}
