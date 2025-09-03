package retrier

import (
	"context"
	"time"
)

var delays = []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

func Do(ctx context.Context, fn func() error, canRetry func(error) bool) error {
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

func SetDelays(ds []time.Duration) {
	delays = ds
}

func ResetDelays() {
	delays = []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}
}
