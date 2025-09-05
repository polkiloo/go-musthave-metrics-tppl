package retrier

import (
	"context"
	"time"
)

var DefaultDelays = []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

// Do executes fn and retries it according to the provided delays while canRetry returns true.
// If delays is empty, DefaultDelays are used.
func Do(ctx context.Context, fn func() error, canRetry func(error) bool, delays []time.Duration) error {
	if len(delays) == 0 {
		delays = DefaultDelays
	}
	err := fn()
	if err == nil || !canRetry(err) {
		return err
	}
	for _, d := range delays {
		select {
		case <-time.After(d):
		case <-ctx.Done():
			return ctx.Err()
		}
		if err = fn(); err == nil || !canRetry(err) {
			return err
		}
	}
	return err
}
