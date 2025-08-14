package server

import (
	"context"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
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
	logPrintf    = func(format string, v ...any) { log.Printf(format, v...) }
	logFatalf    = func(format string, v ...any) { log.Fatalf(format, v...) }
)

func run(lc fx.Lifecycle, r *gin.Engine, cfg *AppConfig) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				addr := cfg.Host + ":" + strconv.Itoa(cfg.Port)
				logPrintf("Server listening on http://%s", addr)
				if err := engineRunner(r, addr); err != nil {
					logFatalf("Server failed: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error { return nil },
	})
}

var Module = fx.Module(
	"server",
	fx.Provide(newEngine),
	fx.Invoke(run),
)
