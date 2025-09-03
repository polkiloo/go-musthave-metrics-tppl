package sender

import (
	"context"
	"errors"
	"net"
	"net/http"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/retrier"
)

func doRequest(ctx context.Context, c *http.Client, req *http.Request) (*http.Response, error) {
	var resp *http.Response
	err := retrier.Do(ctx, func() error {
		var e error
		resp, e = c.Do(req)
		return e
	}, isNetError)
	return resp, err
}

func isNetError(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return false
	}
	var ne net.Error
	return errors.As(err, &ne)
}
