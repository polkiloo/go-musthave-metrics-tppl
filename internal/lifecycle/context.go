package lifecycle

import (
	"context"
	"time"
)

// BoundedContext returns a background context limited by either the parent's deadline
// or the provided timeout when the parent has no deadline.
func BoundedContext(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if dl, ok := ctx.Deadline(); ok {
		return context.WithDeadline(context.Background(), dl)
	}
	return context.WithTimeout(context.Background(), timeout)
}
