package agent

import (
	"errors"
	"net"
	"net/http"
	"testing"
)

func TestResolveAgentIP_PrefersNonLoopback(t *testing.T) {
	mw, err := newRealIPMiddleware(func() ([]net.Addr, error) {
		return []net.Addr{
			&net.IPNet{IP: net.ParseIP("127.0.0.1"), Mask: net.CIDRMask(8, 32)},
			&net.IPNet{IP: net.ParseIP("10.0.0.2"), Mask: net.CIDRMask(24, 32)},
		}, nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mw.ip != "10.0.0.2" {
		t.Fatalf("want non-loopback ip, got %q", mw.ip)
	}
}

func TestResolveAgentIP_FallbackToLoopback(t *testing.T) {
	mw, err := newRealIPMiddleware(func() ([]net.Addr, error) {
		return []net.Addr{
			&net.IPNet{IP: net.ParseIP("::1"), Mask: net.CIDRMask(128, 128)},
			&net.IPNet{IP: net.ParseIP("127.0.0.1"), Mask: net.CIDRMask(8, 32)},
		}, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mw.ip != "127.0.0.1" {
		t.Fatalf("want loopback fallback, got %q", mw.ip)
	}
}

func TestResolveAgentIP_Error(t *testing.T) {
	if _, err := newRealIPMiddleware(func() ([]net.Addr, error) { return nil, errors.New("boom") }); err == nil {
		t.Fatalf("expected error")
	}
}

func TestProvideRealIPMiddleware(t *testing.T) {
	mw, err := newRealIPMiddleware(func() ([]net.Addr, error) {
		return []net.Addr{&net.IPNet{IP: net.ParseIP("192.168.1.10"), Mask: net.CIDRMask(24, 32)}}, nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	mw.Middleware()(req)
	if req.Header.Get("X-Real-IP") != "192.168.1.10" {
		t.Fatalf("expected header to be set, got %q", req.Header.Get("X-Real-IP"))
	}
}
