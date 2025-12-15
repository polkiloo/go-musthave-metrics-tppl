package main

import (
	"context"
	"errors"
	"time"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/lifecycle"
	"go.uber.org/fx"
)

func run(ctx context.Context, app *fx.App) error {
	startCtx, cancel := lifecycle.BoundedContext(ctx, 60*time.Second)
	defer cancel()

	if err := app.Start(startCtx); err != nil {
		return err
	}

	<-ctx.Done()

	stopCtx, cancel := lifecycle.BoundedContext(ctx, 60*time.Second)
	defer cancel()

	if err := app.Stop(stopCtx); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	return nil
}
