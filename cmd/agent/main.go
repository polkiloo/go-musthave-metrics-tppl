package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/agent"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/buildinfo"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/compression"
	agentcfg "github.com/polkiloo/go-musthave-metrics-tppl/internal/config/agent"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/logger"
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
		agentcfg.Module,
		agent.ModuleCollector,
		agent.ModuleSender,
		agent.ModuleAgent,
		agent.ModuleLoopConfig,
		compression.Module,
		agent.ModuleEncryption,
	)

	if err := run(ctx, app); err != nil {
		log.Printf("server stopped with error: %v", err)
	}
}
