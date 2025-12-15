package grpcserver

import (
	"context"
	"testing"
	"time"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/logger"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/server"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/service"
	internaltest "github.com/polkiloo/go-musthave-metrics-tppl/internal/test"
	"go.uber.org/fx"
)

type testLifecycle struct {
	hooks []fx.Hook
}

func (t *testLifecycle) Append(h fx.Hook) {
	t.hooks = append(t.hooks, h)
}

func (t *testLifecycle) Start(ctx context.Context) error {
	for i := 0; i < len(t.hooks); i++ {
		if t.hooks[i].OnStart == nil {
			continue
		}
		if err := t.hooks[i].OnStart(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (t *testLifecycle) Stop(ctx context.Context) error {
	for i := len(t.hooks) - 1; i >= 0; i-- {
		if t.hooks[i].OnStop == nil {
			continue
		}
		if err := t.hooks[i].OnStop(ctx); err != nil {
			return err
		}
	}
	return nil
}

func TestProvideHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil service", func(t *testing.T) {
		t.Parallel()
		if _, err := ProvideHandler(nil); err == nil {
			t.Fatalf("expected error for nil service")
		}
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		handler, err := ProvideHandler(&internaltest.FakeMetricService{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		srv, ok := handler.(*Server)
		if !ok || srv.svc == nil {
			t.Fatalf("handler not initialized: %#v", handler)
		}
	})
}

func TestRun(t *testing.T) {
	t.Parallel()

	t.Run("nil config", func(t *testing.T) {
		t.Parallel()
		if err := Run(&testLifecycle{}, nil, &Server{svc: &internaltest.FakeMetricService{}}, &internaltest.FakeLogger{}); err == nil {
			t.Fatalf("expected error for nil config")
		}
	})

	t.Run("invalid subnet", func(t *testing.T) {
		t.Parallel()
		cfg := &server.AppConfig{TrustedSubnet: "invalid"}
		err := Run(&testLifecycle{}, cfg, &Server{svc: &internaltest.FakeMetricService{}}, &internaltest.FakeLogger{})
		if err == nil {
			t.Fatalf("expected subnet parse error")
		}
	})

	t.Run("listen error", func(t *testing.T) {
		t.Parallel()
		cfg := &server.AppConfig{GRPCHost: "bad host", GRPCPort: 1234}
		err := Run(&testLifecycle{}, cfg, &Server{svc: &internaltest.FakeMetricService{}}, &internaltest.FakeLogger{})
		if err == nil {
			t.Fatalf("expected listen error")
		}
	})

	t.Run("start and stop", func(t *testing.T) {
		t.Parallel()

		lc := &testLifecycle{}
		cfg := &server.AppConfig{GRPCHost: "127.0.0.1", GRPCPort: 0}
		logger := &internaltest.FakeLogger{}

		if err := Run(lc, cfg, &Server{svc: &internaltest.FakeMetricService{}}, logger); err != nil {
			t.Fatalf("run returned error: %v", err)
		}

		if len(lc.hooks) == 0 {
			t.Fatalf("expected hooks to be registered")
		}

		startCtx, startCancel := context.WithTimeout(context.Background(), time.Second)
		defer startCancel()
		if err := lc.Start(startCtx); err != nil {
			t.Fatalf("start error: %v", err)
		}

		stopCtx, stopCancel := context.WithTimeout(context.Background(), time.Second)
		defer stopCancel()
		if err := lc.Stop(stopCtx); err != nil {
			t.Fatalf("stop error: %v", err)
		}

		if errMsg := logger.GetLastErrorMessage(); errMsg != "" {
			t.Fatalf("unexpected error log: %s", errMsg)
		}
	})

	t.Run("interceptor and cancel stop", func(t *testing.T) {
		t.Parallel()

		lc := &testLifecycle{}
		cfg := &server.AppConfig{GRPCHost: "127.0.0.1", GRPCPort: 0, TrustedSubnet: "127.0.0.0/8"}
		logger := &internaltest.FakeLogger{}

		if err := Run(lc, cfg, &Server{svc: &internaltest.FakeMetricService{}}, logger); err != nil {
			t.Fatalf("run returned error: %v", err)
		}

		startCtx, startCancel := context.WithTimeout(context.Background(), time.Second)
		defer startCancel()
		if err := lc.Start(startCtx); err != nil {
			t.Fatalf("start error: %v", err)
		}

		stopCtx, stopCancel := context.WithCancel(context.Background())
		stopCancel()
		if err := lc.Stop(stopCtx); err != nil {
			t.Fatalf("stop error: %v", err)
		}

		if errMsg := logger.GetLastErrorMessage(); errMsg != "" {
			t.Fatalf("unexpected error log: %s", errMsg)
		}
	})

	t.Run("module wiring", func(t *testing.T) {
		t.Parallel()

		lg := &internaltest.FakeLogger{}
		app := fx.New(
			Module,
			fx.Provide(func() *server.AppConfig {
				return &server.AppConfig{GRPCHost: "127.0.0.1", GRPCPort: 0}
			}),
			fx.Supply(fx.Annotate(lg, fx.As(new(logger.Logger)))),
			fx.Supply(fx.Annotate(&internaltest.FakeMetricService{}, fx.As(new(service.MetricServiceInterface)))),
			fx.NopLogger,
		)

		startCtx, startCancel := context.WithTimeout(context.Background(), time.Second)
		defer startCancel()
		if err := app.Start(startCtx); err != nil {
			t.Fatalf("app start error: %v", err)
		}

		stopCtx, stopCancel := context.WithTimeout(context.Background(), time.Second)
		defer stopCancel()
		if err := app.Stop(stopCtx); err != nil {
			t.Fatalf("app stop error: %v", err)
		}

		if errMsg := lg.GetLastErrorMessage(); errMsg != "" {
			t.Fatalf("unexpected error log: %s", errMsg)
		}
	})
}
