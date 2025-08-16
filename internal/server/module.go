package server

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/logger"
	"go.uber.org/fx"
)

type AppConfig struct {
	Host string
	Port int
}

const (
	DefaultAppHost = "localhost"
	DefaultAppPort = 8080
)

var DefaultAppConfig = AppConfig{
	Host: DefaultAppHost,
	Port: DefaultAppPort,
}

func newEngine() *gin.Engine {
	return gin.Default()
}

var (
	engineRunner = func(r *gin.Engine, addr string) error { return r.Run(addr) }
)

func run(lc fx.Lifecycle, r *gin.Engine, cfg *AppConfig, l logger.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				addr := cfg.Host + ":" + strconv.Itoa(cfg.Port)
				l.WriteInfo("server listening", "addr", "http://"+addr)

				if err := engineRunner(r, addr); err != nil {
					l.WriteError("server failed", "error", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return l.Sync()
		},
	})
}

var Module = fx.Module(
	"server",
	fx.Provide(newEngine),
	fx.Invoke(run),
)
