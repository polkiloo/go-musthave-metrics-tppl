package agent

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	models "github.com/polkiloo/go-musthave-metrics-tppl/internal/model"
)

type SenderInterface interface {
	Send(gauges map[string]models.Gauge, counters map[string]models.Counter)
}

type Sender struct {
	baseURL string
	port    int
	client  *http.Client
}

func NewSender(baseURL string, port int) *Sender {
	return &Sender{
		baseURL: strings.TrimRight(baseURL, "/"),
		port:    port,
		client:  http.DefaultClient,
	}
}

func (s *Sender) makeServerURL() string {
	// http://host:port
	return fmt.Sprintf("%s:%d", s.baseURL, s.port)
}

func (s *Sender) Send(gauges map[string]models.Gauge, counters map[string]models.Counter) {
	serverAddr := s.makeServerURL()
	for name, value := range gauges {
		s.sendMetric(serverAddr, models.GaugeType, name, fmt.Sprintf("%f", value))
	}
	for name, value := range counters {
		s.sendMetric(serverAddr, models.CounterType, name, strconv.FormatInt(int64(value), 10))
	}
}

func (s *Sender) sendMetric(serverAddr string, metricType models.MetricType, name, value string) {
	url := fmt.Sprintf("%s/update/%s/%s/%s", serverAddr, metricType, name, value)
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(""))
	if err != nil {
		log.Printf("create request error: %v", err)
		return
	}
	req.Header.Set("Content-Type", "text/plain")

	resp, err := s.client.Do(req)
	if err != nil {
		log.Printf("send error for %s: %v", name, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("non-ok response for %s: %s", name, resp.Status)
	}
}

var _ SenderInterface = NewSender("", 0)
