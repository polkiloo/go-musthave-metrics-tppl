package agent_test

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/agent"

	models "github.com/polkiloo/go-musthave-metrics-tppl/internal/model"

	"github.com/stretchr/testify/assert"
)

func TestSender_Send_AllTypes(t *testing.T) {
	got := make(map[string]struct{})
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.Split(r.URL.Path, "/")
		key := p[2] + ":" + p[3]
		got[key] = struct{}{}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	host, port := parseHostPort(ts.URL)
	sender := agent.NewSender(host, port)
	gauges := map[string]models.Gauge{"Alloc": 1.1}
	counters := map[string]models.Counter{"PollCount": 2}

	sender.Send(gauges, counters)

	assert.Contains(t, got, "gauge:Alloc")
	assert.Contains(t, got, "counter:PollCount")
}

func TestSender_Send_CreateRequestError(t *testing.T) {
	sender := agent.NewSender("://bad_url", 8080)
	sender.Send(map[string]models.Gauge{"Alloc": 1.1}, nil)
}

func TestSender_Send_HTTPError(t *testing.T) {
	addr := getUnusedPort()
	sender := agent.NewSender("http://127.0.0.1", addr)
	sender.Send(map[string]models.Gauge{"Alloc": 2.2}, nil)
}

func TestSender_Send_NonOKResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))
	defer ts.Close()
	host, port := parseHostPort(ts.URL)
	sender := agent.NewSender(host, port)
	sender.Send(map[string]models.Gauge{"Alloc": 3.3}, nil)
}

func TestSender_Send_EmptyMaps(t *testing.T) {
	sender := agent.NewSender("http://localhost", 8080)
	sender.Send(nil, nil)
	sender.Send(map[string]models.Gauge{}, map[string]models.Counter{})
}

func parseHostPort(url string) (string, int) {
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")
	parts := strings.Split(url, ":")
	if len(parts) == 2 {
		return "http://" + parts[0], mustAtoi(parts[1])
	}
	host, port, _ := net.SplitHostPort(url)
	return "http://" + host, mustAtoi(port)
}

func mustAtoi(s string) int {
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return 0
		}
	}
	n, _ := net.LookupPort("tcp", s)
	return n
}

func getUnusedPort() int {
	l, _ := net.Listen("tcp", ":0")
	defer l.Close()
	_, portStr, _ := net.SplitHostPort(l.Addr().String())
	return mustAtoi(portStr)
}
