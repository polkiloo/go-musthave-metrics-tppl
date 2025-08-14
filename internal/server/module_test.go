package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
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
	var logged string

	logPrintf = func(format string, v ...any) {
		logged = sprintf(format, v...)
	}

	engineRunner = func(r *gin.Engine, addr string) error {
		defer wg.Done()
		gotAddr = addr
		return nil
	}

	run(lc, engine, cfg)
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
	if !strings.Contains(logged, "Server listening on http://127.0.0.1:18080") {
		t.Errorf("unexpected log: %q", logged)
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

	var fatalMsg string

	engineRunner = func(r *gin.Engine, addr string) error {
		defer wg.Done()
		return errors.New("bind failed")
	}
	logFatalf = func(format string, v ...any) {
		fatalMsg = sprintf(format, v...)
	}

	run(lc, engine, cfg)
	if len(lc.hooks) != 1 {
		t.Fatalf("expected 1 hook, got %d", len(lc.hooks))
	}
	h := lc.hooks[0]

	if err := h.OnStart(context.Background()); err != nil {
		t.Fatalf("OnStart error: %v", err)
	}

	wg.Wait()

	if !strings.Contains(fatalMsg, "Server failed: bind failed") {
		t.Errorf("unexpected fatal message: %q", fatalMsg)
	}
}

func sprintf(format string, v ...any) string {
	return fmt.Sprintf(format, v...)
}

func resetHooksOverrides() func() {
	prevRun := engineRunner
	prevPrint := logPrintf
	prevFatal := logFatalf
	return func() {
		engineRunner = prevRun
		logPrintf = prevPrint
		logFatalf = prevFatal
	}
}

func swapEngineRunner(t *testing.T, fn func(*gin.Engine, string) error) (restore func()) {
	t.Helper()
	old := engineRunner
	engineRunner = fn
	return func() { engineRunner = old }
}

func swapLogPrintf(t *testing.T, fn func(string, ...any)) (restore func()) {
	t.Helper()
	old := logPrintf
	logPrintf = fn
	return func() { logPrintf = old }
}

func swapLogFatalf(t *testing.T, fn func(string, ...any)) (restore func()) {
	t.Helper()
	old := logFatalf
	logFatalf = fn
	return func() { logFatalf = old }
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
	restore := swapEngineRunner(t, func(e *gin.Engine, addr string) error {
		atomic.AddInt32(&called, 1)
		if e == nil {
			t.Errorf("expected non-nil *gin.Engine")
		}
		gotAddr = addr
		return nil
	})
	defer restore()

	r := gin.New()
	if err := engineRunner(r, ":1234"); err != nil {
		t.Fatalf("stub should return nil, got err=%v", err)
	}

	if atomic.LoadInt32(&called) != 1 {
		t.Fatalf("engineRunner stub was not called exactly once")
	}
	if gotAddr != ":1234" {
		t.Fatalf("engineRunner got addr %q, want %q", gotAddr, ":1234")
	}
}

func TestLogPrintf_OverrideAndCall(t *testing.T) {
	var called int32
	var gotFmt string
	var gotArgs []any

	restore := swapLogPrintf(t, func(format string, v ...any) {
		atomic.AddInt32(&called, 1)
		gotFmt = format
		gotArgs = v
	})
	defer restore()

	logPrintf("hello %s %d", "world", 42)

	if atomic.LoadInt32(&called) != 1 {
		t.Fatalf("logPrintf stub not called")
	}
	if gotFmt != "hello %s %d" {
		t.Fatalf("format mismatch: got %q", gotFmt)
	}
	if len(gotArgs) != 2 || gotArgs[0] != "world" || gotArgs[1] != 42 {
		t.Fatalf("args mismatch: %+v", gotArgs)
	}
}

