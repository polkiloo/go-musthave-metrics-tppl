package sender

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/models"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/test"
)

func TestPlainValue_GaugeOK(t *testing.T) {

	gaugeMetricValue := float64(1.23)
	mg, _ := models.NewGaugeMetrics(models.GaugeNames[0], &gaugeMetricValue)
	got, ok := plainValue(mg)
	if !ok {
		t.Fatalf("expected ok")
	}
	if got != "1.230000" {
		t.Fatalf("unexpected value: %q", got)
	}
}
func TestSender_Send_AllTypes(t *testing.T) {
	got := make(map[string]struct{})

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) >= 4 && parts[1] == "update" {
			key := parts[2] + ":" + parts[3]
			got[key] = struct{}{}
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	fl := &test.FakeLogger{}
	s := NewPlainSender(ts.URL, 0, ts.Client(), fl, "", nil)

	var c int64 = 2
	mc, _ := models.NewCounterMetrics(models.CounterNames[0], &c)

	gv := 1.23
	mg, _ := models.NewGaugeMetrics(models.GaugeNames[0], &gv)

	metrics := []*models.Metrics{mc, mg}
	s.Send(metrics)

}

func TestSender_Send_CreateRequestError(t *testing.T) {
	fl := &test.FakeLogger{}
	s := NewPlainSender("://bad_url", 8080, nil, fl, "", nil)

	gv := 1.23
	mg, _ := models.NewGaugeMetrics(models.GaugeNames[0], &gv)

	s.Send([]*models.Metrics{mg})
}

func TestSender_Send_HTTPError(t *testing.T) {
	port := getUnusedPort()
	fl := &test.FakeLogger{}
	s := NewPlainSender("http://127.0.0.1", port, nil, fl, "", nil)

	gv := 1.23
	mg, _ := models.NewGaugeMetrics(models.GaugeNames[0], &gv)
	s.Send([]*models.Metrics{mg})

}

func TestSender_Send_NonOKResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))
	defer ts.Close()
	host, port := parseHostPort(ts.URL)
	fl := &test.FakeLogger{}
	s := NewPlainSender(host, port, nil, fl, "", nil)

	gv := 1.23
	mg, _ := models.NewGaugeMetrics(models.GaugeNames[0], &gv)
	s.Send([]*models.Metrics{mg})
}

func TestSender_Send_EmptyMaps(t *testing.T) {
	fl := &test.FakeLogger{}
	s := NewPlainSender("http://localhost", 8080, nil, fl, "", nil)
	s.Send(nil)
}

func TestPlainSender_Log_NilMetric(t *testing.T) {
	fl := &test.FakeLogger{}
	s := &PlainSender{
		baseURL: "http://example.com",
		client:  &http.Client{Timeout: time.Second},
		log:     fl,
	}
	s.postMetric(context.Background(), nil)

	if !contains(fl.GetErrorMessages(), ErrSenderNilMetric.Error()) {
		t.Fatalf("expected error log: %+v, got: %+v", ErrSenderNilMetric.Error(), fl.GetErrorMessages())
	}
}

func TestPlainSender_MiddlewareAddsHeader(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Real-IP") != "127.0.0.1" {
			t.Fatalf("expected X-Real-IP header")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	mw := func(req *http.Request) { req.Header.Set("X-Real-IP", "127.0.0.1") }
	sender := NewPlainSender(ts.URL, 0, ts.Client(), &test.FakeLogger{}, "", mw)
	val := 1.0
	metric, _ := models.NewGaugeMetrics("g", &val)
	sender.Send([]*models.Metrics{metric})
}

func contains(ss []string, sub string) bool {
	for _, s := range ss {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

func parseHostPort(url string) (string, int) {
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")
	parts := strings.Split(url, ":")
	if len(parts) == 2 {
		return "http://" + parts[0], mustAtoi(parts[1])
	}
	host, port, _ := net.SplitHostPort(url)
	return "http://" + host, mustAtoi(port)
}

func mustAtoi(s string) int {
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return 0
		}
	}
	n, _ := net.LookupPort("tcp", s)
	return n
}

func getUnusedPort() int {
	l, _ := net.Listen("tcp", ":0")
	defer l.Close()
	_, portStr, _ := net.SplitHostPort(l.Addr().String())
	return mustAtoi(portStr)
}
