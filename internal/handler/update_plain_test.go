package handler

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/models"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/service"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/test"
)

func TestUpdate_NameEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fs := &test.FakeMetricService{}
	h := &GinHandler{service: fs}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{
		gin.Param{Key: "type", Value: string(models.GaugeType)},
		gin.Param{Key: "name", Value: ""},
		gin.Param{Key: "value", Value: "1.0"},
	}

	h.UpdatePlain(c)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 404 when name empty, got %d", w.Code)
	}
	if fs.Metric.ID != "" || fs.Metric.MType != "" {
		t.Errorf("service was called unexpectedly: %+v", fs)
	}
}

func TestUpdate_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	expErr := service.ErrMetricNotFound
	fs := &test.FakeMetricService{Err: expErr}
	h := &GinHandler{service: fs}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	j := int64(5)
	m, _ := models.NewCounterMetrics(models.CounterNames[0], &j)
	c.Params = gin.Params{
		gin.Param{Key: "type", Value: string(m.MType)},
		gin.Param{Key: "name", Value: m.ID},
		gin.Param{Key: "value", Value: strconv.FormatInt(*m.Delta, 10)},
	}
	h.UpdatePlain(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 on service error, got %d", w.Code)
	}
	if body := w.Body.String(); body != expErr.Error() {
		t.Errorf("expected body %q, got %q", expErr.Error(), body)
	}
}

func TestUpdate_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fs := &test.FakeMetricService{Err: nil}
	h := &GinHandler{service: fs}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	j := float64(12.3)
	m, _ := models.NewGaugeMetrics(models.GaugeNames[0], &j)
	c.Params = gin.Params{
		gin.Param{Key: "type", Value: string(m.MType)},
		gin.Param{Key: "name", Value: m.ID},
		gin.Param{Key: "value", Value: strconv.FormatFloat(*m.Value, 'f', -1, 64)},
	}
	h.UpdatePlain(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", w.Code)
	}
	if body := w.Body.String(); body != "ok" {
		t.Errorf("expected 'ok', got %q", body)
	}
	if fs.Metric.MType != models.GaugeType || fs.Metric.ID != m.ID || *fs.Metric.Value != j {
		t.Errorf("service called with wrong args: %+v", fs)
	}
}

func TestUpdate_UnknownMetric(t *testing.T) {
	gin.SetMode(gin.TestMode)

	fs := &test.FakeMetricService{}
	h := &GinHandler{service: fs}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	metricType := "uknown_metric_type"

	c.Params = gin.Params{
		gin.Param{Key: "type", Value: metricType},
	}
	h.UpdatePlain(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 on service error, got %d", w.Code)
	}

}

func TestUpdatePlain_AfterUpdateHook(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fs := &test.FakeMetricService{}
	h := &GinHandler{service: fs}
	called := 0
	h.SetAfterUpdateHook(func() { called++ })
	r := gin.New()
	r.POST("/update/:type/:name/:value", h.UpdatePlain)

	j := float64(5.5)
	m, _ := models.NewGaugeMetrics("g1", &j)
	url := "/update/" + string(m.MType) + "/" + m.ID + "/" + strconv.FormatFloat(*m.Value, 'f', -1, 64)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, url, nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if called != 1 {
		t.Fatalf("afterUpdate not called, got %d", called)
	}
}
