package audit

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	test "github.com/polkiloo/go-musthave-metrics-tppl/internal/test"
)

type failingWriter struct {
	err error
}

func (f failingWriter) Write([]byte) (int, error) { return 0, f.err }
func (f failingWriter) Close() error              { return nil }

func TestFileObserverSuccess(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "audit")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	_ = file.Close()

	obs := NewFileObserver(file.Name())
	if err := obs.Notify(context.Background(), Event{Metrics: []string{"a"}}); err != nil {
		t.Fatalf("notify: %v", err)
	}

	content, err := os.ReadFile(file.Name())
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if len(bytes.TrimSpace(content)) == 0 {
		t.Fatalf("expected content written")
	}
}

func TestFileObserverErrors(t *testing.T) {
	orig := openFile
	t.Cleanup(func() { openFile = orig })

	obs := NewFileObserver("")
	if err := obs.Notify(context.Background(), Event{}); err == nil {
		t.Fatalf("expected error for empty path")
	}

	openFile = func(string, int, os.FileMode) (fileWriter, error) { return nil, errors.New("open") }
	obs = NewFileObserver("path")
	if err := obs.Notify(context.Background(), Event{}); err == nil || err.Error() != "open audit file: open" {
		t.Fatalf("unexpected error: %v", err)
	}

	openFile = func(string, int, os.FileMode) (fileWriter, error) {
		return failingWriter{err: errors.New("write")}, nil
	}
	if err := obs.Notify(context.Background(), Event{}); err == nil || err.Error() != "write audit event: write" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHTTPObserver(t *testing.T) {
	var nilObserver *httpObserver
	if err := nilObserver.Notify(context.Background(), Event{}); err == nil || err.Error() != "http observer not configured" {
		t.Fatalf("unexpected error for nil observer: %v", err)
	}

	if _, err := NewHTTPObserver("://bad", nil); err == nil {
		t.Fatalf("expected parse error")
	}

	obs, err := NewHTTPObserver("http://example.com", test.RoundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("boom")
	}))
	if err != nil {
		t.Fatalf("create observer: %v", err)
	}
	if err := obs.Notify(context.Background(), Event{}); err == nil || err.Error() != "send audit request: boom" {
		t.Fatalf("unexpected error: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	obs, err = NewHTTPObserver(server.URL, http.DefaultClient)
	if err != nil {
		t.Fatalf("observer: %v", err)
	}
	if err := obs.Notify(context.Background(), Event{}); err != nil {
		t.Fatalf("notify: %v", err)
	}

	badStatus := test.RoundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 500, Status: "500", Body: io.NopCloser(bytes.NewReader(nil))}, nil
	})
	endpoint, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("parse server url: %v", err)
	}
	obs = &httpObserver{endpoint: endpoint, client: badStatus}
	if err := obs.Notify(context.Background(), Event{}); err == nil {
		t.Fatalf("expected status error")
	}
}
