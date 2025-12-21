package sender

import (
	"context"
	"fmt"
	"time"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/logger"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/models"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// GRPCSender sends metrics batches via gRPC.
type GRPCSender struct {
	client proto.MetricsClient
	log    logger.Logger
	ip     string
}

// NewGRPCSender creates a sender connected to given host and port.
func NewGRPCSender(host string, port int, l logger.Logger, ip string) (*GRPCSender, error) {
	if host == "" {
		host = "127.0.0.1"
	}
	if port == 0 {
		return nil, fmt.Errorf("grpc port is not set")
	}
	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := grpc.NewClient(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("dial grpc: %w", err)
	}
	conn.Connect()

	return &GRPCSender{client: proto.NewMetricsClient(conn), log: l, ip: ip}, nil
}

// Send implements SenderInterface.
func (s *GRPCSender) Send(metrics []*models.Metrics) { s.SendBatch(metrics) }

// SendBatch sends all metrics in one UpdateMetrics call.
func (s *GRPCSender) SendBatch(metrics []*models.Metrics) {
	s.sendWithContext(context.Background(), metrics)
}

// SendWithContext sends metrics batch with provided context.
func (s *GRPCSender) SendWithContext(ctx context.Context, metrics []*models.Metrics) {
	s.sendWithContext(ctx, metrics)
}

func (s *GRPCSender) sendWithContext(ctx context.Context, metrics []*models.Metrics) {
	if s == nil || s.client == nil {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if len(metrics) == 0 {
		return
	}

	protoMetrics, err := proto.MetricsFromModels(metrics)
	if err != nil {
		s.log.WriteError("convert metrics", "error", err)
		return
	}

	req := &proto.UpdateMetricsRequest{Metrics: protoMetrics}
	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if s.ip != "" {
		reqCtx = metadata.AppendToOutgoingContext(reqCtx, "x-real-ip", s.ip)
	}

	if _, err := s.client.UpdateMetrics(reqCtx, req); err != nil {
		if s.log != nil {
			s.log.WriteError("grpc send failed", "error", err)
		}
		return
	}
	if s.log != nil {
		s.log.WriteInfo("grpc metrics sent", "count", len(protoMetrics))
	}
}

var _ SenderInterface = (*GRPCSender)(nil)
var _ ContextualSender = (*GRPCSender)(nil)
