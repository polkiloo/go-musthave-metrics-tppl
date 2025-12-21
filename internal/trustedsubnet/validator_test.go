package trustedsubnet

import (
	"testing"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/server"
)

func TestNewValidator(t *testing.T) {
	vd, err := newValidator(&server.AppConfig{TrustedSubnet: "10.0.0.0/24"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vd == nil {
		t.Fatalf("expected validator instance")
	}
	if !vd.contains("10.0.0.1") {
		t.Fatalf("expected ip to be allowed")
	}
	if vd.contains("172.16.0.1") {
		t.Fatalf("did not expect ip to be allowed")
	}
}

func TestNewValidator_ConfigMissing(t *testing.T) {
	vd, err := newValidator(&server.AppConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vd != nil {
		t.Fatalf("expected nil validator when subnet empty")
	}
}
