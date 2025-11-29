package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/audit"
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

func TestUpdateJSON_AfterUpdateHook(t *testing.T) {
	val := 1.0
	gin.SetMode(gin.TestMode)
	fs := &test.FakeMetricService{}
	h := &GinHandler{service: fs}
	called := 0
	h.SetAfterUpdateHook(func() { called++ })
	r := gin.New()
	r.POST("/update", h.UpdateJSON)

	m, _ := models.NewGaugeMetrics("g1", &val)
	w := test.DoJSON(r, "/update", m, "application/json")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if called != 1 {
		t.Fatalf("afterUpdate not called, got %d", called)
	}
}

func TestUpdateJSON_MetricsStoredInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fs := &test.FakeMetricService{}
	h := &GinHandler{service: fs}

	value := 2.0
	m, _ := models.NewGaugeMetrics("Alloc", &value)
	body, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	h.UpdateJSON(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	metrics := audit.GetRequestMetricsForTest(ctx)
	if len(metrics) != 1 || metrics[0] != "Alloc" {
		t.Fatalf("unexpected metrics: %v", metrics)
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
