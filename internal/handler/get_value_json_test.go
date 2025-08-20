package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/test"
)

func TestGetValueJSON_UnsupportedMediaType(t *testing.T) {
	fs := &test.FakeMetricService{}
	r := setupRouterWithGetValueJSON(fs)

	w := test.DoJSON(r, "/value", map[string]any{"id": "Alloc"}, "text/plain")
	if w.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnsupportedMediaType)
	}
}

func TestGetValueJSON_BadJSON(t *testing.T) {
	fs := &test.FakeMetricService{}
	r := setupRouterWithGetValueJSON(fs)

	req := httptest.NewRequest(http.MethodPost, "/value", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestGetValueJSON_MissingID(t *testing.T) {
	fs := &test.FakeMetricService{}
	r := setupRouterWithGetValueJSON(fs)

	w := test.DoJSON(r, "/value", map[string]any{"type": "gauge"}, "application/json")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func setupRouterWithGetValueJSON(srvc *test.FakeMetricService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := &GinHandler{service: srvc}
	r := gin.New()
	r.POST("/value", h.GetValueJSON)
	return r
}
