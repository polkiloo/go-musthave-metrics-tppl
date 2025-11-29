package audit

import (
	"context"
	"errors"
	"sync"
)

// Event represents an audit entry describing processed metrics for a request.
type Event struct {
	Timestamp int64    `json:"ts"`
	Metrics   []string `json:"metrics"`
	IPAddress string   `json:"ip_address"`
}

// Dispatcher delivers audit events to all registered observers.
type Dispatcher struct {
	mu        sync.RWMutex
	observers []Observer
}

// NewDispatcher constructs a Dispatcher with the provided observers.
func NewDispatcher(observers ...Observer) *Dispatcher {
	return &Dispatcher{observers: observers}
}

// Publish forwards the event to every registered observer and joins any errors.
func (d *Dispatcher) Publish(ctx context.Context, event Event) error {
	if d == nil {
		return nil
	}
	d.mu.RLock()
	observers := append([]Observer(nil), d.observers...)
	d.mu.RUnlock()
	var errs []error
	for _, o := range observers {
		if o == nil {
			continue
		}
		if err := o.Notify(ctx, event); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
