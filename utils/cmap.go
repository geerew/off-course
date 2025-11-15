package utils

import "sync"

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CMap is a thread-safe concurrent map with generic type support
type CMap[K comparable, V any] struct {
	data map[K]V
	lock sync.RWMutex
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewCMap creates a new concurrent map
func NewCMap[K comparable, V any]() CMap[K, V] {
	return CMap[K, V]{
		data: make(map[K]V),
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get retrieves a value from the map
func (m *CMap[K, V]) Get(key K) (V, bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	ret, ok := m.data[key]

	return ret, ok
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetOrCreate retrieves a value or creates it if it doesn't exist
// Returns the value and a boolean indicating if it was created
func (m *CMap[K, V]) GetOrCreate(key K, create func() V) (V, bool) {
	m.lock.RLock()

	if ret, ok := m.data[key]; ok {
		m.lock.RUnlock()
		return ret, false
	}
	m.lock.RUnlock()

	m.lock.Lock()
	defer m.lock.Unlock()

	if ret, ok := m.data[key]; ok {
		m.lock.RUnlock()
		return ret, false
	}

	val := create()
	m.data[key] = val

	return val, true
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetOrSet retrieves a value or sets it if it doesn't exist
func (m *CMap[K, V]) GetOrSet(key K, val V) (V, bool) {
	return m.GetOrCreate(key, func() V { return val })
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Set stores a value in the map
func (m *CMap[K, V]) Set(key K, val V) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.data[key] = val
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Remove deletes a value from the map
func (m *CMap[K, V]) Remove(key K) {
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.data, key)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetAndRemove retrieves and deletes a value from the map
func (m *CMap[K, V]) GetAndRemove(key K) (V, bool) {
	m.lock.Lock()
	defer m.lock.Unlock()

	val, ok := m.data[key]
	delete(m.data, key)

	return val, ok
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Len returns the number of elements in the map
func (m *CMap[K, V]) Len() int {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return len(m.data)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Keys returns a slice of all keys in the map
func (m *CMap[K, V]) Keys() []K {
	m.lock.RLock()
	defer m.lock.RUnlock()

	keys := make([]K, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}

	return keys
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Values returns a slice of all values in the map
func (m *CMap[K, V]) Values() []V {
	m.lock.RLock()
	defer m.lock.RUnlock()

	values := make([]V, 0, len(m.data))
	for _, v := range m.data {
		values = append(values, v)
	}

	return values
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Range calls fn for each key-value pair in the map
func (m *CMap[K, V]) Range(fn func(K, V) bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	for k, v := range m.data {
		if !fn(k, v) {
			break
		}
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Clear removes all elements from the map
func (m *CMap[K, V]) Clear() {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.data = make(map[K]V)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ForEach calls fn for each key-value pair in the map while holding the write lock
func (m *CMap[K, V]) ForEach(fn func(K, V)) {
	m.lock.Lock()
	defer m.lock.Unlock()

	for k, v := range m.data {
		fn(k, v)
	}
}
