package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/compression"
	dbcfg "github.com/polkiloo/go-musthave-metrics-tppl/internal/config/db"
	config "github.com/polkiloo/go-musthave-metrics-tppl/internal/config/server"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/db"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/handler"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/logger"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/server"
	"go.uber.org/fx"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	app := fx.New(
		fx.Provide(func() context.Context { return ctx }),
		logger.Module,
		config.Module,
		dbcfg.Module,
		db.Module,
		handler.Module,
		server.Module,
		compression.Module,
	)

	run(ctx, app)
}
