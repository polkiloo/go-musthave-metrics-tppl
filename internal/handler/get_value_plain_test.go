package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/models"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/service"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/test"
)

func TestGetValue_NameEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fs := &test.FakeMetricService{Err: models.ErrInvalidMetricType}
	h := &GinHandler{service: fs}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{
		gin.Param{Key: "type", Value: string(models.GaugeType)},
		gin.Param{Key: "name", Value: ""},
	}

	h.GetValuePlain(c)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty name, got %d", w.Code)
	}
}

func TestGetValue_UnknownMetricType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fs := &test.FakeMetricService{Err: models.ErrInvalidMetricType}
	h := &GinHandler{service: fs}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{
		gin.Param{Key: "type", Value: "unknown"},
		gin.Param{Key: "name", Value: "foo"},
	}

	h.GetValuePlain(c)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for unknown metric type, got %d", w.Code)
	}
}

func TestGetValue_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	expErr := errors.New("some error")
	fs := &test.FakeMetricService{Err: expErr}
	h := &GinHandler{service: fs}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{
		gin.Param{Key: "type", Value: string(models.CounterType)},
		gin.Param{Key: "name", Value: "hits"},
	}

	h.GetValuePlain(c)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for service error, got %d", w.Code)
	}
}

func TestGetValue_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var v int64 = 123
	mc, _ := models.NewCounterMetrics(models.CounterNames[0], &v)
	fs := &test.FakeMetricService{Metric: *mc}
	h := &GinHandler{service: fs}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{
		gin.Param{Key: "type", Value: string(mc.MType)},
		gin.Param{Key: "name", Value: mc.ID},
	}

	h.GetValuePlain(c)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", w.Code)
	}
	if w.Body.String() != strconv.FormatInt(v, 10) {
		t.Errorf("expected body '123', got %q", w.Body.String())
	}
}

func TestGetValue_MetricNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	fs := &test.FakeMetricService{Err: service.ErrMetricNotFound}
	h := &GinHandler{service: fs}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{
		{Key: "type", Value: string(models.CounterType)},
		{Key: "name", Value: models.CounterNames[0]},
	}

	h.GetValuePlain(c)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 NotFound on metric not found, got %d", w.Code)
	}
}
