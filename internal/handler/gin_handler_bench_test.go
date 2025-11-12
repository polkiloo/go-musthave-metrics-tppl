package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/models"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/service"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/storage"
)

func setupBenchmarkServer(b *testing.B) (*httptest.Server, *http.Client) {
	b.Helper()

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	store := storage.NewMemStorage()
	svc := service.NewMetricService(store)
	handler := NewGinHandler(svc, NewJSONMetricsPool())
	handler.RegisterUpdate(engine)

	server := httptest.NewServer(engine)
	client := server.Client()
	client.Timeout = time.Second

	b.Cleanup(func() {
		server.Close()
	})

	return server, client
}

func BenchmarkGinHandlerUpdateJSONNetwork(b *testing.B) {
	server, client := setupBenchmarkServer(b)

	value := 123.456
	metric := models.Metrics{
		ID:    "Alloc",
		MType: models.GaugeType,
		Value: &value,
	}

	payload, err := json.Marshal(metric)
	if err != nil {
		b.Fatalf("failed to marshal metric: %v", err)
	}

	endpoint, err := url.JoinPath(server.URL, "update")
	if err != nil {
		b.Fatalf("failed to build request URL: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(payload))
		if err != nil {
			b.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			b.Fatalf("failed to send request: %v", err)
		}

		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			b.Fatalf("unexpected status: %s", resp.Status)
		}
	}
}

func BenchmarkGinHandlerUpdatesJSONNetwork(b *testing.B) {
	server, client := setupBenchmarkServer(b)

	metrics := make([]models.Metrics, 0, len(models.GaugeNames)+len(models.CounterNames))
	for i, name := range models.GaugeNames {
		value := float64(i)
		metrics = append(metrics, models.Metrics{
			ID:    name,
			MType: models.GaugeType,
			Value: &value,
		})
	}

	for i, name := range models.CounterNames {
		delta := int64(i)
		metrics = append(metrics, models.Metrics{
			ID:    name,
			MType: models.CounterType,
			Delta: &delta,
		})
	}

	payload, err := json.Marshal(metrics)
	if err != nil {
		b.Fatalf("failed to marshal metrics: %v", err)
	}

	endpoint, err := url.JoinPath(server.URL, "updates")
	if err != nil {
		b.Fatalf("failed to build request URL: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(payload))
		if err != nil {
			b.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			b.Fatalf("failed to send request: %v", err)
		}

		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			b.Fatalf("unexpected status: %s", resp.Status)
		}
	}
}
