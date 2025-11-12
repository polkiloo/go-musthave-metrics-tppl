package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/audit"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/compression"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/db"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/logger"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/models"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/service"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/sign"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/storage"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/test"
	"go.uber.org/fx"
)

func TestRegisterUpdate_JSONRoute_ContentTypeCheck(t *testing.T) {
	fs := &test.FakeMetricService{}
	h := newTestGinHandler(fs)
	r := newRouterWithHandler(h)

	w := test.DoJSON(r, "/update", map[string]any{"id": "Alloc", "type": "gauge", "value": 1.0}, "text/plain")
	if w.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("POST /update wrong content-type: got %d, want %d", w.Code, http.StatusUnsupportedMediaType)
	}
}

func TestRegisterUpdate_JSONRoute_ContentTypeCheckc(t *testing.T) {
	fs := &test.FakeMetricService{}
	h := newTestGinHandler(fs)
	r := newRouterWithHandler(h)

	w := test.DoJSON(r, "/update/", map[string]any{"id": "Alloc", "type": "gauge", "value": 1.0}, "text/plain")
	if w.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("POST /update wrong content-type: got %d, want %d", w.Code, http.StatusUnsupportedMediaType)
	}
}

func TestRegisterUpdate_JSONRoute_CallsUpdateJSON(t *testing.T) {
	f := float64(1.23)
	m, _ := models.NewGaugeMetrics(models.GaugeNames[0], &f)

	fs := &test.FakeMetricService{}
	h := newTestGinHandler(fs)
	r := newRouterWithHandler(h)

	w := test.DoJSON(r, "/update", m, "application/json; charset=utf-8")
	if w.Code != http.StatusOK {
		t.Fatalf("POST /update expected 200, got %d", w.Code)
	}
	if fs.Metric.ID != models.GaugeNames[0] || fs.Metric.MType != models.GaugeType {
		t.Fatalf("ProcessUpdate not called correctly")
	}
	if fs.Metric.ID != models.GaugeNames[0] {
		t.Fatalf("ProcessGetValue not called with id=Alloc, got %q", fs.Metric.ID)
	}
}

func TestRegisterUpdate_PlainRoute_CallsUpdatePlain(t *testing.T) {
	fs := &test.FakeMetricService{}
	h := newTestGinHandler(fs)
	r := newRouterWithHandler(h)

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/Alloc/1.230000", bytes.NewReader(nil))
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("POST /update/:type/:name/:value expected 200, got %d", w.Code)
	}
	if fs.Metric.ID != "Alloc" || fs.Metric.MType != models.GaugeType {
		t.Fatalf("ProcessUpdate not called correctly")
	}
}

func TestRegisterGetValue_JSONRoute_ContentTypeCheck(t *testing.T) {
	fs := &test.FakeMetricService{}
	h := newTestGinHandler(fs)
	r := newRouterWithHandler(h)

	w := test.DoJSON(r, "/value/", map[string]any{"id": "Alloc"}, "text/plain")
	if w.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("POST /value wrong content-type: got %d, want %d", w.Code, http.StatusUnsupportedMediaType)
	}
}

func TestRegisterGetValue_JSONRoute_CallsGetValueJSON(t *testing.T) {
	f := float64(1.23)
	m, _ := models.NewGaugeMetrics(models.GaugeNames[0], &f)

	fs := &test.FakeMetricService{Metric: *m}
	h := newTestGinHandler(fs)
	r := newRouterWithHandler(h)

	w := test.DoJSON(r, "/value", m, "application/json")
	if w.Code != http.StatusOK {
		t.Fatalf("POST /value expected 200, got %d", w.Code)
	}
	if fs.Metric.ID != m.ID {
		t.Fatalf("ProcessGetValue not called with id=Alloc, got %q", fs.Metric.ID)
	}
}

func TestRegisterUpdate_JSONRoute_ResponseIsJSON(t *testing.T) {
	v := int64(33)
	m, _ := models.NewCounterMetrics(models.CounterNames[0], &v)

	fs := &test.FakeMetricService{Metric: *m}
	h := newTestGinHandler(fs)
	r := newRouterWithHandler(h)

	w := test.DoJSON(r, "/update", m, "application/json")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type = %q, want %q", ct, "application/json")
	}

	var resp models.Metrics
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response is not valid JSON: %v\nbody=%s", err, w.Body.String())
	}
	if resp.ID != m.ID || resp.MType != models.CounterType || *resp.Delta != *m.Delta {
		t.Fatalf("unexpected JSON body: %+v", resp)
	}
}

func newRouterWithHandler(h *GinHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h.RegisterUpdate(r)
	h.RegisterGetValue(r)
	return r
}

func TestRegisterRoutes_RegistersOnce_NoPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	h := &GinHandler{service: &test.FakeMetricService{}}

	before := len(r.Routes())

	RegisterRoutes(r, h, nil)

	after := len(r.Routes())
	if after <= before {
		t.Fatalf("expected routes to increase: before=%d after=%d", before, after)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)
}

func Test_register_AddsMiddlewareAndRegisters(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	h := &GinHandler{service: &test.FakeMetricService{}}

	var l logger.Logger = &test.FakeLogger{}

	if len(r.Handlers) != 0 {
		t.Fatalf("expected no global middlewares initially, got %d", len(r.Handlers))
	}
	before := len(r.Routes())

	var c compression.Compressor = test.NewFakeCompressor("")

	register(struct {
		fx.In
		R     *gin.Engine
		H     *GinHandler
		L     logger.Logger
		C     compression.Compressor
		S     sign.Signer
		K     sign.SignKey
		A     audit.Publisher `optional:"true"`
		Clock audit.Clock     `optional:"true"`
		Pool  db.Pool         `optional:"true"`
	}{R: r, H: h, L: l, C: c, S: sign.NewSignerSHA256(), K: ""})

	if len(r.Handlers) == 0 {
		t.Fatalf("expected global middleware to be added")
	}
	after := len(r.Routes())
	if after <= before {
		t.Fatalf("expected routes to be registered: before=%d after=%d", before, after)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)
}

func TestNewGinHandler_ServiceConcreteTypeIsMetricService(t *testing.T) {
	h := NewGinHandler(service.NewMetricService(storage.NewMemStorage()), NewJSONMetricsPool())

	got := reflect.TypeOf(h.service).String()
	want := "*service.MetricService"
	if got != want {
		t.Fatalf("unexpected concrete service type: want %q, got %q", want, got)
	}
}
