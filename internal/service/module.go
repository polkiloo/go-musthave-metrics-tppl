package service

import (
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/db"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/storage"
	"go.uber.org/fx"
)

func provideStorage(pool db.Pool) storage.MetricStorage {
	if pool != nil {
		return storage.NewDBStorage(pool)
	}
	return storage.NewMemStorage()
}

func newMetricService(st storage.MetricStorage) MetricServiceInterface {
	return NewMetricService(st)
}

var Module = fx.Module(
	"metric-service",
	fx.Provide(
		provideStorage,
		newMetricService,
	),
)
