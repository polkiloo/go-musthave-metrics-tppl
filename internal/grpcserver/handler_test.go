package grpcserver

import (
	"errors"
	"testing"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/models"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/proto"
	internaltest "github.com/polkiloo/go-musthave-metrics-tppl/internal/test"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type errService struct{}

func (errService) ProcessUpdate(*models.Metrics) error { return nil }
func (errService) ProcessGetValue(string, models.MetricType) (*models.Metrics, error) {
	return nil, nil
}
func (errService) SaveFile(string) error { return nil }
func (errService) LoadFile(string) error { return nil }
func (errService) ProcessUpdates([]models.Metrics) error {
	return errors.New("update error")
}

func TestServer_UpdateMetrics(t *testing.T) {
	t.Parallel()

	gauge := 1.0
	req := &proto.UpdateMetricsRequest{Metrics: []*proto.Metric{{Id: "g", Type: proto.Metric_GAUGE, Value: gauge}}}

	t.Run("nil service", func(t *testing.T) {
		t.Parallel()
		srv := &Server{}
		_, err := srv.UpdateMetrics(t.Context(), req)
		if status.Code(err) != codes.Internal {
			t.Fatalf("expected internal code, got %v", status.Code(err))
		}
	})

	t.Run("conversion error", func(t *testing.T) {
		t.Parallel()
		srv := &Server{svc: &internaltest.FakeMetricService{}}
		badReq := &proto.UpdateMetricsRequest{Metrics: []*proto.Metric{{Id: "bad", Type: proto.Metric_MType(-1)}}}
		_, err := srv.UpdateMetrics(t.Context(), badReq)
		if status.Code(err) != codes.InvalidArgument {
			t.Fatalf("expected invalid argument, got %v", status.Code(err))
		}
	})

	t.Run("service error", func(t *testing.T) {
		t.Parallel()
		srv := &Server{svc: errService{}}
		_, err := srv.UpdateMetrics(t.Context(), req)
		if status.Code(err) != codes.Internal {
			t.Fatalf("expected internal code, got %v", status.Code(err))
		}
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		svc := &internaltest.FakeMetricService{}
		srv := &Server{svc: svc}
		if _, err := srv.UpdateMetrics(t.Context(), req); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if svc.Metric.ID != "g" || svc.Metric.Value == nil || *svc.Metric.Value != gauge {
			t.Fatalf("metric not processed: %+v", svc.Metric)
		}
	})
}
