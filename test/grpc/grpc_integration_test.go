package grpc_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/grpcserver"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/proto"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/server"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/service"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/storage"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/trustedsubnet"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func startIntegrationServer(t *testing.T, cfg *server.AppConfig, svc service.MetricServiceInterface) (string, func()) {
	t.Helper()

	interceptor, err := trustedsubnet.NewUnaryInterceptor(cfg)
	if err != nil {
		t.Fatalf("interceptor: %v", err)
	}

	opts := []grpc.ServerOption{}
	if interceptor != nil {
		opts = append(opts, grpc.UnaryInterceptor(interceptor))
	}

	srv := grpc.NewServer(opts...)
	handler, err := grpcserver.ProvideHandler(svc)
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	proto.RegisterMetricsServer(srv, handler)

	lis, err := net.Listen("tcp", net.JoinHostPort(cfg.GRPCHost, "0"))
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	go func() { _ = srv.Serve(lis) }()

	return lis.Addr().String(), srv.Stop
}

func TestGRPCIntegration_AllowsTrustedSubnet(t *testing.T) {
	t.Parallel()

	store := storage.NewMemStorage()
	svc := service.NewMetricService(store)
	addr, stop := startIntegrationServer(t, &server.AppConfig{GRPCHost: "127.0.0.1", TrustedSubnet: "127.0.0.0/8"}, svc)
	t.Cleanup(stop)

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	client := proto.NewMetricsClient(conn)
	gauge := 1.5
	delta := int64(4)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, "x-real-ip", "127.0.0.1")
	_, err = client.UpdateMetrics(ctx, &proto.UpdateMetricsRequest{Metrics: []*proto.Metric{{Id: "g", Type: proto.Metric_GAUGE, Value: gauge}, {Id: "c", Type: proto.Metric_COUNTER, Delta: delta}}})
	if err != nil {
		t.Fatalf("update metrics: %v", err)
	}

	v, err := store.GetGauge("g")
	if err != nil || v != gauge {
		t.Fatalf("gauge not stored: %v %v", v, err)
	}
	c, err := store.GetCounter("c")
	if err != nil || c != delta {
		t.Fatalf("counter not stored: %v %v", c, err)
	}
}

func TestGRPCIntegration_DeniesUntrustedSubnet(t *testing.T) {
	t.Parallel()

	svc := service.NewMetricService(storage.NewMemStorage())
	addr, stop := startIntegrationServer(t, &server.AppConfig{GRPCHost: "127.0.0.1", TrustedSubnet: "127.0.0.0/8"}, svc)
	t.Cleanup(stop)

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	client := proto.NewMetricsClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, "x-real-ip", "10.0.0.5")
	_, err = client.UpdateMetrics(ctx, &proto.UpdateMetricsRequest{})
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("expected permission denied, got %v", status.Code(err))
	}
}
