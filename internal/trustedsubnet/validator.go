package trustedsubnet

import (
	"net"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/server"
)

type validator struct {
	cidr *net.IPNet
}

func newValidator(cfg *server.AppConfig) (*validator, error) {
	if cfg == nil || cfg.TrustedSubnet == "" {
		return nil, nil
	}

	_, cidr, err := net.ParseCIDR(cfg.TrustedSubnet)
	if err != nil {
		return nil, err
	}

	return &validator{cidr: cidr}, nil
}

func (v *validator) contains(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	return ip != nil && v.cidr.Contains(ip)
}
