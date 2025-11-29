package audit

import (
	"context"
	"errors"
	"reflect"
	"testing"

	test "github.com/polkiloo/go-musthave-metrics-tppl/internal/test"
)

func TestDispatcherPublishFanOut(t *testing.T) {
	event := Event{Timestamp: 42, Metrics: []string{"A"}}
	first := &test.FakeObserver[Event]{}
	second := &test.FakeObserver[Event]{Err: errors.New("fail")}
	d := NewDispatcher(first, nil, second)

	if err := d.Publish(context.Background(), event); err == nil || !errors.Is(err, second.Err) {
		t.Fatalf("expected joined error, got %v", err)
	}

	if len(first.GetEvents()) != 1 || len(second.GetEvents()) != 1 {
		t.Fatalf("observers not invoked: %d %d", len(first.GetEvents()), len(second.GetEvents()))
	}

	events := first.GetEvents()
	if len(events) != 1 || !reflect.DeepEqual(events[0], event) {
		t.Fatalf("unexpected event %+v", events)
	}
}

func TestDispatcherPublishNil(t *testing.T) {
	var d *Dispatcher
	if err := d.Publish(context.Background(), Event{}); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}
