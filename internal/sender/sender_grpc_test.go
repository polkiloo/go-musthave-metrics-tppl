package sender

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/models"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/proto"
	internaltest "github.com/polkiloo/go-musthave-metrics-tppl/internal/test"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type recordingMetricsServer struct {
	proto.UnimplementedMetricsServer
	received []*proto.Metric
	meta     metadata.MD
}

func (r *recordingMetricsServer) UpdateMetrics(ctx context.Context, req *proto.UpdateMetricsRequest) (*proto.UpdateMetricsResponse, error) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		r.meta = md
	}
	r.received = req.GetMetrics()
	return &proto.UpdateMetricsResponse{}, nil
}

func startGRPCServer(t *testing.T, handler proto.MetricsServer) (addr string, stop func()) {
	t.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := grpc.NewServer()
	proto.RegisterMetricsServer(srv, handler)
	go func() {
		_ = srv.Serve(lis)
	}()
	return lis.Addr().String(), srv.Stop
}

func TestNewGRPCSender_PortRequired(t *testing.T) {
	t.Parallel()

	if _, err := NewGRPCSender("localhost", 0, nil, ""); err == nil {
		t.Fatalf("expected error for missing port")
	}
}

func TestGRPCSender_SendBatchSuccess(t *testing.T) {
	// t.Skip()
	t.Parallel()

	recorder := &recordingMetricsServer{}
	addr, stop := startGRPCServer(t, recorder)
	t.Cleanup(stop)

	l := &internaltest.FakeLogger{}
	sender, err := NewGRPCSender("", addrPort(addr), l, "127.0.0.1")
	if err != nil {
		t.Fatalf("sender init: %v", err)
	}

	gauge := 2.5
	delta := int64(3)
	sender.SendBatch([]*models.Metrics{{ID: "g", MType: models.GaugeType, Value: &gauge}, {ID: "c", MType: models.CounterType, Delta: &delta}})

	if len(recorder.received) != 2 {
		t.Fatalf("expected 2 metrics, got %d", len(recorder.received))
	}
	if recorder.meta.Get("x-real-ip")[0] != "127.0.0.1" {
		t.Fatalf("metadata ip not propagated")
	}
	if len(l.GetInfoMessages()) == 0 {
		t.Fatalf("expected info log")
	}
}

func TestGRPCSender_ConversionError(t *testing.T) {
	t.Parallel()

	recorder := &recordingMetricsServer{}
	addr, stop := startGRPCServer(t, recorder)
	t.Cleanup(stop)

	l := &internaltest.FakeLogger{}
	sender, err := NewGRPCSender("", addrPort(addr), l, "")
	if err != nil {
		t.Fatalf("sender init: %v", err)
	}

	sender.SendBatch([]*models.Metrics{{MType: "bad"}})
	if len(recorder.received) != 0 {
		t.Fatalf("no metrics should be sent on conversion error")
	}
	if len(l.GetErrorMessages()) == 0 {
		t.Fatalf("expected error log")
	}
}

func addrPort(addr string) int {
	_, port, _ := net.SplitHostPort(addr)
	var p int
	_, _ = fmt.Sscan(port, &p)
	return p
}

type mockMetricsClient struct {
	receivedCtx context.Context
	request     *proto.UpdateMetricsRequest
	err         error
	calls       int
}

func (m *mockMetricsClient) UpdateMetrics(ctx context.Context, req *proto.UpdateMetricsRequest, _ ...grpc.CallOption) (*proto.UpdateMetricsResponse, error) {
	m.calls++
	m.receivedCtx = ctx
	m.request = req
	if m.err != nil {
		return nil, m.err
	}
	return &proto.UpdateMetricsResponse{}, nil
}

func TestGRPCSender_NilSafe(t *testing.T) {
	t.Parallel()

	var sender *GRPCSender
	g := 1.0

	sender.SendBatch([]*models.Metrics{{ID: "g", MType: models.GaugeType, Value: &g}})
}

func TestGRPCSender_NoClient(t *testing.T) {
	t.Parallel()

	s := &GRPCSender{}
	g := 1.0

	s.SendBatch([]*models.Metrics{{ID: "g", MType: models.GaugeType, Value: &g}})
}

func TestGRPCSender_EmptyMetrics(t *testing.T) {
	t.Parallel()

	mockClient := &mockMetricsClient{}
	s := &GRPCSender{client: mockClient}

	s.SendBatch(nil)
	if mockClient.calls != 0 {
		t.Fatalf("expected no call for empty metrics, got %d", mockClient.calls)
	}
}

func TestGRPCSender_SendErrorLogs(t *testing.T) {
	t.Parallel()

	g := 1.0
	mockClient := &mockMetricsClient{err: fmt.Errorf("boom")}
	log := &internaltest.FakeLogger{}
	s := &GRPCSender{client: mockClient, log: log}

	s.SendBatch([]*models.Metrics{{ID: "g", MType: models.GaugeType, Value: &g}})

	if mockClient.calls != 1 {
		t.Fatalf("expected one call, got %d", mockClient.calls)
	}
	if len(log.GetErrorMessages()) == 0 {
		t.Fatalf("expected error to be logged")
	}
}
