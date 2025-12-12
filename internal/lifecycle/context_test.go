package lifecycle

import (
	"context"
	"testing"
	"time"
)

func TestBoundedContextWithParentDeadline(t *testing.T) {
	parent, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	ctx, ctxCancel := BoundedContext(parent, 5*time.Second)
	t.Cleanup(ctxCancel)

	parentDeadline, parentOK := parent.Deadline()
	if !parentOK {
		t.Fatalf("expected parent to have deadline")
	}

	childDeadline, childOK := ctx.Deadline()
	if !childOK {
		t.Fatalf("expected bounded context to inherit deadline")
	}

	if !parentDeadline.Equal(childDeadline) {
		t.Fatalf("expected deadlines to match; parent %v child %v", parentDeadline, childDeadline)
	}
}

func TestBoundedContextWithTimeout(t *testing.T) {
	timeout := 75 * time.Millisecond
	ctx, cancel := BoundedContext(context.Background(), timeout)
	t.Cleanup(cancel)

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatalf("expected bounded context to set timeout")
	}

	now := time.Now()
	if deadline.Before(now.Add(timeout-10*time.Millisecond)) || deadline.After(now.Add(timeout+10*time.Millisecond)) {
		t.Fatalf("deadline %v not within expected window around %v", deadline, now.Add(timeout))
	}
}
