package grpcserver

import (
	"context"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/proto"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server handles gRPC metrics requests.
type Server struct {
	proto.UnimplementedMetricsServer
	svc service.MetricServiceInterface
}

// UpdateMetrics processes batch metrics update.
func (s *Server) UpdateMetrics(ctx context.Context, req *proto.UpdateMetricsRequest) (*proto.UpdateMetricsResponse, error) {
	if s == nil || s.svc == nil {
		return nil, status.Error(codes.Internal, "metric service is unavailable")
	}
	models, err := proto.MetricsToModels(req.GetMetrics())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "convert metrics: %v", err)
	}
	if err := s.svc.ProcessUpdates(models); err != nil {
		return nil, status.Errorf(codes.Internal, "process metrics: %v", err)
	}
	return &proto.UpdateMetricsResponse{}, nil
}

var _ proto.MetricsServer = (*Server)(nil)
