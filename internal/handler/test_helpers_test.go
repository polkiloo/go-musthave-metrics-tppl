package handler

import "github.com/polkiloo/go-musthave-metrics-tppl/internal/service"

func newTestGinHandler(s service.MetricServiceInterface) *GinHandler {
	return NewGinHandler(s, NewJSONMetricsPool())
}
