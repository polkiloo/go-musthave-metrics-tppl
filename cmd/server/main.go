package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/audit"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/buildinfo"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/compression"
	dbcfg "github.com/polkiloo/go-musthave-metrics-tppl/internal/config/db"
	config "github.com/polkiloo/go-musthave-metrics-tppl/internal/config/server"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/db"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/handler"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/logger"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/server"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/service"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/sign"
	"go.uber.org/fx"
)

var buildVersion = buildinfo.InfoData().Version
var buildDate = buildinfo.InfoData().Date
var buildCommit = buildinfo.InfoData().Commit

func main() {

	buildinfo.Print(os.Stdout, buildinfo.Info{Version: buildVersion, Date: buildDate, Commit: buildCommit})

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	app := fx.New(
		fx.Provide(func() context.Context { return ctx }),
		logger.Module,
		config.Module,
		dbcfg.Module,
		db.Module,
		service.Module,
		handler.Module,
		server.Module,
		compression.Module,
		sign.Module,
		audit.Module,
	)

	if err := run(ctx, app); err != nil {
		log.Printf("server stopped with error: %v", err)
	}
}
