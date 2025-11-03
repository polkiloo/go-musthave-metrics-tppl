package server

import (
	"context"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/handler"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/logger"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/sign"
	"go.uber.org/fx"
)

type AppConfig struct {
	Host            string
	Port            int
	StoreInterval   int
	FileStoragePath string
	Restore         bool
	SignKey         sign.SignKey
}

const (
	DefaultAppHost         = "localhost"
	DefaultAppPort         = 8080
	DefaultStoreInterval   = 300
	DefaultFileStoragePath = "/tmp/metrics-db.json"
	DefaultRestore         = true
)

var DefaultAppConfig = AppConfig{
	Host:            DefaultAppHost,
	Port:            DefaultAppPort,
	StoreInterval:   DefaultStoreInterval,
	FileStoragePath: DefaultFileStoragePath,
	Restore:         DefaultRestore,
}

func newEngine() *gin.Engine {
	return gin.Default()
}

var (
	engineRunner = func(r *gin.Engine, addr string) error { return r.Run(addr) }
)

func run(lc fx.Lifecycle, r *gin.Engine, cfg *AppConfig, l logger.Logger, h *handler.GinHandler) {
	var stopSaver chan struct{}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if cfg.Restore {
				if err := h.Service().LoadFile(cfg.FileStoragePath); err != nil {
					l.WriteError("restore failed", "error", err)
				}
			}

			if cfg.StoreInterval > 0 {
				stopSaver = make(chan struct{})
				go func() {
					ticker := time.NewTicker(time.Duration(cfg.StoreInterval) * time.Second)
					defer ticker.Stop()
					for {
						select {
						case <-ticker.C:
							if err := h.Service().SaveFile(cfg.FileStoragePath); err != nil {
								l.WriteError("save failed", "error", err)
							}
						case <-stopSaver:
							return
						}
					}
				}()
			} else {
				h.SetAfterUpdateHook(func() {
					if err := h.Service().SaveFile(cfg.FileStoragePath); err != nil {
						l.WriteError("save failed", "error", err)
					}
				})
			}
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
			if cfg.StoreInterval > 0 {
				close(stopSaver)
				if err := h.Service().SaveFile(cfg.FileStoragePath); err != nil {
					l.WriteError("save failed", "error", err)
				}
			}
			return l.Sync()
		},
	})
}

var Module = fx.Module(
	"server",
	fx.Provide(newEngine),
	fx.Invoke(run),
)
