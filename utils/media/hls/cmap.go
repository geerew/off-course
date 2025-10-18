package hls

import "sync"

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CMap is a concurrent map implementation
// Based on Kyoo's CMap but adapted for our HLS package
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
	val, ok := m.data[key]
	return val, ok
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Set stores a value in the map
func (m *CMap[K, V]) Set(key K, value V) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.data[key] = value
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Remove deletes a value from the map
func (m *CMap[K, V]) Remove(key K) {
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.data, key)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetOrCreate retrieves a value or creates it if it doesn't exist
func (m *CMap[K, V]) GetOrCreate(key K, createFn func() V) V {
	m.lock.Lock()
	defer m.lock.Unlock()

	if val, ok := m.data[key]; ok {
		return val
	}

	val := createFn()
	m.data[key] = val
	return val
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Keys returns all keys in the map
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

// Values returns all values in the map
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

// Len returns the number of items in the map
func (m *CMap[K, V]) Len() int {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return len(m.data)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Clear removes all items from the map
func (m *CMap[K, V]) Clear() {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.data = make(map[K]V)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ForEach applies a function to each key-value pair
func (m *CMap[K, V]) ForEach(fn func(K, V)) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	for k, v := range m.data {
		fn(k, v)
	}
}
