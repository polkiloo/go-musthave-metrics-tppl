package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/compression"
	dbcfg "github.com/polkiloo/go-musthave-metrics-tppl/internal/config/db"
	config "github.com/polkiloo/go-musthave-metrics-tppl/internal/config/server"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/db"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/handler"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/logger"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/models"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/server"
	"go.uber.org/fx"
)

func TestMain_WiringIsValid(t *testing.T) {
	err := fx.ValidateApp(
		fx.Provide(func() context.Context { return context.Background() }),
		logger.Module,
		config.Module,
		dbcfg.Module,
		db.Module,
		handler.Module,
		server.Module,
		compression.Module,
		fx.NopLogger,
	)
	if err != nil {
		t.Fatalf("fx wiring validation failed: %v", err)
	}
}

func TestMain_GracefulRun(t *testing.T) {
	done := make(chan struct{})

	go func() {
		main()
		close(done)
	}()

	time.Sleep(150 * time.Millisecond)

	proc, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("FindProcess: %v", err)
	}
	if err := proc.Signal(syscall.SIGINT); err != nil {
		t.Fatalf("send SIGINT: %v", err)
	}

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatalf("main() did not exit in time")
	}
}

func TestMain_PersistRestore(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()
	addr := fmt.Sprintf("localhost:%d", port)

	file := filepath.Join(t.TempDir(), "metrics.json")

	envs := map[string]string{
		"ADDRESS":           addr,
		"STORE_INTERVAL":    "0",
		"FILE_STORAGE_PATH": file,
		"RESTORE":           "true",
	}
	for k, v := range envs {
		os.Setenv(k, v)
		defer os.Unsetenv(k)
	}

	start := func(done chan struct{}) {
		go func() { main(); close(done) }()
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		for {
			req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+addr+"/", nil)
			resp, err := http.DefaultClient.Do(req)
			if err == nil {
				_, _ = io.Copy(io.Discard, resp.Body)
				_ = resp.Body.Close()
				break
			}
			select {
			case <-ctx.Done():
				t.Fatalf("server did not start: %v", ctx.Err())
			default:
				time.Sleep(10 * time.Millisecond)
			}
		}
	}

	done := make(chan struct{})
	start(done)

	val := 12.3
	m, _ := models.NewGaugeMetrics("test", &val)
	body, _ := json.Marshal(m)

	resp, err := http.Post("http://"+addr+"/update", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	func() {
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("update failed: status=%v", resp.StatusCode)
		}
		_, _ = io.Copy(io.Discard, resp.Body)
	}()

	proc, _ := os.FindProcess(os.Getpid())
	_ = proc.Signal(syscall.SIGINT)
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatalf("server did not stop")
	}

	done = make(chan struct{})
	start(done)
	defer func() {
		proc, _ := os.FindProcess(os.Getpid())
		_ = proc.Signal(syscall.SIGINT)
		<-done
	}()

	request := struct {
		ID    string `json:"id"`
		MType string `json:"type"`
	}{ID: "test", MType: string(models.GaugeType)}
	reqBody, _ := json.Marshal(request)

	resp, err = http.Post("http://"+addr+"/value/", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		t.Fatalf("value request failed: %v", err)
	}
	defer resp.Body.Close()

	var out models.Metrics
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK || out.Value == nil || *out.Value != val {
		t.Fatalf("restore failed: status=%d val=%v", resp.StatusCode, out.Value)
	}
}
