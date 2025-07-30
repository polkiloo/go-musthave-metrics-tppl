package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	models "github.com/polkiloo/go-musthave-metrics-tppl/internal/model"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/service"
)

func TestUpdate_NameEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fs := &service.FakeMetricService{}
	h := &GinHandler{service: fs}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{
		gin.Param{Key: "type", Value: string(models.GaugeType)},
		gin.Param{Key: "name", Value: ""},
		gin.Param{Key: "value", Value: "1.0"},
	}

	h.Update(c)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 404 when name empty, got %d", w.Code)
	}
	if fs.MName != "" || fs.MType != "" {
		t.Errorf("service was called unexpectedly: %+v", fs)
	}
}

func TestUpdate_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	expErr := errors.New("run error")
	fs := &service.FakeMetricService{Err: expErr}
	h := &GinHandler{service: fs}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{
		gin.Param{Key: "type", Value: string(models.CounterType)},
		gin.Param{Key: "name", Value: "hits"},
		gin.Param{Key: "value", Value: "5"},
	}

	h.Update(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 on service error, got %d", w.Code)
	}
	if body := w.Body.String(); body != expErr.Error() {
		t.Errorf("expected body %q, got %q", expErr.Error(), body)
	}
	if fs.MType != models.CounterType || fs.MName != "hits" || fs.MValue != "5" {
		t.Errorf("service called with wrong args: %+v", fs)
	}
}

func TestUpdate_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fs := &service.FakeMetricService{Err: nil}
	h := &GinHandler{service: fs}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{
		gin.Param{Key: "type", Value: string(models.GaugeType)},
		gin.Param{Key: "name", Value: "temp"},
		gin.Param{Key: "value", Value: "12.3"},
	}

	h.Update(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", w.Code)
	}
	if body := w.Body.String(); body != "ok" {
		t.Errorf("expected 'ok', got %q", body)
	}
	if fs.MType != models.GaugeType || fs.MName != "temp" || fs.MValue != "12.3" {
		t.Errorf("service called with wrong args: %+v", fs)
	}
}

func TestUpdate_UnknownMetricType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fs := &service.FakeMetricService{Err: service.ErrUnknownMetricType}
	h := &GinHandler{service: fs}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{
		{Key: "type", Value: "badtype"},
		{Key: "name", Value: "m"},
		{Key: "value", Value: "0"},
	}

	h.Update(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for unknown metric type, got %d", w.Code)
	}
	// body should be empty
	if body := w.Body.String(); body != "" {
		t.Errorf("expected empty body for unknown type, got %q", body)
	}
	// service called with raw metricType
	if fs.MType != models.MetricType("badtype") {
		t.Errorf("service called with type %v; want badtype", fs.MType)
	}
}
