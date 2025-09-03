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

func TestUpdatesJSON_UnsupportedMediaType(t *testing.T) {
	fs := &test.FakeMetricService{}
	r := setupRouterWithUpdatesJSON(fs)

	w := test.DoJSON(r, "/updates", []models.Metrics{}, "text/plain")
	if w.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnsupportedMediaType)
	}
}

func TestUpdatesJSON_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fs := &test.FakeMetricService{}
	r := setupRouterWithUpdatesJSON(fs)

	req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestUpdatesJSON_MissingIDOrType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fs := &test.FakeMetricService{}
	r := setupRouterWithUpdatesJSON(fs)

	body := []map[string]any{{"type": "gauge", "value": 1.0}}
	w := test.DoJSON(r, "/updates", body, "application/json")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d (missing id)", w.Code, http.StatusBadRequest)
	}

	body = []map[string]any{{"id": "Alloc", "value": 1.0}}
	w = test.DoJSON(r, "/updates", body, "application/json")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d (missing type)", w.Code, http.StatusBadRequest)
	}
}

func TestUpdatesJSON_Success(t *testing.T) {
	val := 3.14
	metricName := models.GaugeNames[0]
	out := &models.Metrics{ID: metricName, MType: models.GaugeType, Value: &val}

	gin.SetMode(gin.TestMode)
	fs := &test.FakeMetricService{Metric: *out}
	r := setupRouterWithUpdatesJSON(fs)

	m, _ := models.NewGaugeMetrics(metricName, &val)
	w := test.DoJSON(r, "/updates", []models.Metrics{*m}, "application/json")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type = %q, want %q", ct, "application/json")
	}
}

func setupRouterWithUpdatesJSON(srvc *test.FakeMetricService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := &GinHandler{service: srvc}
	r := gin.New()
	r.POST("/updates", h.UpdatesJSON)
	r.POST("/updates/", h.UpdatesJSON)
	return r
}
