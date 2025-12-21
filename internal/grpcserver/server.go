package grpcserver

import (
	"context"
	"fmt"
	"net"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/logger"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/proto"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/server"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/service"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/trustedsubnet"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

// Run starts gRPC server for metrics handling.
func Run(lc fx.Lifecycle, cfg *server.AppConfig, handler proto.MetricsServer, l logger.Logger) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}
	address := fmt.Sprintf("%s:%d", cfg.GRPCHost, cfg.GRPCPort)

	opts := make([]grpc.ServerOption, 0)
	interceptor, err := trustedsubnet.NewUnaryInterceptor(cfg)
	if err != nil {
		return err
	}
	if interceptor != nil {
		opts = append(opts, grpc.UnaryInterceptor(interceptor))
	}

	srv := grpc.NewServer(opts...)
	proto.RegisterMetricsServer(srv, handler)

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("listen grpc: %w", err)
	}

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				l.WriteInfo("grpc server listening", "addr", address)
				if err := srv.Serve(listener); err != nil {
					l.WriteError("grpc server failed", "error", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			stopped := make(chan struct{})
			go func() {
				srv.GracefulStop()
				close(stopped)
			}()

			select {
			case <-ctx.Done():
				srv.Stop()
			case <-stopped:
			}
			return nil
		},
	})

	return nil
}

// ProvideHandler builds MetricsServer implementation.
func ProvideHandler(svc service.MetricServiceInterface) (proto.MetricsServer, error) {
	if svc == nil {
		return nil, fmt.Errorf("metric service is nil")
	}
	return &Server{svc: svc}, nil
}

// Module registers gRPC server runtime to fx app.
var Module = fx.Module("grpcserver",
	fx.Provide(ProvideHandler),
	fx.Invoke(func(lc fx.Lifecycle, cfg *server.AppConfig, h proto.MetricsServer, l logger.Logger) error {
		return Run(lc, cfg, h, l)
	}),
)
