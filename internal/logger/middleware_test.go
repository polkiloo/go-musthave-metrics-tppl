package logger

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type logEntry struct {
	msg    string
	fields map[string]any
}

type fakeLogger struct {
	mu      sync.Mutex
	entries []logEntry
}

func (f *fakeLogger) WriteInfo(msg string, kv ...any) {
	fs := map[string]any{}
	for i := 0; i+1 < len(kv); i += 2 {
		if k, ok := kv[i].(string); ok {
			fs[k] = kv[i+1]
		}
	}
	f.mu.Lock()
	f.entries = append(f.entries, logEntry{msg: msg, fields: fs})
	f.mu.Unlock()
}
func (f *fakeLogger) WriteError(msg string, kv ...any) { f.WriteInfo(msg, kv...) }
func (f *fakeLogger) Sync() error                      { return nil }

func (f *fakeLogger) snapshot() []logEntry {
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := make([]logEntry, len(f.entries))
	copy(cp, f.entries)
	return cp
}

func TestMiddleware_LogsRequestAndResponse_WithBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	fl := &fakeLogger{}
	r := gin.New()
	r.Use(Middleware(fl))

	const body = "pong"
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, body)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping?x=1", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d, want %d", w.Code, http.StatusOK)
	}

	entries := fl.snapshot()
	if len(entries) != 2 {
		t.Fatalf("expected 2 log entries (request, response), got %d", len(entries))
	}

	reqEnt := entries[0]
	if reqEnt.msg != "request" {
		t.Errorf("first entry msg=%q; want %q", reqEnt.msg, "request")
	}
	if got := reqEnt.fields["method"]; got != http.MethodGet {
		t.Errorf("request.method=%v; want %v", got, http.MethodGet)
	}
	if got := reqEnt.fields["uri"]; got != "/ping?x=1" {
		t.Errorf("request.uri=%v; want %v", got, "/ping?x=1")
	}
	dur, ok := reqEnt.fields["duration"].(time.Duration)
	if !ok {
		t.Fatalf("request.duration type=%T; want time.Duration", reqEnt.fields["duration"])
	}
	if dur < 0 {
		t.Errorf("request.duration=%v; want >=0", dur)
	}

	respEnt := entries[1]
	if respEnt.msg != "response" {
		t.Errorf("second entry msg=%q; want %q", respEnt.msg, "response")
	}
	if got := respEnt.fields["status"]; got != http.StatusOK {
		t.Errorf("response.status=%v; want %v", got, http.StatusOK)
	}
	if got := respEnt.fields["size"]; got != len(body) {
		t.Errorf("response.size=%v; want %v", got, len(body))
	}
}

func TestMiddleware_LogsResponseSizeZero_WhenNoBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	fl := &fakeLogger{}
	r := gin.New()
	r.Use(Middleware(fl))

	r.GET("/nobody", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/nobody", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("unexpected status: got %d, want %d", w.Code, http.StatusNoContent)
	}

	entries := fl.snapshot()
	if len(entries) != 2 {
		t.Fatalf("expected 2 log entries (request, response), got %d", len(entries))
	}

	reqEnt := entries[0]
	if reqEnt.msg != "request" {
		t.Errorf("first entry msg=%q; want %q", reqEnt.msg, "request")
	}
	if got := reqEnt.fields["method"]; got != http.MethodGet {
		t.Errorf("request.method=%v; want %v", got, http.MethodGet)
	}
	if got := reqEnt.fields["uri"]; got != "/nobody" {
		t.Errorf("request.uri=%v; want %v", got, "/nobody")
	}
	if d, ok := reqEnt.fields["duration"].(time.Duration); !ok || d < 0 {
		t.Errorf("request.duration invalid: %v (ok=%v)", reqEnt.fields["duration"], ok)
	}

	respEnt := entries[1]
	if respEnt.msg != "response" {
		t.Errorf("second entry msg=%q; want %q", respEnt.msg, "response")
	}
	if got := respEnt.fields["status"]; got != http.StatusNoContent {
		t.Errorf("response.status=%v; want %v", got, http.StatusNoContent)
	}
	if got := respEnt.fields["size"]; got != 0 {
		t.Errorf("response.size=%v; want %v", got, 0)
	}
}
