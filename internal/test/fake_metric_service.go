package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/models"
)

type FakeMetricService struct {
	Err       error
	Metric    models.Metrics
	SaveCalls int
	LoadCalls int
}

func (f *FakeMetricService) ProcessUpdate(m *models.Metrics) error {
	f.Metric.MType = m.MType
	f.Metric.ID = m.ID

	switch m.MType {
	case models.GaugeType:
		f.Metric.Value = m.Value
	case models.CounterType:
		f.Metric.Delta = m.Delta
	default:
		return models.ErrMetricInvalidType
	}
	return f.Err
}

func (f *FakeMetricService) ProcessGetValue(metricName string, metricType models.MetricType) (*models.Metrics, error) {
	var m *models.Metrics

	switch {
	case models.IsGauge(metricType):
		m, _ = models.NewGaugeMetrics(metricName, f.Metric.Value)
	case models.IsCounter(metricType):
		m, _ = models.NewCounterMetrics(metricName, f.Metric.Delta)
	default:
		return nil, f.Err
	}
	return m, f.Err
}

func (f *FakeMetricService) SaveFile(path string) error {
	f.SaveCalls++
	return nil
}

func (f *FakeMetricService) LoadFile(path string) error {
	f.LoadCalls++
	return nil
}

func DoJSON(r *gin.Engine, url string, body any, contentType string) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(http.MethodPost, url, &buf)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func DoGET(r *gin.Engine, url, accept string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, url, nil)
	if accept != "" {
		req.Header.Set("Accept", accept)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
