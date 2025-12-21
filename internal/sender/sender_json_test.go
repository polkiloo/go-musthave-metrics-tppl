package sender_test

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
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
	s := sender.NewJSONSender(host, port, ts.Client(), log, comp, "", nil, nil)

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
	s := sender.NewJSONSender(host, port, ts.Client(), log, comp, "", nil, nil)

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

	s := sender.NewJSONSender("127.0.0.1", 65535, cl, log, comp, "", nil, nil)

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

	s := sender.NewJSONSender("localhost", 1, cl, log, comp, "", nil, nil)

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
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
	}))
	defer ts.Close()

	host, port := hostPortFromServer(t, ts)
	cl := &http.Client{Timeout: 100 * time.Millisecond}
	log := &test.FakeLogger{}
	comp := test.NewFakeCompressor("gzip")

	s := sender.NewJSONSender(host, port, cl, log, comp, "", nil, nil)

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

	s := sender.NewJSONSender(host, port, ts.Client(), log, comp, "", nil, nil)

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
func TestJSONSender_Send_Success_NoCompression(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		io.Copy(io.Discard, r.Body)
		r.Body.Close()

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	log := &test.FakeLogger{}
	comp := test.NewFakeCompressor("gzip")
	host, port := hostPortFromServer(t, srv)
	s := sender.NewJSONSender(host, port, srv.Client(), log, comp, "", nil, nil)
	c := int64(5)
	g := 1.23
	mCounter := &models.Metrics{ID: "PollCount", MType: models.CounterType, Delta: &c}
	mGauge := &models.Metrics{ID: "Alloc", MType: models.GaugeType, Value: &g}

	s.Send([]*models.Metrics{mCounter, mGauge})

	if gotPath != "/update" {
		t.Fatalf("expected path /update, got %q", gotPath)
	}
}

func TestJSONSender_Send_Success_WithGzip(t *testing.T) {
	var gotCE, gotAE string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCE = r.Header.Get("Content-Encoding")
		gotAE = r.Header.Get("Accept-Encoding")
		io.Copy(io.Discard, r.Body)
		r.Body.Close()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	log := &test.FakeLogger{}
	comp := compression.NewGzip(gzip.BestSpeed)
	host, port := hostPortFromServer(t, srv)
	s := sender.NewJSONSender(host, port, srv.Client(), log, comp, "", nil, nil)

	var counterValue int64 = 123
	counter, _ := models.NewCounterMetrics("counter", &counterValue)

	s.Send([]*models.Metrics{counter})

	if gotCE != "gzip" {
		t.Fatalf("expected Content-Encoding=gzip, got %q", gotCE)
	}
	if gotAE != "gzip" {
		t.Fatalf("expected Accept-Encoding=gzip, got %q", gotAE)
	}
}

func TestJSONSender_Send_MarshalError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatalf("server should not be called on marshal error")
	}))
	t.Cleanup(srv.Close)

	log := &test.FakeLogger{}
	comp := test.NewFakeCompressor("gzip")

	host, port := hostPortFromServer(t, srv)
	s := sender.NewJSONSender(host, port, srv.Client(), log, comp, "", nil, nil)

	badValue := math.NaN()
	bad, _ := models.NewGaugeMetrics("nan", &badValue)

	s.Send([]*models.Metrics{bad})
}

func TestJSONSender_Send_BuildRequestError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatalf("server should not be called on build request error")
	}))
	t.Cleanup(srv.Close)

	log := &test.FakeLogger{}
	comp := test.NewFakeCompressor("gzip")

	s := sender.NewJSONSender("http://%zz", 0, srv.Client(), log, comp, "", nil, nil)

	var counterValue int64 = 123
	counter, _ := models.NewCounterMetrics("counter", &counterValue)

	s.Send([]*models.Metrics{counter})

}

func TestJSONSender_Send_DoError(t *testing.T) {

	cl := &http.Client{
		Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
			return nil, errors.New("network down")
		}),
		Timeout: 200 * time.Millisecond,
	}

	log := &test.FakeLogger{}
	comp := test.NewFakeCompressor("gzip")

	s := sender.NewJSONSender("http://example.invalid", 0, cl, log, comp, "", nil, nil)
	var counterValue int64 = 123
	counter, _ := models.NewCounterMetrics("counter", &counterValue)

	s.Send([]*models.Metrics{counter})
}

func TestJSONSender_Send_UnexpectedContentType(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		r.Body.Close()
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	log := &test.FakeLogger{}
	comp := test.NewFakeCompressor("gzip")
	host, port := hostPortFromServer(t, srv)
	s := sender.NewJSONSender(host, port, srv.Client(), log, comp, "", nil, nil)

	var counterValue int64 = 123
	counter, _ := models.NewCounterMetrics("counter", &counterValue)

	s.Send([]*models.Metrics{counter})
}

func TestJSONSender_Send_UnexpectedStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		r.Body.Close()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	log := &test.FakeLogger{}
	comp := test.NewFakeCompressor("gzip")
	host, port := hostPortFromServer(t, srv)
	s := sender.NewJSONSender(host, port, srv.Client(), log, comp, "", nil, nil)

	var counterValue int64 = 123
	counter, _ := models.NewCounterMetrics("counter", &counterValue)

	s.Send([]*models.Metrics{counter})
}

