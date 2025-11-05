package sender

import (
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/models"
)

// SenderInterface describes how the agent sends metrics to the server.
type SenderInterface interface {
	Send(metrics []*models.Metrics)
	SendBatch(metrics []*models.Metrics)
}
