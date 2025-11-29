package test

import (
	"context"
	"net/http"
	"sync"
)

type FakeObserver[T any] struct {
	mu     sync.Mutex
	Events []T
	Err    error
}

func (f *FakeObserver[T]) Notify(_ context.Context, e T) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Events = append(f.Events, e)
	return f.Err
}

func (f *FakeObserver[T]) GetEvents() []T {
	f.mu.Lock()
	defer f.mu.Unlock()
	events := make([]T, len(f.Events))
	copy(events, f.Events)
	return events
}

type FakePublisher[T any] struct {
	mu     sync.Mutex
	Events []T
	Err    error
}

func (f *FakePublisher[T]) Publish(_ context.Context, e T) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Events = append(f.Events, e)
	return f.Err
}

func (f *FakePublisher[T]) GetEvents() []T {
	f.mu.Lock()
	defer f.mu.Unlock()
	events := make([]T, len(f.Events))
	copy(events, f.Events)
	return events
}

type RoundTripFunc func(*http.Request) (*http.Response, error)

func (f RoundTripFunc) Do(r *http.Request) (*http.Response, error) {
	return f(r)
}
