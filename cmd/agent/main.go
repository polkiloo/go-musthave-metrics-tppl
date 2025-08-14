package main

import (
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/agent"
	config "github.com/polkiloo/go-musthave-metrics-tppl/internal/config/agent"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		config.Module,
		fx.Provide(
			agent.ProvideCollector,
			agent.ProvideSender,
			agent.ProvideConfig,
		),
		fx.Invoke(agent.RunAgent),
	).Run()
}
