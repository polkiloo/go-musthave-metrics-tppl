package trustedsubnet

import (
	"context"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const realIPMetadataKey = "x-real-ip"

// NewUnaryInterceptor validates x-real-ip metadata against trusted subnet when configured.
func NewUnaryInterceptor(cfg *server.AppConfig) (grpc.UnaryServerInterceptor, error) {
	validator, err := newValidator(cfg)
	if err != nil {
		return nil, err
	}

	if validator == nil {
		return nil, nil
	}
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.PermissionDenied, "missing metadata")
		}
		vals := md.Get(realIPMetadataKey)
		if len(vals) == 0 {
			return nil, status.Error(codes.PermissionDenied, "missing x-real-ip")
		}

		if !validator.contains(vals[0]) {
			return nil, status.Error(codes.PermissionDenied, "ip not allowed")
		}
		return handler(ctx, req)
	}, nil
}
