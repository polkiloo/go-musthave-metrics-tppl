package sender

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/retrier"
)

type fakeRT struct {
	errs  []error
	calls int
}

func (f *fakeRT) RoundTrip(*http.Request) (*http.Response, error) {
	err := f.errs[f.calls]
	f.calls++
	if err != nil {
		return nil, err
	}
	return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("ok"))}, nil
}

type netTempErr struct{}

func (netTempErr) Error() string   { return "temp" }
func (netTempErr) Timeout() bool   { return true }
func (netTempErr) Temporary() bool { return true }
func TestDoRequest_RetriesOnNetError(t *testing.T) {
	retrier.SetDelays([]time.Duration{time.Millisecond})
	t.Cleanup(retrier.ResetDelays)
	rt := &fakeRT{errs: []error{netTempErr{}, nil}}
	c := &http.Client{Transport: rt}
	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	resp, err := doRequest(context.Background(), c, req)
	if err != nil || resp == nil || rt.calls != 2 {
		t.Fatalf("expected retry success, calls=2 err=%v resp=%v", err, resp)
	}
	resp.Body.Close()
}
func TestDoRequest_NoRetryOnContextError(t *testing.T) {
	retrier.SetDelays([]time.Duration{time.Millisecond})
	t.Cleanup(retrier.ResetDelays)
	rt := &fakeRT{errs: []error{context.Canceled}}
	c := &http.Client{Transport: rt}
	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	_, err := doRequest(context.Background(), c, req)
	if !errors.Is(err, context.Canceled) || rt.calls != 1 {
		t.Fatalf("expected context canceled, calls=1 err=%v", err)
	}
}
func TestIsNetError(t *testing.T) {
	if !isNetError(netTempErr{}) {
		t.Fatalf("net error not detected")
	}
	if isNetError(errors.New("x")) {
		t.Fatalf("generic error marked as net")
	}
	if isNetError(context.DeadlineExceeded) {
		t.Fatalf("deadline exceeded considered net error")
	}
}
