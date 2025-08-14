package commoncfg

import (
	"net"
	"os"
	"strconv"
)

type HostPort struct {
	Host string
	Port *int
}

func ReadHostPortEnv(varName string) (HostPort, error) {
	val := os.Getenv(varName)
	if val == "" {
		return HostPort{}, nil
	}
	h, p, ok := splitHostPort(val)
	if !ok {
		return HostPort{}, nil
	}
	return HostPort{Host: h, Port: &p}, nil
}

func splitHostPort(addr string) (string, int, bool) {
	h, ps, err := net.SplitHostPort(addr)
	if err != nil || ps == "" {
		return "", 0, false
	}
	p, err := strconv.Atoi(ps)
	if err != nil || p <= 0 || p > 65535 {
		return "", 0, false
	}
	return h, p, true
}
