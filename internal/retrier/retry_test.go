package retrier

import (
	"context"
	"errors"
	"testing"
	"time"
)

type tempErr struct{}

func (tempErr) Error() string   { return "temp" }
func (tempErr) Timeout() bool   { return true }
func (tempErr) Temporary() bool { return true }

func TestDo_NoRetryOnSuccess(t *testing.T) {
	called := 0
	err := Do(context.Background(), func() error {
		called++
		return nil
	}, func(error) bool { return true }, []time.Duration{time.Millisecond})
	if err != nil || called != 1 {
		t.Fatalf("expected 1 call no error, got %d %v", called, err)
	}
}
func TestDo_RetryUntilSuccess(t *testing.T) {
	attempts := 0
	err := Do(context.Background(), func() error {
		if attempts < 2 {
			attempts++
			return tempErr{}
		}
		attempts++
		return nil
	}, func(err error) bool { return true }, []time.Duration{time.Millisecond, time.Millisecond})
	if err != nil || attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d err %v", attempts, err)
	}
}
func TestDo_StopOnNonRetriable(t *testing.T) {
	attempts := 0
	want := errors.New("boom")
	err := Do(context.Background(), func() error {
		attempts++
		return want
	}, func(error) bool { return false }, []time.Duration{time.Millisecond})
	if err != want || attempts != 1 {
		t.Fatalf("expected 1 attempt and raw error, got %d %v", attempts, err)
	}
}
func TestDo_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	attempts := 0
	err := Do(ctx, func() error {
		attempts++
		cancel()
		return tempErr{}
	}, func(err error) bool { return true }, []time.Duration{time.Second})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
	if attempts != 1 {
		t.Fatalf("expected 1 attempt before cancel, got %d", attempts)
	}
}
func TestDo_ExhaustRetries(t *testing.T) {
	attempts := 0
	want := tempErr{}
	err := Do(context.Background(), func() error {
		attempts++
		return want
	}, func(err error) bool { return true }, []time.Duration{time.Millisecond, time.Millisecond})
	if !errors.Is(err, want) {
		t.Fatalf("expected final error %v", err)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}
