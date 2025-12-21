package server

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/handler"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/logger"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/sign"
	"go.uber.org/fx"
)

// AppConfig describes configuration for the HTTP server application.
type AppConfig struct {
	Host            string
	Port            int
	StoreInterval   int
	FileStoragePath string
	Restore         bool
	SignKey         sign.SignKey
	AuditFile       string
	AuditURL        string
	CryptoKeyPath   string
	TrustedSubnet   string
	GRPCHost        string
	GRPCPort        int
}

const (
	// DefaultAppHost is the hostname used when no host is provided.
	DefaultAppHost = "localhost"
	// DefaultAppPort is the default port the server listens on.
	DefaultAppPort = 8080
	// DefaultStoreInterval controls how often metrics are flushed to disk.
	DefaultStoreInterval = 300
	// DefaultFileStoragePath is the default path for the metrics snapshot file.
	DefaultFileStoragePath = "/tmp/metrics-db.json"
	// DefaultRestore indicates whether the service loads state on start by default.
	DefaultRestore = true
	// DefaultCryptoKeyPath is the default path for key used for encryption metrics.
	DefaultCryptoKeyPath = ""
	// DefaultGRPCHost is the default gRPC server host.
	DefaultGRPCHost = "localhost"
	// DefaultGRPCPort is the default gRPC server port.
	DefaultGRPCPort = 3200
)

// DefaultAppConfig provides baseline server configuration values.
var DefaultAppConfig = AppConfig{
	Host:            DefaultAppHost,
	Port:            DefaultAppPort,
	StoreInterval:   DefaultStoreInterval,
	FileStoragePath: DefaultFileStoragePath,
	Restore:         DefaultRestore,
	GRPCHost:        DefaultGRPCHost,
	GRPCPort:        DefaultGRPCPort,
}

func newEngine() *gin.Engine {
	return gin.Default()
}

var (
	serverFactory = func(addr string, handler http.Handler) *http.Server {
		return &http.Server{Addr: addr, Handler: handler}
	}
	serverRunner = func(srv *http.Server) error {
		return srv.ListenAndServe()
	}
	serverShutdown = func(ctx context.Context, srv *http.Server) error {
		return srv.Shutdown(ctx)
	}
)

func run(lc fx.Lifecycle, r *gin.Engine, cfg *AppConfig, l logger.Logger, h *handler.GinHandler) {
	var (
		stopSaver chan struct{}
		srv       *http.Server
	)
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
				srv = serverFactory(addr, r)

				l.WriteInfo("server listening", "addr", "http://"+addr)

				if err := serverRunner(srv); err != nil && !errors.Is(err, http.ErrServerClosed) {
					l.WriteError("server failed", "error", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if srv != nil {
				shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
				defer cancel()

				if err := serverShutdown(shutdownCtx, srv); err != nil &&
					!errors.Is(err, context.Canceled) && !errors.Is(err, http.ErrServerClosed) {
					l.WriteError("server shutdown failed", "error", err)
				}
			}
			if cfg.StoreInterval > 0 {
				if stopSaver != nil {
					close(stopSaver)
				}
			}
			if err := h.Service().SaveFile(cfg.FileStoragePath); err != nil {
				l.WriteError("save failed", "error", err)
			}
			return l.Sync()
		},
	})
}

// Module wires the HTTP server lifecycle hooks into the fx application.
var Module = fx.Module(
	"server",
	fx.Provide(newEngine),
	fx.Invoke(run),
)
