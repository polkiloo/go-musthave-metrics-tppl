package sender

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
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

type fakeRTStatus struct {
	codes []int
	calls int
}

func (f *fakeRTStatus) RoundTrip(*http.Request) (*http.Response, error) {
	code := f.codes[f.calls]
	f.calls++
	return &http.Response{StatusCode: code, Body: io.NopCloser(strings.NewReader(""))}, nil
}

type netTempErr struct{}

func (netTempErr) Error() string   { return "temp" }
func (netTempErr) Timeout() bool   { return true }
func (netTempErr) Temporary() bool { return true }
func TestDoRequest_RetriesOnNetError(t *testing.T) {

	rt := &fakeRT{errs: []error{netTempErr{}, nil}}
	c := &http.Client{Transport: rt}
	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	resp, err := doRequest(context.Background(), c, req, []time.Duration{time.Millisecond})
	if err != nil || resp == nil || rt.calls != 2 {
		t.Fatalf("expected retry success, calls=2 err=%v resp=%v", err, resp)
	}
	resp.Body.Close()
}
func TestDoRequest_NoRetryOnContextError(t *testing.T) {
	rt := &fakeRT{errs: []error{context.Canceled}}
	c := &http.Client{Transport: rt}
	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	resp, err := doRequest(context.Background(), c, req, []time.Duration{time.Millisecond})
	if resp != nil {
		resp.Body.Close()
	}
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
	if !isNetError(statusError{code: http.StatusInternalServerError}) {
		t.Fatalf("5xx status not considered net error")
	}
	if isNetError(statusError{code: http.StatusBadRequest}) {
		t.Fatalf("4xx status considered net error")
	}
}
