package main

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/agent"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/compression"
	agentcfg "github.com/polkiloo/go-musthave-metrics-tppl/internal/config/agent"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/logger"
	"go.uber.org/fx"
)

func TestMain_WiringIsValid(t *testing.T) {
	err := fx.ValidateApp(
		fx.Provide(func() context.Context { return context.Background() }),
		logger.Module,
		agentcfg.Module,
		agent.ModuleCollector,
		agent.ModuleSender,
		agent.ModuleAgent,
		agent.ModuleLoopConfig,
		compression.Module,
		agent.ModuleEncryption,
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
