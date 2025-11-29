package audit

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/logger"
	"go.uber.org/fx"
)

// Publisher publishes audit events to downstream sinks.
type Publisher interface {
	Publish(context.Context, Event) error
}

// Config describes audit outputs configured for the service.
type Config struct {
	FilePath string
	Endpoint string
}

// Clock abstracts time retrieval to simplify testing.
type Clock interface {
	Now() time.Time
}

type clockFunc func() time.Time

func (f clockFunc) Now() time.Time { return f() }

var systemClock Clock = clockFunc(time.Now)

type PublisherParams struct {
	fx.In
	Config Config
	Client HTTPClient `optional:"true"`
}

// NewPublisher constructs a Publisher from configuration.
func NewPublisher(p PublisherParams) (Publisher, error) {
	observers := make([]Observer, 0, 2)
	if p.Config.FilePath != "" {
		observers = append(observers, NewFileObserver(p.Config.FilePath))
	}
	if p.Config.Endpoint != "" {
		client := p.Client
		if client == nil {
			client = http.DefaultClient
		}
		httpObserver, err := NewHTTPObserver(p.Config.Endpoint, client)
		if err != nil {
			return nil, err
		}
		observers = append(observers, httpObserver)
	}
	if len(observers) == 0 {
		return nil, nil
	}
	return NewDispatcher(observers...), nil
}

const contextMetricsKey = "audit.metrics"

var metricsPool = sync.Pool{
	New: func() any { return make([]string, 0, 4) },
}

// AddRequestMetrics records metric identifiers for the current Gin request.
func AddRequestMetrics(c *gin.Context, metrics ...string) {
	if c == nil || len(metrics) == 0 {
		return
	}
	stored, exists := c.Get(contextMetricsKey)
	var collected []string
	if !exists || stored == nil {
		collected = metricsPool.Get().([]string)
		collected = collected[:0]
	} else {
		collected = stored.([]string)
	}
	for _, m := range metrics {
		if m == "" {
			continue
		}
		collected = append(collected, m)
	}
	c.Set(contextMetricsKey, collected)
}

func takeRequestMetrics(c *gin.Context) []string {
	if c == nil {
		return nil
	}
	stored, exists := c.Get(contextMetricsKey)
	if !exists || stored == nil {
		return nil
	}
	c.Set(contextMetricsKey, nil)
	return stored.([]string)
}

// GetRequestMetricsForTest exposes collected metrics for unit tests.
func GetRequestMetricsForTest(c *gin.Context) []string {
	collected := takeRequestMetrics(c)
	if collected == nil {
		return nil
	}
	metrics := append([]string(nil), collected...)
	metricsPool.Put(collected[:0])
	return metrics
}

// Middleware publishes audit events after the request is processed.
func Middleware(pub Publisher, l logger.Logger, clock Clock) gin.HandlerFunc {
	if pub == nil {
		return func(c *gin.Context) { c.Next() }
	}
	if clock == nil {
		clock = systemClock
	}
	return func(c *gin.Context) {
		c.Next()
		collected := takeRequestMetrics(c)
		if len(collected) == 0 {
			if collected != nil {
				metricsPool.Put(collected[:0])
			}
			return
		}
		metrics := append([]string(nil), collected...)
		ip := ""
		if c != nil {
			ip = c.ClientIP()
		}
		eventCtx := context.Background()
		if c != nil && c.Request != nil {
			eventCtx = c.Request.Context()
		}
		event := Event{
			Timestamp: clock.Now().Unix(),
			Metrics:   metrics,
			IPAddress: ip,
		}
		if err := pub.Publish(eventCtx, event); err != nil && l != nil {
			l.WriteError("audit publish failed", "error", err)
		}
		metricsPool.Put(collected[:0])
	}
}
