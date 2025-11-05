package commoncfg

import (
	"flag"
	"io"
)

// FlagValue represents a parsed value returned by flag handlers.
type FlagValue interface{}

// FlagHandler converts raw flag values into typed FlagValue instances.
type FlagHandler func(value string, present bool) (FlagValue, error)

// Lift adapts a strongly typed parser into a generic FlagHandler.
func Lift[T any](fn func(string, bool) (T, error)) FlagHandler {
	return func(v string, present bool) (FlagValue, error) {
		res, err := fn(v, present)
		if err != nil {
			return nil, err
		}
		return res, nil
	}
}

// Dispatcher orchestrates parsing and applying command-line flag values to configuration structs.
type Dispatcher[U any] struct {
	fs       *flag.FlagSet
	handlers map[string]FlagHandler
	apply    func(dst *U, v FlagValue) error
}

// NewDispatcher constructs a Dispatcher for the supplied flag set and apply function.
func NewDispatcher[U any](fs *flag.FlagSet, apply func(*U, FlagValue) error) *Dispatcher[U] {
	fs.SetOutput(io.Discard)
	return &Dispatcher[U]{fs: fs, handlers: make(map[string]FlagHandler), apply: apply}
}

// Handle registers a handler for the named flag.
func (d *Dispatcher[U]) Handle(name string, h FlagHandler) *Dispatcher[U] {
	d.handlers[name] = h
	return d
}

// Parse executes the configured handlers and applies their values to the destination type.
func (d *Dispatcher[U]) Parse(args []string) (U, error) {
	var zero U
	if err := d.fs.Parse(args); err != nil {
		return zero, err
	}

	set := map[string]bool{}
	d.fs.Visit(func(f *flag.Flag) { set[f.Name] = true })

	out := zero
	for name, h := range d.handlers {
		if f := d.fs.Lookup(name); f != nil {
			part, err := h(f.Value.String(), set[name])
			if err != nil {
				return zero, err
			}
			if err := d.apply(&out, part); err != nil {
				return zero, err
			}
		}
	}
	return out, nil
}
