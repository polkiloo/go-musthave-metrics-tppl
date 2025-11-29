package audit

import (
	"context"
	"errors"
	"testing"

	test "github.com/polkiloo/go-musthave-metrics-tppl/internal/test"
)

func TestDispatcher_Publish_NotifiesAll(t *testing.T) {
	e := Event{Timestamp: 1, Metrics: []string{"Alloc"}, IPAddress: "127.0.0.1"}

	o1 := &test.FakeObserver[Event]{}
	o2 := &test.FakeObserver[Event]{}

	d := NewDispatcher(o1, o2)
	if err := d.Publish(context.Background(), e); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(o1.GetEvents()) != 1 || len(o2.GetEvents()) != 1 {
		t.Fatalf("observers not notified: o1=%d o2=%d", len(o1.GetEvents()), len(o2.GetEvents()))
	}
}

func TestDispatcher_Publish_CollectsErrors(t *testing.T) {
	e := Event{Timestamp: 1}
	err1 := errors.New("one")
	err2 := errors.New("two")

	d := NewDispatcher(&test.FakeObserver[Event]{Err: err1}, &test.FakeObserver[Event]{Err: err2})
	if err := d.Publish(context.Background(), e); err == nil {
		t.Fatalf("expected error, got nil")
	}
}
