package agent

import (
	"fmt"
	"net"
	"net/http"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/sender"
	"go.uber.org/fx"
)

var interfaceAddrs = net.InterfaceAddrs

func resolveAgentIP() (string, error) {
	addrs, err := interfaceAddrs()
	if err != nil {
		return "", err
	}

	var fallback string
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok {
			ip := ipnet.IP
			if ip == nil {
				continue
			}
			if ip.To4() != nil {
				if !ip.IsLoopback() {
					return ip.String(), nil
				}
				if fallback == "" {
					fallback = ip.String()
				}
			}
		}
	}

	if fallback != "" {
		return fallback, nil
	}

	return "", fmt.Errorf("no suitable ip found")
}

// ProvideRealIPMiddleware builds a RequestMiddleware that injects X-Real-IP header.
func ProvideRealIPMiddleware() (sender.RequestMiddleware, error) {
	ip, err := resolveAgentIP()
	if err != nil {
		return nil, err
	}

	return func(req *http.Request) {
		if req != nil {
			req.Header.Set("X-Real-IP", ip)
		}
	}, nil
}

// ModuleRequestMiddleware provides the request middleware via fx.
var ModuleRequestMiddleware = fx.Module("agent-request-middleware",
	fx.Provide(ProvideRealIPMiddleware),
)
