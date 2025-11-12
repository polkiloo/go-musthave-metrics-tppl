package handler

import (
	"sync"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/models"
)

const (
	defaultBatchCapacity = 64
)

type jsonMetricsPool struct {
	metricPool sync.Pool
	batchPool  sync.Pool
}

// NewJSONMetricsPool constructs a reusable pool for single metrics and batches consumed by JSON handlers.
func NewJSONMetricsPool() *jsonMetricsPool {
	pool := &jsonMetricsPool{}
	pool.metricPool.New = func() any {
		return &models.Metrics{}
	}
	pool.batchPool.New = func() any {
		batch := make([]models.Metrics, 0, defaultBatchCapacity)
		return &batch
	}
	return pool
}

func (p *jsonMetricsPool) AcquireMetric() *models.Metrics {
	metric := p.metricPool.Get().(*models.Metrics)
	*metric = models.Metrics{}
	return metric
}

func (p *jsonMetricsPool) ReleaseMetric(metric *models.Metrics) {
	if metric == nil {
		return
	}
	*metric = models.Metrics{}
	p.metricPool.Put(metric)
}

func (p *jsonMetricsPool) AcquireBatch() *[]models.Metrics {
	batchPtr := p.batchPool.Get().(*[]models.Metrics)
	batch := *batchPtr
	if cap(batch) == 0 {
		batch = make([]models.Metrics, 0, defaultBatchCapacity)
	}
	*batchPtr = batch[:0]
	return batchPtr
}

func (p *jsonMetricsPool) ReleaseBatch(batchPtr *[]models.Metrics) {
	if batchPtr == nil {
		return
	}
	batch := *batchPtr
	for i := range batch {
		batch[i] = models.Metrics{}
	}
	*batchPtr = batch[:0]
	p.batchPool.Put(batchPtr)
}
