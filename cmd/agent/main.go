package main

import (
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/agent"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/compression"
	agentcfg "github.com/polkiloo/go-musthave-metrics-tppl/internal/config/agent"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/logger"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		logger.Module,
		agentcfg.Module,
		agent.ModuleCollector,
		agent.ModuleSender,
		agent.ModuleAgent,
		agent.ModuleLoopConfig,
		compression.Module,
	).Run()
}
