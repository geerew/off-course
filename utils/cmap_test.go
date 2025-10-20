package utils

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCMap_BasicOps(t *testing.T) {
	m := NewCMap[string, int]()

	// Set and Get
	m.Set("a", 1)
	v, ok := m.Get("a")
	require.True(t, ok)
	require.Equal(t, 1, v)

	// GetOrSet
	v, created := m.GetOrSet("a", 2)
	require.False(t, created)
	require.Equal(t, 1, v)

	v, created = m.GetOrSet("b", 3)
	require.True(t, created)
	require.Equal(t, 3, v)

	// GetOrCreate
	v, created = m.GetOrCreate("b", func() int { return 4 })
	require.False(t, created)
	require.Equal(t, 3, v)

	v, created = m.GetOrCreate("c", func() int { return 5 })
	require.True(t, created)
	require.Equal(t, 5, v)

	// Len, Keys, Values
	require.Equal(t, 3, m.Len())
	keys := m.Keys()
	require.ElementsMatch(t, []string{"a", "b", "c"}, keys)
	vals := m.Values()
	require.ElementsMatch(t, []int{1, 3, 5}, vals)

	// Range
	seen := make(map[string]int)
	m.Range(func(k string, v int) bool { seen[k] = v; return true })
	require.Len(t, seen, 3)

	// GetAndRemove
	v, ok = m.GetAndRemove("b")
	require.True(t, ok)
	require.Equal(t, 3, v)
	_, ok = m.Get("b")
	require.False(t, ok)

	// Remove
	m.Remove("c")
	_, ok = m.Get("c")
	require.False(t, ok)

	// Clear
	m.Clear()
	require.Equal(t, 0, m.Len())
}

func TestCMap_GetOrCreate_ConcurrentSingleCreation(t *testing.T) {
	m := NewCMap[string, int]()
	var createdCount int
	var wg sync.WaitGroup
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			_, created := m.GetOrCreate("x", func() int { createdCount++; return 42 })
			_ = created
		}()
	}
	wg.Wait()
	v, ok := m.Get("x")
	require.True(t, ok)
	require.Equal(t, 42, v)
	require.Equal(t, 1, createdCount)
}
