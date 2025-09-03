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
		if req.GetBody != nil {
			body, err := req.GetBody()
			if err != nil {
				return err
			}
			req.Body = body
		}

		r, e := c.Do(req)
		if e != nil {
			if r != nil && r.Body != nil {
				r.Body.Close()
			}
			return e
		}
		resp = r
		return nil
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
