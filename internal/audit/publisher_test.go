package audit

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	test "github.com/polkiloo/go-musthave-metrics-tppl/internal/test"
)

func TestNewPublisher(t *testing.T) {
	pub, err := NewPublisher(PublisherParams{Config: Config{}})
	if err != nil || pub != nil {
		t.Fatalf("expected nil publisher, got %v %v", pub, err)
	}

	_, err = NewPublisher(PublisherParams{Config: Config{Endpoint: "://bad"}})
	if err == nil {
		t.Fatalf("expected error for invalid url")
	}

	pub, err = NewPublisher(PublisherParams{Config: Config{FilePath: filepath.Join(t.TempDir(), "audit.log"), Endpoint: "http://x"}, Client: test.RoundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(nil))}, nil
	})})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := pub.Publish(context.Background(), Event{}); err != nil {
		t.Fatalf("publish: %v", err)
	}
}

func TestAddRequestMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)
	AddRequestMetrics(c, "a")
	AddRequestMetrics(c, "", "b")
	metrics := takeRequestMetrics(c)
	if !reflect.DeepEqual(metrics, []string{"a", "b"}) {
		t.Fatalf("unexpected metrics: %v", metrics)
	}
	metricsPool.Put(metrics[:0])

	if m := takeRequestMetrics(c); m != nil {
		t.Fatalf("expected nil metrics, got %v", m)
	}
}

func TestMiddlewarePublishesEvent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fake := &test.FakePublisher[Event]{}
	clock := clockFunc(func() time.Time { return time.Unix(123, 0) })
	r := gin.New()
	r.Use(Middleware(fake, nil, clock))
	r.GET("/", func(c *gin.Context) {
		AddRequestMetrics(c, "m1", "")
		AddRequestMetrics(c, "m2")
		c.Status(http.StatusNoContent)
	})
	resp := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.0.1:1234"
	r.ServeHTTP(resp, req)

	events := fake.GetEvents()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	got := events[0]
	if got.Timestamp != 123 || got.IPAddress == "" || !reflect.DeepEqual(got.Metrics, []string{"m1", "m2"}) {
		t.Fatalf("unexpected event %+v", got)
	}
}

func TestMiddlewareHandlesErrorsAndNilPublisher(t *testing.T) {
	gin.SetMode(gin.TestMode)
	if mw := Middleware(nil, nil, nil); mw == nil {
		t.Fatalf("expected middleware")
	}

	fake := &test.FakePublisher[Event]{Err: errors.New("boom")}
	log := &test.FakeLogger{}
	r := gin.New()
	r.Use(Middleware(fake, log, clockFunc(func() time.Time { return time.Unix(0, 0) })))
	r.GET("/", func(c *gin.Context) {
		AddRequestMetrics(c, "")
		AddRequestMetrics(c)
		c.Status(http.StatusOK)
	})
	r.POST("/with", func(c *gin.Context) {
		AddRequestMetrics(c, "metric")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(httptest.NewRecorder(), req)

	req = httptest.NewRequest(http.MethodPost, "/with", nil)
	type ctxKey string
	const fakeKey ctxKey = "fake"
	req = req.WithContext(context.WithValue(context.Background(), fakeKey, 1))
	req.RemoteAddr = "10.0.0.1:55"
	r.ServeHTTP(httptest.NewRecorder(), req)

	if len(log.GetErrorMessages()) == 0 {
		t.Fatalf("expected logged error")
	}
}
