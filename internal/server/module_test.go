package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/handler"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/test"
	"go.uber.org/fx"
)

type fakeLifecycle struct{ hooks []fx.Hook }

func (f *fakeLifecycle) Append(h fx.Hook) { f.hooks = append(f.hooks, h) }

func TestRun_OnStart_SuccessPath_CoversGoroutine(t *testing.T) {
	t.Cleanup(resetHooksOverrides())
	gin.SetMode(gin.TestMode)

	lc := &fakeLifecycle{}
	engine := gin.New()
	cfg := &AppConfig{Host: "127.0.0.1", Port: 18080}

	var wg sync.WaitGroup
	wg.Add(1)

	var gotAddr string

	serverRunner = func(srv *http.Server) error {
		defer wg.Done()
		gotAddr = srv.Addr
		return nil
	}
	serverShutdown = func(context.Context, *http.Server) error { return nil }

	logger := &test.FakeLogger{}
	hand := handler.NewGinHandler(&test.FakeMetricService{}, handler.NewJSONMetricsPool())
	run(lc, engine, cfg, logger, hand)

	if len(lc.hooks) != 1 {
		t.Fatalf("expected 1 hook, got %d", len(lc.hooks))
	}
	h := lc.hooks[0]

	if err := h.OnStart(context.Background()); err != nil {
		t.Fatalf("OnStart error: %v", err)
	}

	wg.Wait()

	if gotAddr != "127.0.0.1:18080" {
		t.Errorf("got addr %q; want %q", gotAddr, "127.0.0.1:18080")
	}

	infoMessages := logger.GetInfoMessages()
	if len(infoMessages) == 0 {
		t.Fatalf("expected at least one info message, got none")
	}

	expectedLog := "server listening addr=http://127.0.0.1:18080"
	if !strings.Contains(infoMessages[0], expectedLog) {
		t.Errorf("unexpected log: %q", infoMessages[0])
	}

	if err := h.OnStop(context.Background()); err != nil {
		t.Errorf("OnStop error: %v", err)
	}
}

func TestRun_OnStart_FailurePath_CoversFatal(t *testing.T) {
	t.Cleanup(resetHooksOverrides())
	gin.SetMode(gin.TestMode)

	lc := &fakeLifecycle{}
	engine := gin.New()
	cfg := &AppConfig{Host: "localhost", Port: 9999}

	var wg sync.WaitGroup
	wg.Add(1)

	serverRunner = func(srv *http.Server) error {
		defer wg.Done()
		return errors.New("bind failed")
	}

	serverShutdown = func(context.Context, *http.Server) error { return nil }

	logger := &test.FakeLogger{}

	hand := handler.NewGinHandler(&test.FakeMetricService{}, handler.NewJSONMetricsPool())
	run(lc, engine, cfg, logger, hand)

	if len(lc.hooks) != 1 {
		t.Fatalf("expected 1 hook, got %d", len(lc.hooks))
	}
	h := lc.hooks[0]

	if err := h.OnStart(context.Background()); err != nil {
		t.Fatalf("OnStart error: %v", err)
	}

	wg.Wait()

	errorMessages := logger.GetErrorMessages()
	if len(errorMessages) == 0 {
		t.Fatalf("expected at least one error message, got none")
	}

	expectedError := "server failed error=bind failed"
	if !strings.Contains(errorMessages[0], expectedError) {
		t.Errorf("unexpected error message: %q", errorMessages[0])
	}
}

func sprintf(format string, v ...any) string {
	return fmt.Sprintf(format, v...)
}

func resetHooksOverrides() func() {
	prevFactory := serverFactory
	prevRun := serverRunner
	prevShutdown := serverShutdown
	return func() {
		serverFactory = prevFactory
		serverRunner = prevRun
		serverShutdown = prevShutdown
	}
}

func TestNewEngine_ServesHTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := newEngine()
	if r == nil {
		t.Fatalf("newEngine returned nil")
	}

	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d, want %d", w.Code, http.StatusOK)
	}
	if got := w.Body.String(); got != "pong" {
		t.Fatalf("unexpected body: got %q, want %q", got, "pong")
	}
}

func TestEngineRunner_CallsStubWithArgs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var called int32
	var gotAddr string
	restore := swapServerRunner(t, func(srv *http.Server) error {
		atomic.AddInt32(&called, 1)
		if srv == nil {
			t.Errorf("expected non-nil *http.Server")
			return errors.New("nil server passed to runner")
		}

		gotAddr = srv.Addr
		return nil
	})
	defer restore()

	srv := serverFactory(":1234", gin.New())
	if err := serverRunner(srv); err != nil {
		t.Fatalf("stub should return nil, got err=%v", err)
	}

	if atomic.LoadInt32(&called) != 1 {
		t.Fatalf("engineRunner stub was not called exactly once")
	}
	if gotAddr != ":1234" {
		t.Fatalf("engineRunner got addr %q, want %q", gotAddr, ":1234")
	}
}

func TestEngineRunner_DefaultCallsRun_AddrInUse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port
	addr := fmt.Sprintf("127.0.0.1:%d", port)

	srv := serverFactory(addr, gin.New())
	if err := serverRunner(srv); err == nil {
		t.Fatalf("expected error (addr in use), got nil")
	}
}

func swapServerRunner(t *testing.T, fn func(*http.Server) error) (restore func()) {
	t.Helper()
	old := serverRunner
	serverRunner = fn
	return func() { serverRunner = old }
}
