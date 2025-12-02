package pool

import "sync"

// Resettable describes types that can reset their internal state.
type Resettable interface{ Reset() }

// Pool is a generic wrapper around sync.Pool for types that implement Resettable.
type Pool[T Resettable] struct {
	pool sync.Pool
}

// New constructs a new Pool instance using the provided constructor for new elements.
func New[T Resettable](constructor func() T) *Pool[T] {
	return &Pool[T]{
		pool: sync.Pool{
			New: func() any { return constructor() },
		},
	}
}

// Get retrieves an object from the pool, creating a new one if necessary.
func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

// Put resets the object state and returns it to the pool for reuse.
func (p *Pool[T]) Put(obj T) {
	obj.Reset()
	p.pool.Put(obj)
}
