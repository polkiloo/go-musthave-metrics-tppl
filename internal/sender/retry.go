package sender

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/retrier"
)

func doRequest(ctx context.Context, c *http.Client, req *http.Request, delays []time.Duration) (*http.Response, error) {
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
	}, isNetError, delays)
	return resp, err
}

type statusError struct {
	code int
}

func (e statusError) Error() string {
	return fmt.Sprintf("unexpected status code: %d", e.code)
}

func isNetError(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return false
	}
	var ne net.Error
	if errors.As(err, &ne) {
		return true
	}
	var se statusError
	if errors.As(err, &se) {
		return se.code >= 500 && se.code <= 599
	}
	return false
}