func TestJSONSender_Send_Headers_Golden(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		r.Body.Close()

		got = fmt.Sprintf("Content-Type: %s\nContent-Encoding: %s\nAccept-Encoding: %s\n",
			r.Header.Get("Content-Type"),
			r.Header.Get("Content-Encoding"),
			r.Header.Get("Accept-Encoding"),
		)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	log := &test.FakeLogger{}
	comp := test.NewFakeCompressor("gzip")
	host, port := hostPortFromServer(t, srv)
	s := sender.NewJSONSender(host, port, srv.Client(), log, comp, "", nil, nil)

	var counterValue int64 = 123
	counter, _ := models.NewCounterMetrics("counter", &counterValue)

	s.Send([]*models.Metrics{counter})

	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Join(filepath.Dir(filename), "testdata")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir testdata: %v", err)
	}
	golden := filepath.Join(dir, "headers.golden")

	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.WriteFile(golden, []byte(got), 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
	}
}

func BenchmarkJSONSender_Send_Gzip(b *testing.B) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		r.Body.Close()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}))
	b.Cleanup(srv.Close)

	log := &test.FakeLogger{}
	comp := compression.NewGzip(compression.BestSpeed)
	host, port := hostPortFromServer(b, srv)
	s := sender.NewJSONSender(host, port, srv.Client(), log, comp, "", nil, nil)

	data := bytes.Repeat([]byte("abcdefghijklmnopqrstuvwxyz0123456789"), 1024) // ~64KiB
	_ = data
	v := 0.0

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		v += 1
		m := &models.Metrics{ID: "bench", MType: models.GaugeType, Value: &v}
		s.Send([]*models.Metrics{m})
	}
}

func TestJSONSender_Send_BodyIsGzipCompressed(t *testing.T) {
	var compressed bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ce := r.Header.Get("Content-Encoding")
		if ce == "gzip" {
			gr, err := gzip.NewReader(r.Body)
			if err == nil {
				io.Copy(io.Discard, gr)
				gr.Close()
				compressed = true
			}
		}
		r.Body.Close()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	log := &test.FakeLogger{}
	comp := compression.NewGzip(compression.BestSpeed)
	host, port := hostPortFromServer(t, srv)
	s := sender.NewJSONSender(host, port, srv.Client(), log, comp, "", nil, nil)

	const n = 100
	metrics := make([]*models.Metrics, 0, n)

	for i := 0; i < n; i++ {
		if rand.Intn(2) == 0 {
			val := rand.Float64() * 100
			m, _ := models.NewGaugeMetrics(fmt.Sprintf("gauge_%d", i), &val)
			metrics = append(metrics, m)
		} else {
			delta := rand.Int63n(1000)
			m, _ := models.NewCounterMetrics(fmt.Sprintf("counter_%d", i), &delta)
			metrics = append(metrics, m)
		}
	}

	s.Send(metrics)

	if !compressed {
		t.Fatalf("server did not detect gzip-compressed body")
	}
}

func hostPortFromServer(tb testing.TB, ts *httptest.Server) (string, int) {
	tb.Helper()
	u, err := url.Parse(ts.URL)
	if err != nil {
		tb.Fatalf("parse url: %v", err)
	}
	host, portStr, err := net.SplitHostPort(u.Host)
	if err != nil {
		tb.Fatalf("split host port: %v", err)
	}
	p, err := strconv.Atoi(portStr)
	if err != nil {
		tb.Fatalf("atoi: %v", err)
	}
	return host, p
}

func TestJSONSender_SendBatch_SendsAll(t *testing.T) {
	var gotCount int
	var gotBody []byte
	var gotCT string
	var gotPath string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCount++
		gotCT = r.Header.Get("Content-Type")
		gotPath = r.URL.Path
		defer r.Body.Close()
		if r.Header.Get("Content-Encoding") == "gzip" {
			zr, _ := gzip.NewReader(r.Body)
			gotBody, _ = io.ReadAll(zr)
			zr.Close()
		} else {
			gotBody, _ = io.ReadAll(r.Body)
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	host, port := hostPortFromServer(t, ts)
	log := &test.FakeLogger{}
	comp := compression.NewGzip(compression.BestSpeed)
	s := sender.NewJSONSender(host, port, ts.Client(), log, comp, "", nil, nil)

	c := int64(5)
	g := 1.23
	mCounter := &models.Metrics{ID: "PollCount", MType: models.CounterType, Delta: &c}
	mGauge := &models.Metrics{ID: "Alloc", MType: models.GaugeType, Value: &g}

	s.SendBatch([]*models.Metrics{mCounter, mGauge})

	if gotCount != 1 {
		t.Fatalf("expected 1 request, got %d", gotCount)
	}
	if gotPath != "/updates" {
		t.Fatalf("expected path /updates, got %q", gotPath)
	}
	if !strings.HasPrefix(gotCT, "application/json") {
		t.Fatalf("content-type: want application/json, got %q", gotCT)
	}

	var ms []models.Metrics
	if err := json.Unmarshal(gotBody, &ms); err != nil {
		t.Fatalf("invalid json body: %v", err)
	}
	if len(ms) != 2 {
		t.Fatalf("want 2 metrics, got %d", len(ms))
	}
	if last := log.GetLastInfoMessage(); !strings.HasPrefix(last, "metrics batch sent") {
		t.Errorf("expected an info 'metrics batch sent', got %q", last)
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
