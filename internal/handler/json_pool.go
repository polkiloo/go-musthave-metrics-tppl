package handler

import (
	"sync"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/models"
)

const (
	defaultBatchCapacity = 64
)

var metricPool = sync.Pool{
	New: func() any {
		return &models.Metrics{}
	},
}

var metricsBatchPool = sync.Pool{
	New: func() any {
		batch := make([]models.Metrics, 0, defaultBatchCapacity)
		return &batch
	},
}

func acquireMetric() *models.Metrics {
	metric := metricPool.Get().(*models.Metrics)
	*metric = models.Metrics{}
	return metric
}

func releaseMetric(metric *models.Metrics) {
	if metric == nil {
		return
	}
	*metric = models.Metrics{}
	metricPool.Put(metric)
}

func acquireMetricsBatch() *[]models.Metrics {
	batchPtr := metricsBatchPool.Get().(*[]models.Metrics)
	batch := *batchPtr
	if cap(batch) == 0 {
		batch = make([]models.Metrics, 0, defaultBatchCapacity)
	}
	batch = batch[:0]
	*batchPtr = batch
	return batchPtr
}

func releaseMetricsBatch(batchPtr *[]models.Metrics) {
	if batchPtr == nil {
		return
	}
	batch := *batchPtr
	for i := range batch {
		batch[i] = models.Metrics{}
	}
	*batchPtr = batch[:0]
	metricsBatchPool.Put(batchPtr)
}
