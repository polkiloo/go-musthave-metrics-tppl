package sender

import (
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/models"
)

type SenderInterface interface {
	Send(metrics []*models.Metrics)
	SendBatch(metrics []*models.Metrics)
}
