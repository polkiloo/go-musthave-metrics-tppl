package main

import (
	config "github.com/polkiloo/go-musthave-metrics-tppl/internal/config/server"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/handler"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/server"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		config.Module,
		handler.Module,
		server.Module,
	).Run()
}
