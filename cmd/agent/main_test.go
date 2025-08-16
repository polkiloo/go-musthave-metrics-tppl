package main

import (
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/agent"
	agentcfg "github.com/polkiloo/go-musthave-metrics-tppl/internal/config/agent"
	"go.uber.org/fx"
)

func TestMain_WiringIsValid(t *testing.T) {
	err := fx.ValidateApp(
		agentcfg.Module,
		fx.Provide(
			agent.ProvideCollector,
			agent.ProvideSender,
			agent.ProvideConfig,
		),
		fx.Invoke(agent.RunAgent),
		fx.NopLogger,
		fx.Supply("http://localhost:8080"),
	)
	if err != nil {
		t.Fatalf("fx validation failed: %v", err)
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
