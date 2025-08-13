package main

import (
	"log"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/agent"
	"go.uber.org/fx"
)

func main() {
	args, err := parseFlags()
	if err != nil {
		log.Fatalf("flags: %v", err)
	}
	app := fx.New(
		fx.Provide(
			func() agent.AppConfig { return args.ToAppConfig() },
			agent.ProvideCollector,
			agent.ProvideSender,
			agent.ProvideConfig,
		),
		fx.Invoke(agent.RunAgent),
	)
	app.Run()
}
