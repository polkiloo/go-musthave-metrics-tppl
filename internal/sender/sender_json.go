package sender

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/compression"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/logger"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/models"
)

type JSONSender struct {
	baseURL string
	port    int
	client  *http.Client
	log     logger.Logger
	comp    compression.Compressor
}

func NewJSONSender(baseURL string, port int, client *http.Client, l logger.Logger, c compression.Compressor) *JSONSender {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	return &JSONSender{
		baseURL: fmt.Sprintf("http://%s:%d", baseURL, port),
		client:  client,
		log:     l,
		comp:    c,
	}
}

func (s *JSONSender) Send(metrics []*models.Metrics) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, val := range metrics {
		s.postMetric(ctx, val)
	}
}

func (s *JSONSender) postMetric(ctx context.Context, m *models.Metrics) {
	body, err := json.Marshal(m)
	if err != nil {
		if s.log != nil {
			s.log.WriteError("marshal metric failed", "id", m.ID, "type", m.MType, "error", err)
		}
		return
	}

	var buf bytes.Buffer
	if s.comp != nil {
		zw, err := s.comp.NewWriter(&buf)
		if err != nil {
			if s.log != nil {
				s.log.WriteError("encode body failed", "id", m.ID, "type", m.MType, "error", err)
			}
			return
		}
		if _, err := zw.Write(body); err != nil {
			if s.log != nil {
				s.log.WriteError("encode body failed", "id", m.ID, "type", m.MType, "error", err)
			}
			zw.Close()
			return
		}
		if err := zw.Close(); err != nil {
			if s.log != nil {
				s.log.WriteError("encode body failed", "id", m.ID, "type", m.MType, "error", err)
			}
			return
		}
	} else {
		buf.Write(body)
	}

	reader := bytes.NewReader(buf.Bytes())

	u := s.baseURL + "/update"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, reader)
	if err != nil {
		if s.log != nil {
			s.log.WriteError("build request failed", "url", u, "error", err)
		}
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if s.comp != nil {
		enc := s.comp.ContentEncoding()
		req.Header.Set("Content-Encoding", enc)
		req.Header.Set("Accept-Encoding", enc)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		if s.log != nil {
			s.log.WriteError("post metric failed", "url", u, "id", m.ID, "type", m.MType, "error", err)
		}
		return
	}
	defer resp.Body.Close()

	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		if s.log != nil {
			s.log.WriteError("unexpected content-type", "url", u, "got", ct)
		}
	}
	if resp.StatusCode != http.StatusOK {
		if s.log != nil {
			s.log.WriteError("unexpected status", "url", u, "status", resp.Status)
		}
		return
	}

	if s.log != nil {
		s.log.WriteInfo("metric sent", "id", m.ID, "type", m.MType, "endpoint", u)
	}
}

var _ SenderInterface = NewJSONSender("", 0, nil, nil, nil)
