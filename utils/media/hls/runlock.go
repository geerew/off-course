package hls

import (
	"errors"
	"sync"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// RunLock prevents duplicate concurrent operations for the same key
type RunLock[K comparable, V any] struct {
	running map[K]*Task[V]
	lock    sync.Mutex
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Task represents an in-progress operation
type Task[V any] struct {
	ready     sync.WaitGroup
	listeners []chan Result[V]
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Result holds the outcome of an operation
type Result[V any] struct {
	ok  V
	err error
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewRunLock creates a new RunLock
func NewRunLock[K comparable, V any]() RunLock[K, V] {
	return RunLock[K, V]{
		running: make(map[K]*Task[V]),
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Start begins a new operation or waits for an existing one
// Returns either:
//   - A function to wait for an existing operation (first return)
//   - A function to complete a new operation (second return)
func (r *RunLock[K, V]) Start(key K) (func() (V, error), func(val V, err error) (V, error)) {
	r.lock.Lock()
	defer r.lock.Unlock()
	task, ok := r.running[key]

	if ok {
		// Operation already running, add listener
		ret := make(chan Result[V])
		task.listeners = append(task.listeners, ret)
		return func() (V, error) {
			res := <-ret
			return res.ok, res.err
		}, nil
	}

	// Start new operation
	r.running[key] = &Task[V]{
		listeners: make([]chan Result[V], 0),
	}

	return nil, func(val V, err error) (V, error) {
		r.lock.Lock()
		defer r.lock.Unlock()

		task, ok = r.running[key]
		if !ok {
			return val, errors.New("invalid run lock state. aborting")
		}

		// Notify all listeners
		for _, listener := range task.listeners {
			listener <- Result[V]{ok: val, err: err}
			close(listener)
		}
		delete(r.running, key)
		return val, err
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// WaitFor waits for an operation to complete if it's running
func (r *RunLock[K, V]) WaitFor(key K) (V, error) {
	r.lock.Lock()
	task, ok := r.running[key]

	if !ok {
		r.lock.Unlock()
		var val V
		return val, nil
	}

	ret := make(chan Result[V])
	task.listeners = append(task.listeners, ret)

	r.lock.Unlock()
	res := <-ret
	return res.ok, res.err
}
