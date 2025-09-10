package sender

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/logger"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/models"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/retrier"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/sign"
)

var (
	ErrSenderNilMetric        = fmt.Errorf("nil metric passed to plain sender")
	ErrSenderMissingValue     = fmt.Errorf("missing value for metric")
	ErrSenderBuildURL         = fmt.Errorf("build url failed")
	ErrSenderBuildRequest     = fmt.Errorf("build request failed")
	ErrSenderPostMetric       = fmt.Errorf("post metric failed")
	ErrSenderUnexpectedStatus = fmt.Errorf("unexpected status")
)

type PlainSender struct {
	baseURL string
	port    int
	client  *http.Client
	log     logger.Logger
	signKey sign.SignKey
}

func NewPlainSender(baseURL string, port int, client *http.Client, l logger.Logger, k sign.SignKey) *PlainSender {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	return &PlainSender{
		baseURL: fmt.Sprintf("http://%s:%d", baseURL, port),
		client:  client,
		log:     l,
		signKey: k,
	}
}

func (s *PlainSender) Send(metrics []*models.Metrics) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, val := range metrics {
		s.postMetric(ctx, val)
	}

}

func (s *PlainSender) SendBatch(metrics []*models.Metrics) {
	s.Send(metrics)
}

func (s *PlainSender) postMetric(ctx context.Context, m *models.Metrics) {
	if m == nil {
		if s.log != nil {
			s.log.WriteError(ErrSenderNilMetric.Error())
		}
		return
	}
	raw, ok := plainValue(m)
	if !ok {
		if s.log != nil {
			s.log.WriteError(ErrSenderMissingValue.Error())
		}
		return
	}

	u, err := url.JoinPath(s.baseURL, "update", string(m.MType), url.PathEscape(m.ID), url.PathEscape(raw))
	if err != nil {
		if s.log != nil {
			s.log.WriteError(ErrSenderBuildURL.Error())
		}
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, http.NoBody)
	if err != nil {
		if s.log != nil {
			s.log.WriteError(ErrSenderBuildRequest.Error())
		}
		return
	}
	req.Header.Set("Content-Type", "text/plain")
	if s.signKey != "" {
		req.Header.Set("HashSHA256", sign.NewSignerSHA256().Sign(nil, s.signKey))
	}

	resp, err := doRequest(ctx, s.client, req, retrier.DefaultDelays)
	if err != nil {
		if s.log != nil {
			s.log.WriteError(ErrSenderPostMetric.Error())
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if s.log != nil {
			s.log.WriteError(ErrSenderUnexpectedStatus.Error())
		}
		return
	}

	if s.log != nil {
		s.log.WriteInfo("metric sent (plain)", "id", m.ID, "type", m.MType, "endpoint", u)
	}
}

func plainValue(m *models.Metrics) (string, bool) {
	switch m.MType {
	case models.GaugeType:
		if m.Value == nil {
			return "", false
		}
		return strconv.FormatFloat(*m.Value, 'f', 6, 64), true
	case models.CounterType:
		if m.Delta == nil {
			return "", false
		}
		return strconv.FormatInt(*m.Delta, 10), true
	default:
		return "", false
	}
}

var _ SenderInterface = NewPlainSender("", 0, nil, nil, "")
