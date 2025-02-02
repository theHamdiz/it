package pool

import "sync"

// Pool is a generic object pool because allocation is expensive,
// and we're all about that performance life
type Pool[T any] struct {
	pool sync.Pool
	new  func() T
}

// NewPool creates a new generic object pool because why allocate
// when you can reuse.
func NewPool[T any](new func() T) *Pool[T] {
	return &Pool[T]{
		pool: sync.Pool{
			New: func() interface{} {
				return new()
			},
		},
		new: new,
	}
}

// Get retrieves an object from the pool, or creates a new one if empty.
func (p *Pool[T]) Get() T {
	obj := p.pool.Get()
	if obj == nil {
		// Explicitly create a new object if needed
		return p.new()
	}
	return obj.(T)
}

// Put returns an object to the pool, but prevents nil values from being stored.
func (p *Pool[T]) Put(x T) {
	// Ensure we don't store nil values (only applicable for pointer types)
	var zero T
	// Workaround to check for nil for generic types
	if any(x) == any(zero) {
		return
	}
	p.pool.Put(x)
}

// PutAll returns all objects to the pool, but prevents nil values from being stored.
func (p *Pool[T]) PutAll(xs []T) {
	for _, x := range xs {
		p.Put(x)
	}
}
