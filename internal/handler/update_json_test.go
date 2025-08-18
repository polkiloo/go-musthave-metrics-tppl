package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/models"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/test"
)

func TestUpdateJSON_UnsupportedMediaType(t *testing.T) {
	fs := &test.FakeMetricService{}
	r := setupRouterWithUpdateJSON(fs)

	w := test.DoJSON(r, "/update", map[string]any{}, "text/plain")
	if w.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnsupportedMediaType)
	}
}

func TestUpdateJSON_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fs := &test.FakeMetricService{}
	r := setupRouterWithUpdateJSON(fs)

	req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestUpdateJSON_MissingIDOrType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fs := &test.FakeMetricService{}
	r := setupRouterWithUpdateJSON(fs)

	w := test.DoJSON(r, "/update/", map[string]any{"type": "gauge", "value": 1.0}, "application/json")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d (missing id)", w.Code, http.StatusBadRequest)
	}

	w = test.DoJSON(r, "/update", map[string]any{"id": "Alloc", "value": 1.0}, "application/json")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d (missing type)", w.Code, http.StatusBadRequest)
	}
}

func TestUpdateJSON_UnknownType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fs := &test.FakeMetricService{}
	r := setupRouterWithUpdateJSON(fs)

	w := test.DoJSON(r, "/update", map[string]any{"id": "Alloc", "type": "weird", "value": 1.0}, "application/json")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d (unknown type)", w.Code, http.StatusBadRequest)
	}
}

func TestUpdateJSON_Success_Gauge(t *testing.T) {
	val := 3.14
	metricName := models.GaugeNames[0]
	out := &models.Metrics{ID: metricName, MType: models.GaugeType, Value: &val}

	gin.SetMode(gin.TestMode)
	fs := &test.FakeMetricService{Metric: *out}
	r := setupRouterWithUpdateJSON(fs)

	m, _ := models.NewGaugeMetrics(metricName, &val)
	w := test.DoJSON(r, "/update", m, "application/json")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type = %q, want %q", ct, "application/json")
	}

}

func TestUpdateJSON_Success_Counter(t *testing.T) {
	d := int64(42)
	metricName := models.GaugeNames[0]
	out, _ := models.NewCounterMetrics(metricName, &d)

	gin.SetMode(gin.TestMode)
	fs := &test.FakeMetricService{Metric: *out}
	r := setupRouterWithUpdateJSON(fs)

	m, _ := models.NewCounterMetrics(metricName, &d)
	w := test.DoJSON(r, "/update", m, "application/json")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func setupRouterWithUpdateJSON(srvc *test.FakeMetricService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := &GinHandler{service: srvc}
	r := gin.New()
	r.POST("/update", h.UpdateJSON)
	r.POST("/update/", h.UpdateJSON)
	return r
}
