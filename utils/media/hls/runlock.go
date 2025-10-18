package hls

import "sync"

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Task represents a running task with its result
type Task[V any] struct {
	Result V
	Error  error
	Done   chan struct{}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// RunLock ensures a task runs only once per key
// Based on Kyoo's RunLock but adapted for our HLS package
type RunLock[K comparable, V any] struct {
	running map[K]*Task[V]
	lock    sync.Mutex
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewRunLock creates a new run lock
func NewRunLock[K comparable, V any]() RunLock[K, V] {
	return RunLock[K, V]{
		running: make(map[K]*Task[V]),
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Start starts a task if it's not already running
// Returns a function to get the result and a function to set the result
func (r *RunLock[K, V]) Start(key K) (func() (V, error), func(V, error) (V, error)) {
	r.lock.Lock()
	defer r.lock.Unlock()

	// If task is already running, return the existing task's getter
	if task, exists := r.running[key]; exists {
		return func() (V, error) {
			<-task.Done
			return task.Result, task.Error
		}, nil
	}

	// Create new task
	task := &Task[V]{
		Done: make(chan struct{}),
	}
	r.running[key] = task

	// Return getter and setter
	getter := func() (V, error) {
		<-task.Done
		return task.Result, task.Error
	}

	setter := func(result V, err error) (V, error) {
		task.Result = result
		task.Error = err
		close(task.Done)

		// Clean up the running task
		r.lock.Lock()
		delete(r.running, key)
		r.lock.Unlock()

		return result, err
	}

	return getter, setter
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// IsRunning checks if a task is currently running for the given key
func (r *RunLock[K, V]) IsRunning(key K) bool {
	r.lock.Lock()
	defer r.lock.Unlock()
	_, exists := r.running[key]
	return exists
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// RunningCount returns the number of currently running tasks
func (r *RunLock[K, V]) RunningCount() int {
	r.lock.Lock()
	defer r.lock.Unlock()
	return len(r.running)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Clear removes all running tasks (use with caution)
func (r *RunLock[K, V]) Clear() {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.running = make(map[K]*Task[V])
}
