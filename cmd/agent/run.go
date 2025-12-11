package main

import (
	"context"
	"errors"
	"time"

	"go.uber.org/fx"
)

func run(ctx context.Context, app *fx.App) error {
	startCtx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	if err := app.Start(startCtx); err != nil {
		return err
	}

	<-ctx.Done()

	stopCtx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	if err := app.Stop(stopCtx); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	return nil
}
