package maps_test

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	maps "gosalusa.com/extra/maps"
)

func TestSync_MapInterface(t *testing.T) {
	var _ maps.Map[string, int] = &maps.Sync[string, int]{}
}

func collectAll(m *maps.Sync[string, int]) []string {
	keys := []string{}
	for k, v := range m.All() {
		_ = v
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func TestSync(t *testing.T) {
	m := &maps.Sync[string, int]{}

	m.Set("a", 1)
	m.Store("b", 2)

	v, ok := m.Get("a")
	assert.True(t, ok)
	assert.Equal(t, 1, v)

	v, ok = m.Load("b")
	assert.True(t, ok)
	assert.Equal(t, 2, v)

	v, ok = m.Load("missing")
	assert.False(t, ok)
	assert.Equal(t, 0, v)

	t.Run("CompareAndSwap", func(t *testing.T) {
		assert.True(t, m.CompareAndSwap("a", 1, 3))
		v, _ := m.Load("a")
		assert.Equal(t, 3, v)
		assert.False(t, m.CompareAndSwap("a", 99, 4))
	})

	t.Run("CompareAndDelete", func(t *testing.T) {
		m.Store("cad", 1)
		assert.True(t, m.CompareAndDelete("cad", 1))
		assert.False(t, m.CompareAndDelete("cad", 1))
	})

	t.Run("Swap", func(t *testing.T) {
		prev, loaded := m.Swap("a", 10)
		assert.True(t, loaded)
		assert.Equal(t, 3, prev)
		prev, loaded = m.Swap("new_key", 5)
		assert.False(t, loaded)
		assert.Equal(t, 0, prev)
		v, _ := m.Load("a")
		assert.Equal(t, 10, v)
	})

	t.Run("LoadOrStore", func(t *testing.T) {
		v, loaded := m.LoadOrStore("a", 999)
		assert.True(t, loaded)
		assert.Equal(t, 10, v)

		v, loaded = m.LoadOrStore("los", 7)
		assert.False(t, loaded)
		assert.Equal(t, 7, v)
	})

	t.Run("LoadAndDelete", func(t *testing.T) {
		m.Store("lad", 42)
		v, loaded := m.LoadAndDelete("lad")
		assert.True(t, loaded)
		assert.Equal(t, 42, v)

		v, loaded = m.LoadAndDelete("lad")
		assert.False(t, loaded)
		assert.Equal(t, 0, v)
	})

	t.Run("Remove and Delete", func(t *testing.T) {
		m.Store("r1", 1)
		m.Remove("r1")
		_, ok := m.Load("r1")
		assert.False(t, ok)

		m.Store("r2", 2)
		m.Delete("r2")
		_, ok = m.Load("r2")
		assert.False(t, ok)
	})

	t.Run("Range", func(t *testing.T) {
		keys := []string{}
		m.Range(func(k string, v int) bool {
			keys = append(keys, k)
			return true
		})
		sort.Strings(keys)
		assert.Equal(t, []string{"a", "b", "los", "new_key"}, keys)
	})

	t.Run("Range stops early", func(t *testing.T) {
		count := 0
		m.Range(func(k string, v int) bool {
			count++
			return false
		})
		assert.Equal(t, 1, count)
	})

	t.Run("All", func(t *testing.T) {
		assert.Equal(t, []string{"a", "b", "los", "new_key"}, collectAll(m))
	})

	t.Run("Clear", func(t *testing.T) {
		m.Clear()
		assert.Equal(t, []string{}, collectAll(m))
	})
}
