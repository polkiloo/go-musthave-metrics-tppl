package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	models "github.com/polkiloo/go-musthave-metrics-tppl/internal/model"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/service"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/test"
)

func TestGetValue_NameEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fs := &test.FakeMetricService{Err: service.ErrUnknownMetricType}
	h := &GinHandler{service: fs}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{
		gin.Param{Key: "type", Value: string(models.GaugeType)},
		gin.Param{Key: "name", Value: ""},
	}

	h.GetValue(c)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty name, got %d", w.Code)
	}
}

func TestGetValue_UnknownMetricType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fs := &test.FakeMetricService{Err: service.ErrUnknownMetricType}
	h := &GinHandler{service: fs}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{
		gin.Param{Key: "type", Value: "unknown"},
		gin.Param{Key: "name", Value: "foo"},
	}

	h.GetValue(c)
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

	h.GetValue(c)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for service error, got %d", w.Code)
	}
}

func TestGetValue_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	fs := &test.FakeMetricService{MValue: "123"}
	h := &GinHandler{service: fs}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{
		gin.Param{Key: "type", Value: string(models.CounterType)},
		gin.Param{Key: "name", Value: "hits"},
	}

	h.GetValue(c)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", w.Code)
	}
	if w.Body.String() != "123" {
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
		{Key: "name", Value: "missing"},
	}

	h.GetValue(c)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 NotFound on metric not found, got %d", w.Code)
	}
	if fs.MName != "missing" {
		t.Errorf("service called with wrong name %q", fs.MName)
	}
}
