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

// func TestGetValueJSON_NotFound_UnknownName(t *testing.T) {
// 	svc := &fakeService{getErr: models.ErrUnknownMetricName}
// 	r := makeRouterValue(svc)

// 	w := doJSON(r, "/value", map[string]any{"id": "Nope"}, "application/json")
// 	if w.Code != http.StatusNotFound {
// 		t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
// 	}
// }

// func TestGetValueJSON_OtherError_400(t *testing.T) {
// 	svc := &fakeService{getErr: errors.New("storage down")}
// 	r := makeRouterValue(svc)

// 	w := doJSON(r, "/value", map[string]any{"id": "Alloc"}, "application/json")
// 	if w.Code != http.StatusBadRequest {
// 		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
// 	}
// }

// func TestGetValueJSON_Success_Gauge_NoTypeProvided(t *testing.T) {
// 	val := 2.5
// 	out := &models.Metrics{ID: "Alloc", MType: models.GaugeType, Value: &val}

// 	svc := &fakeService{getResult: out}
// 	r := makeRouterValue(svc)

// 	w := doJSON(r, "/value", map[string]any{"id": "Alloc"}, "application/json; charset=utf-8")
// 	if w.Code != http.StatusOK {
// 		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
// 	}
// 	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
// 		t.Fatalf("Content-Type = %q, want %q", ct, "application/json")
// 	}

// 	var got models.Metrics
// 	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
// 		t.Fatalf("unmarshal resp: %v", err)
// 	}
// 	if got.MType != models.GaugeType || got.ID != "Alloc" || got.Value == nil || *got.Value != val || got.Delta != nil {
// 		t.Fatalf("unexpected body: %+v", got)
// 	}
// 	if svc.lastGetName != "Alloc" {
// 		t.Fatalf("ProcessGetValue called with %q, want %q", svc.lastGetName, "Alloc")
// 	}
// }

// func TestGetValueJSON_Success_Counter_WithMatchingType(t *testing.T) {
// 	d := int64(42)
// 	out := &models.Metrics{ID: "PollCount", MType: models.CounterType, Delta: &d}

// 	svc := &fakeService{getResult: out}
// 	r := makeRouterValue(svc)

// 	// Клиент прислал корректный type
// 	w := doJSON(r, "/value", map[string]any{"id": "PollCount", "type": "counter"}, "application/json")
// 	if w.Code != http.StatusOK {
// 		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
// 	}

// 	var got models.Metrics
// 	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
// 		t.Fatalf("unmarshal resp: %v", err)
// 	}
// 	if got.MType != models.CounterType || got.ID != "PollCount" || got.Delta == nil || *got.Delta != d || got.Value != nil {
// 		t.Fatalf("unexpected body: %+v", got)
// 	}
// 	if svc.lastGetName != "PollCount" {
// 		t.Fatalf("ProcessGetValue called with %q, want %q", svc.lastGetName, "PollCount")
// 	}
// }

// func TestGetValueJSON_TypeMismatch_400(t *testing.T) {
// 	d := int64(7)
// 	out := &models.Metrics{ID: "PollCount", MType: models.CounterType, Delta: &d}

// 	svc := &fakeService{getResult: out}
// 	r := makeRouterValue(svc)

// 	// Клиент прислал type "gauge", а фактический — counter
// 	w := doJSON(r, "/value", map[string]any{"id": "PollCount", "type": "gauge"}, "application/json")
// 	if w.Code != http.StatusBadRequest {
// 		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
// 	}
// }

// func TestGetValueJSON_TypeUnknown_400(t *testing.T) {
// 	val := 1.0
// 	out := &models.Metrics{ID: "Alloc", MType: models.GaugeType, Value: &val}

// 	svc := &fakeService{getResult: out}
// 	r := makeRouterValue(svc)

// 	// Клиент прислал непонятный type — трактуем как несовпадение с фактом
// 	w := doJSON(r, "/value", map[string]any{"id": "Alloc", "type": "weird"}, "application/json")
// 	if w.Code != http.StatusBadRequest {
// 		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
// 	}
// }

func setupRouterWithGetValueJSON(srvc *test.FakeMetricService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := &GinHandler{service: srvc}
	r := gin.New()
	r.POST("/value", h.GetValueJSON)
	return r
}
