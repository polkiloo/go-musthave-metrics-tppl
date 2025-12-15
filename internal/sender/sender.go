package sender

import (
	"context"
	"net/http"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/models"
)

// SenderInterface describes how the agent sends metrics to the server.
type SenderInterface interface {
	Send(metrics []*models.Metrics)
	SendBatch(metrics []*models.Metrics)
}

// RequestMiddleware decorates outgoing HTTP requests before they are executed.
type RequestMiddleware func(*http.Request)

// ContextualSender extends SenderInterface with context-aware sending.
type ContextualSender interface {
	SenderInterface
	SendWithContext(ctx context.Context, metrics []*models.Metrics)
}
