package agent

import (
	"fmt"
	"net"
	"net/http"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/sender"
	"go.uber.org/fx"
)

type realIPMiddleware struct {
	interfaceAddrs func() ([]net.Addr, error)
	ip             string
}

func newRealIPMiddleware(addrsFn func() ([]net.Addr, error)) (*realIPMiddleware, error) {
	mw := &realIPMiddleware{interfaceAddrs: addrsFn}
	ip, err := mw.resolveAgentIP()
	if err != nil {
		return nil, err
	}
	mw.ip = ip
	return mw, nil
}

func NewRealIPMiddleware() (*realIPMiddleware, error) {
	return newRealIPMiddleware(net.InterfaceAddrs)
}

func (m *realIPMiddleware) resolveAgentIP() (string, error) {
	addrs, err := m.interfaceAddrs()
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
	mw, err := NewRealIPMiddleware()
	if err != nil {
		return nil, err
	}
	return mw.Middleware(), nil
}

func (m *realIPMiddleware) Middleware() sender.RequestMiddleware {
	return func(req *http.Request) {
		if req != nil {
			req.Header.Set("X-Real-IP", m.ip)
		}
	}
}

// ModuleRequestMiddleware provides the request middleware via fx.
var ModuleRequestMiddleware = fx.Module("agent-request-middleware",
	fx.Provide(ProvideRealIPMiddleware),
)
