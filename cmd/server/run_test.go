package main

import (
	"context"
	"testing"

	"go.uber.org/fx"
)

func TestRun_StartsWithDedicatedContext(t *testing.T) {
	var (
		started     bool
		startCtxErr error
		stopped     bool
		stopCtxErr  error
	)

	app := fx.New(
		fx.Invoke(func(lc fx.Lifecycle) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					started = true
					startCtxErr = ctx.Err()
					return nil
				},
				OnStop: func(ctx context.Context) error {
					stopped = true
					stopCtxErr = ctx.Err()
					return nil
				},
			})
		}),
	)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	run(ctx, app)

	if !started {
		t.Fatalf("app did not start")
	}
	if startCtxErr != nil {
		t.Fatalf("start context error: %v", startCtxErr)
	}
	if !stopped {
		t.Fatalf("app did not stop")
	}
	if stopCtxErr != nil {
		t.Fatalf("stop context error: %v", stopCtxErr)
	}
}
