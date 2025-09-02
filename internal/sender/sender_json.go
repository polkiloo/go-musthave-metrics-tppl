package sender

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/compression"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/logger"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/models"
)

var (
	ErrJSONSenderMarshal               = errors.New("marshal metric failed")
	ErrJSONSenderEncodeBody            = errors.New("encode body failed")
	ErrJSONSenderBuildRequest          = errors.New("build request failed")
	ErrJSONSenderUnexpectedContentType = errors.New("unexpected content-type")
	ErrJSONSenderUnexpectedStatus      = errors.New("unexpected status")
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

func (s *JSONSender) SendBatch(metrics []*models.Metrics) {
	if len(metrics) == 0 {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	body, err := s.marshalMetrics(metrics)
	if err != nil {
		s.log.WriteError(ErrJSONSenderMarshal.Error(), "error", err)
		return
	}

	encoded, err := s.encodeBody(body)
	if err != nil {
		s.log.WriteError(ErrJSONSenderEncodeBody.Error(), "error", err)
		return
	}

	req, err := s.buildBatchRequest(ctx, encoded)
	if err != nil {
		s.log.WriteError(ErrJSONSenderBuildRequest.Error(), "url", s.baseURL+"/updates", "error", err)
		return
	}

	resp, err := s.client.Do(req)
	if err != nil {
		s.log.WriteError("post metric failed", "url", s.baseURL+"/updates", "error", err)
		return
	}
	defer resp.Body.Close()

	if err := s.validateResponse(resp); err != nil {
		s.log.WriteError(err.Error(), "url", s.baseURL+"/updates")
		return
	}

	s.log.WriteInfo("metrics batch sent", "count", len(metrics), "endpoint", s.baseURL+"/updates")
}

func (s *JSONSender) postMetric(ctx context.Context, m *models.Metrics) {
	body, err := s.marshalMetric(m)
	if err != nil {
		s.log.WriteError(ErrJSONSenderMarshal.Error(), "id", m.ID, "type", m.MType, "error", err)
		return
	}

	encoded, err := s.encodeBody(body)
	if err != nil {
		s.log.WriteError(ErrJSONSenderEncodeBody.Error(), "id", m.ID, "type", m.MType, "error", err)
		return
	}

	req, err := s.buildRequest(ctx, encoded)
	if err != nil {
		s.log.WriteError(ErrJSONSenderBuildRequest.Error(), "url", s.baseURL+"/update", "error", err)
		return
	}

	resp, err := s.client.Do(req)
	if err != nil {
		s.log.WriteError("post metric failed", "url", s.baseURL+"/update", "id", m.ID, "type", m.MType, "error", err)
		return
	}
	defer resp.Body.Close()

	if err := s.validateResponse(resp); err != nil {
		s.log.WriteError(err.Error(), "url", s.baseURL+"/update")
		return
	}

	s.log.WriteInfo("metric sent", "id", m.ID, "type", m.MType, "endpoint", s.baseURL+"/update")
}

func (s *JSONSender) marshalMetric(m *models.Metrics) ([]byte, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrJSONSenderMarshal, err)
	}
	return b, nil
}

func (s *JSONSender) encodeBody(raw []byte) ([]byte, error) {
	var buf bytes.Buffer
	if s.comp == nil {
		_, _ = buf.Write(raw)
		return buf.Bytes(), nil
	}
	zw, err := s.comp.NewWriter(&buf)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrJSONSenderEncodeBody, err)
	}
	if _, err := zw.Write(raw); err != nil {
		_ = zw.Close()
		return nil, fmt.Errorf("%w: %v", ErrJSONSenderEncodeBody, err)
	}
	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrJSONSenderEncodeBody, err)
	}
	return buf.Bytes(), nil
}

func (s *JSONSender) buildRequest(ctx context.Context, body []byte) (*http.Request, error) {
	u := s.baseURL + "/update"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrJSONSenderBuildRequest, err)
	}
	req.Header.Set("Content-Type", "application/json")
	if s.comp != nil {
		enc := s.comp.ContentEncoding()
		req.Header.Set("Content-Encoding", enc)
		req.Header.Set("Accept-Encoding", enc)
	}
	return req, nil
}

func (s *JSONSender) validateResponse(resp *http.Response) error {
	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "application/json") {
		return fmt.Errorf("%w: got %q", ErrJSONSenderUnexpectedContentType, ct)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: %s", ErrJSONSenderUnexpectedStatus, resp.Status)
	}
	return nil
}

func (s *JSONSender) marshalMetrics(m []*models.Metrics) ([]byte, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrJSONSenderMarshal, err)
	}
	return b, nil
}

func (s *JSONSender) buildBatchRequest(ctx context.Context, body []byte) (*http.Request, error) {
	u := s.baseURL + "/updates"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrJSONSenderBuildRequest, err)
	}
	req.Header.Set("Content-Type", "application/json")
	if s.comp != nil {
		enc := s.comp.ContentEncoding()
		req.Header.Set("Content-Encoding", enc)
		req.Header.Set("Accept-Encoding", enc)
	}
	return req, nil
}

var _ SenderInterface = NewJSONSender("", 0, nil, nil, nil)
