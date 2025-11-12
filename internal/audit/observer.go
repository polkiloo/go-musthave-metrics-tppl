package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sync"
)

// Observer receives notifications about published audit events.
type Observer interface {
	Notify(context.Context, Event) error
}

// HTTPClient represents the subset of http.Client used by observers.
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

type fileObserver struct {
	path string
	open fileOpener
	mu   sync.Mutex
}

type fileWriter interface {
	io.WriteCloser
}

type fileOpener func(string, int, os.FileMode) (fileWriter, error)

// NewFileObserver writes audit events line-by-line to the configured file path.
func NewFileObserver(path string) Observer {
	return &fileObserver{path: path, open: openAuditFile}
}

func (f *fileObserver) Notify(_ context.Context, event Event) error {
	if f == nil || f.path == "" {
		return errors.New("file observer path not configured")
	}
	if f.open == nil {
		f.open = openAuditFile
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	fd, err := f.open(f.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open audit file: %w", err)
	}
	defer fd.Close()
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal audit event: %w", err)
	}
	if _, err = fd.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("write audit event: %w", err)
	}
	return nil
}

func openAuditFile(name string, flag int, perm os.FileMode) (fileWriter, error) {
	return os.OpenFile(name, flag, perm)
}

type httpObserver struct {
	endpoint *url.URL
	client   HTTPClient
}

// NewHTTPObserver posts audit events to the supplied URL.
func NewHTTPObserver(rawURL string, client HTTPClient) (Observer, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse audit url: %w", err)
	}
	if client == nil {
		client = http.DefaultClient
	}
	return &httpObserver{endpoint: parsed, client: client}, nil
}

func (h *httpObserver) Notify(ctx context.Context, event Event) error {
	if h == nil || h.endpoint == nil || h.client == nil {
		return errors.New("http observer not configured")
	}
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal audit event: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.endpoint.String(), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build audit request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("send audit request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("audit request failed: %s", resp.Status)
	}
	return nil
}
