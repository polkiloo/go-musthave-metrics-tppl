package sender_test

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/compression"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/models"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/sender"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/test"
)

func TestJSONSender_Send_SendsAll(t *testing.T) {
	var gotCount int
	var gotBodies [][]byte
	var gotCT []string
	var gotPaths []string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCount++
		gotCT = append(gotCT, r.Header.Get("Content-Type"))
		gotPaths = append(gotPaths, r.URL.Path)
		defer r.Body.Close()
		var b []byte
		if r.Header.Get("Content-Encoding") == "gzip" {
			zr, _ := gzip.NewReader(r.Body)
			b, _ = io.ReadAll(zr)
			zr.Close()
		} else {
			b, _ = io.ReadAll(r.Body)
		}
		gotBodies = append(gotBodies, append([]byte(nil), b...))

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	host, port := hostPortFromServer(t, ts)
	log := &test.FakeLogger{}
	comp := compression.NewGzip(compression.BestSpeed)
	s := sender.NewJSONSender(host, port, ts.Client(), log, comp)

	c := int64(5)
	g := 1.23
	mCounter := &models.Metrics{ID: "PollCount", MType: models.CounterType, Delta: &c}
	mGauge := &models.Metrics{ID: "Alloc", MType: models.GaugeType, Value: &g}

	s.Send([]*models.Metrics{mCounter, mGauge})

	if gotCount != 2 {
		t.Fatalf("expected 2 requests, got %d", gotCount)
	}
	for i, p := range gotPaths {
		if p != "/update" {
			t.Errorf("req %d path: want /update, got %q", i, p)
		}
	}
	for i, ct := range gotCT {
		if !strings.HasPrefix(ct, "application/json") {
			t.Errorf("req %d content-type: want application/json, got %q", i, ct)
		}
	}

	checkBody := func(i int, wantID string, wantType models.MetricType) {
		var m models.Metrics
		if err := json.Unmarshal(gotBodies[i], &m); err != nil {
			t.Fatalf("req %d: invalid json body: %v", i, err)
		}
		if m.ID != wantID || m.MType != wantType {
			t.Errorf("req %d: want (id=%s,type=%s), got (id=%s,type=%s)",
				i, wantID, wantType, m.ID, m.MType)
		}
	}
	checkBody(0, "PollCount", models.CounterType)
	checkBody(1, "Alloc", models.GaugeType)

	if last := log.GetLastInfoMessage(); !strings.HasPrefix(last, "metric sent") {
		t.Errorf("expected an info 'metric sent', got %q", last)
	}
}

func TestJSONSender_NonJSONContentType_Logged(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	host, port := hostPortFromServer(t, ts)
	log := &test.FakeLogger{}
	comp := test.NewFakeCompressor("gzip")
	s := sender.NewJSONSender(host, port, ts.Client(), log, comp)

	g := 1.0
	s.Send([]*models.Metrics{{ID: "Alloc", MType: models.GaugeType, Value: &g}})

	found := false
	for _, msg := range log.GetErrorMessages() {
		if strings.HasPrefix(msg, "unexpected content-type") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'unexpected content-type' to be logged")
	}
}

func TestJSONSender_NonOKStatus_Logged(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer ts.Close()

	host, port := hostPortFromServer(t, ts)
	log := &test.FakeLogger{}
	comp := test.NewFakeCompressor("gzip")
	s := sender.NewJSONSender(host, port, ts.Client(), log, comp)

	g := 1.0
	s.Send([]*models.Metrics{{ID: "Alloc", MType: models.GaugeType, Value: &g}})

	found := false
	for _, msg := range log.GetErrorMessages() {
		if strings.HasPrefix(msg, "unexpected status") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'unexpected status' to be logged")
	}
}

func TestJSONSender_ClientError_Logged(t *testing.T) {
	wantErr := errors.New("transport down")
	cl := &http.Client{
		Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
			return nil, wantErr
		}),
		Timeout: 200 * time.Millisecond,
	}

	log := &test.FakeLogger{}
	comp := test.NewFakeCompressor("gzip")

	s := sender.NewJSONSender("127.0.0.1", 65535, cl, log, comp)

	g := 1.0
	s.Send([]*models.Metrics{{ID: "Alloc", MType: models.GaugeType, Value: &g}})

	found := false
	for _, msg := range log.GetErrorMessages() {
		if strings.HasPrefix(msg, "post metric failed") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'post metric failed' to be logged")
	}
}

func TestJSONSender_MarshalError_Logged(t *testing.T) {
	// До сети не дойдём — важен сам факт ошибки marshal.
	cl := &http.Client{
		Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(bytes.NewReader(nil)),
			}, nil
		}),
		Timeout: 200 * time.Millisecond,
	}

	log := &test.FakeLogger{}
	comp := test.NewFakeCompressor("gzip")

	s := sender.NewJSONSender("localhost", 1, cl, log, comp)

	nan := math.NaN()
	s.Send([]*models.Metrics{{ID: "Alloc", MType: models.GaugeType, Value: &nan}})

	found := false
	for _, msg := range log.GetErrorMessages() {
		if strings.HasPrefix(msg, "marshal metric failed") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'marshal metric failed' to be logged")
	}
}

func TestJSONSender_Send_RespectsClientTimeout(t *testing.T) {
	// Медленный сервер; берём короткий таймаут клиента, чтобы тест был быстрым.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
	}))
	defer ts.Close()

	host, port := hostPortFromServer(t, ts)
	cl := &http.Client{Timeout: 100 * time.Millisecond}
	log := &test.FakeLogger{}
	comp := test.NewFakeCompressor("gzip")

	s := sender.NewJSONSender(host, port, cl, log, comp)

	val := 1.0
	start := time.Now()
	s.Send([]*models.Metrics{{ID: "Alloc", MType: models.GaugeType, Value: &val}})
	if time.Since(start) > time.Second {
		t.Fatalf("send took too long")
	}

	found := false
	for _, msg := range log.GetErrorMessages() {
		if strings.HasPrefix(msg, "post metric failed") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected timeout logged as 'post metric failed'")
	}
}

func TestJSONSender_BodyShape_Minimal(t *testing.T) {
	var body []byte
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") == "gzip" {
			zr, _ := gzip.NewReader(r.Body)
			body, _ = io.ReadAll(zr)
			zr.Close()
		} else {
			body, _ = io.ReadAll(r.Body)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	host, port := hostPortFromServer(t, ts)
	log := &test.FakeLogger{}
	comp := compression.NewGzip(compression.BestSpeed)

	s := sender.NewJSONSender(host, port, ts.Client(), log, comp)

	delta := int64(7)
	gauge := 3.14
	in := []*models.Metrics{
		{ID: "A", MType: models.CounterType, Delta: &delta},
		{ID: "B", MType: models.GaugeType, Value: &gauge},
	}
	s.Send(in)

	var m models.Metrics
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&m); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if m.ID == "" {
		t.Fatalf("empty id in body")
	}
}

func hostPortFromServer(t *testing.T, ts *httptest.Server) (string, int) {
	t.Helper()
	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}
	host, portStr, err := net.SplitHostPort(u.Host)
	if err != nil {
		t.Fatalf("split host port: %v", err)
	}
	p, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatalf("atoi: %v", err)
	}
	return host, p
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
