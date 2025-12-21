package trustedsubnet

import (
	"context"
	"net"
	"testing"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestNewUnaryInterceptor_ConfigMissing(t *testing.T) {
	t.Parallel()

	interceptor, err := NewUnaryInterceptor(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if interceptor != nil {
		t.Fatalf("expected nil interceptor for missing config")
	}
}

func TestNewUnaryInterceptor_InvalidCIDR(t *testing.T) {
	t.Parallel()

	_, err := NewUnaryInterceptor(&server.AppConfig{TrustedSubnet: "bad subnet"})
	if err == nil {
		t.Fatalf("expected error for invalid cidr")
	}
}

func TestUnaryInterceptor_AccessControl(t *testing.T) {
	t.Parallel()

	interceptor, err := NewUnaryInterceptor(&server.AppConfig{TrustedSubnet: "192.168.1.0/24"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	handlerCalled := false
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		handlerCalled = true
		return "ok", nil
	}

	tests := []struct {
		name string
		ctx  context.Context
		code codes.Code
		call bool
	}{
		{name: "no metadata", ctx: context.Background(), code: codes.PermissionDenied},
		{name: "missing header", ctx: metadata.NewIncomingContext(context.Background(), metadata.MD{}), code: codes.PermissionDenied},
		{name: "ip not allowed", ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs(realIPMetadataKey, "10.0.0.1")), code: codes.PermissionDenied},
		{name: "allowed", ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs(realIPMetadataKey, "192.168.1.10")), code: codes.OK, call: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			handlerCalled = false
			_, err := interceptor(tt.ctx, nil, &grpc.UnaryServerInfo{FullMethod: "test"}, handler)
			if status.Code(err) != tt.code {
				t.Fatalf("expected code %v, got %v", tt.code, status.Code(err))
			}
			if handlerCalled != tt.call {
				t.Fatalf("handler call mismatch: expected %v", tt.call)
			}
		})
	}

	mdCtx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(realIPMetadataKey, "192.168.1.10"))
	_, err = interceptor(mdCtx, nil, &grpc.UnaryServerInfo{FullMethod: "test"}, handler)
	if err != nil {
		t.Fatalf("unexpected error on allowed request: %v", err)
	}
}

func TestNewUnaryInterceptor_EmptyTrustedSubnet(t *testing.T) {
	t.Parallel()

	interceptor, err := NewUnaryInterceptor(&server.AppConfig{TrustedSubnet: ""})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if interceptor != nil {
		t.Fatalf("expected nil interceptor for empty subnet")
	}
}

func TestTrustedSubnetCIDRParsing(t *testing.T) {
	t.Parallel()

	_, cidr, err := net.ParseCIDR("192.168.1.0/24")
	if err != nil || cidr == nil {
		t.Fatalf("expected cidr to parse in test")
	}
}