func TestLogFatalf_OverrideAndCall_NoExit(t *testing.T) {
	var called int32
	var gotFmt string
	var gotArgs []any

	restore := swapLogFatalf(t, func(format string, v ...any) {
		atomic.AddInt32(&called, 1)
		gotFmt = format
		gotArgs = v
	})
	defer restore()

	logFatalf("fatal %s", "oops")

	if atomic.LoadInt32(&called) != 1 {
		t.Fatalf("logFatalf stub not called")
	}
	if gotFmt != "fatal %s" || len(gotArgs) != 1 || gotArgs[0] != "oops" {
		t.Fatalf("logFatalf args mismatch: fmt=%q args=%v", gotFmt, gotArgs)
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

	r := gin.New()
	if err := engineRunner(r, addr); err == nil {
		t.Fatalf("expected error (addr in use), got nil")
	}
}

type fakeLC struct{ hook fx.Hook }

func (f *fakeLC) Append(h fx.Hook) { f.hook = h }

func TestRun_OnStart_UsesRunnerAndLogs_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var runnerCalled int32
	var gotAddr string

	done := make(chan struct{})

	restoreRunner := swapEngineRunner(t, func(e *gin.Engine, addr string) error {
		atomic.AddInt32(&runnerCalled, 1)
		gotAddr = addr
		close(done)
		return nil
	})
	defer restoreRunner()

	var printfCalled int32
	restorePrintf := swapLogPrintf(t, func(format string, v ...any) {
		atomic.AddInt32(&printfCalled, 1)
	})
	defer restorePrintf()

	var fatalCalled int32
	restoreFatal := swapLogFatalf(t, func(format string, v ...any) {
		atomic.AddInt32(&fatalCalled, 1)
	})
	defer restoreFatal()

	lc := &fakeLC{}
	r := gin.New()
	cfg := &AppConfig{Host: "127.0.0.1", Port: 54321}

	run(lc, r, cfg)

	if err := lc.hook.OnStart(context.Background()); err != nil {
		t.Fatalf("OnStart: %v", err)
	}

	select {
	case <-done:

	case <-time.After(1 * time.Second):
		t.Fatalf("engineRunner was not called")
	}

	if atomic.LoadInt32(&runnerCalled) != 1 {
		t.Fatalf("runner not called exactly once")
	}
	if gotAddr != "127.0.0.1:54321" {
		t.Fatalf("runner addr = %q, want %q", gotAddr, "127.0.0.1:54321")
	}
	if atomic.LoadInt32(&printfCalled) == 0 {
		t.Fatalf("logPrintf was not called")
	}
	if atomic.LoadInt32(&fatalCalled) != 0 {
		t.Fatalf("logFatalf should not be called on success")
	}
}

func TestRun_OnStart_RunnerErrorCallsFatal(t *testing.T) {
	gin.SetMode(gin.TestMode)

	restoreRunner := swapEngineRunner(t, func(e *gin.Engine, addr string) error {
		return fmt.Errorf("boom")
	})
	defer restoreRunner()

	var fatalCalled int32
	restoreFatal := swapLogFatalf(t, func(format string, v ...any) {
		atomic.AddInt32(&fatalCalled, 1)
	})
	defer restoreFatal()

	lc := &fakeLC{}
	r := gin.New()
	cfg := &AppConfig{Host: "127.0.0.1", Port: 1}

	run(lc, r, cfg)

	if err := lc.hook.OnStart(context.Background()); err != nil {
		t.Fatalf("OnStart: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	if atomic.LoadInt32(&fatalCalled) == 0 {
		t.Fatalf("logFatalf was not called on runner error")
	}
}

func TestDefaultLogFatalf_ExitsProcess(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-test.run=TestHelperProcess_LogFatalf")
	cmd.Env = append(os.Environ(), "HELPER_LOGFATALF=1")
	err := cmd.Run()
	if err == nil {
		t.Fatalf("expected child process to exit with error (os.Exit), got nil")
	}
	if exitErr, ok := err.(*exec.ExitError); !ok || exitErr.Success() {
		t.Fatalf("expected non-zero exit status, got: %v", err)
	}
}

func TestHelperProcess_LogFatalf(t *testing.T) {
	if os.Getenv("HELPER_LOGFATALF") != "1" {
		t.Skip("helper process only")
	}
	logFatalf("fatal %s", "boom")
	t.Fatalf("logFatalf did not exit the process")
}
