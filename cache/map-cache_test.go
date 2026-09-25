package cache

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMapCache(t *testing.T) {
	c := NewMapCache()
	require.NotNil(t, c)
	assert.Equal(t, 5*time.Minute, c.ttl)

	_, ok := c.m.Get("anything")
	assert.False(t, ok)
}

func TestMapCache_ImplementsMemoryCache(t *testing.T) {
	var c MemoryCache = NewMapCache()
	require.NotNil(t, c)
}

func TestMapCacheItem_Expired(t *testing.T) {
	t.Run("zero expiration never expires", func(t *testing.T) {
		i := MapCacheItem{Data: []byte("v")}
		assert.False(t, i.Expired())
	})

	t.Run("past expiration is expired", func(t *testing.T) {
		i := MapCacheItem{Expiration: time.Now().Add(-time.Hour)}
		assert.True(t, i.Expired())
	})

	t.Run("future expiration is not expired", func(t *testing.T) {
		i := MapCacheItem{Expiration: time.Now().Add(time.Hour)}
		assert.False(t, i.Expired())
	})
}

func TestMapCache_Get(t *testing.T) {
	t.Run("missing key returns the default", func(t *testing.T) {
		c := NewMapCache()

		b, err := c.Get("missing", []byte("fallback"))
		require.NoError(t, err)
		assert.Equal(t, []byte("fallback"), b)
	})

	t.Run("missing key returns a nil default", func(t *testing.T) {
		c := NewMapCache()

		b, err := c.Get("missing", nil)
		require.NoError(t, err)
		assert.Nil(t, b)
	})

	t.Run("existing key returns the stored value", func(t *testing.T) {
		c := NewMapCache()
		require.NoError(t, c.Set("key", []byte("value"), nil))

		b, err := c.Get("key", []byte("fallback"))
		require.NoError(t, err)
		assert.Equal(t, []byte("value"), b)
	})

	t.Run("overwritten value is returned", func(t *testing.T) {
		c := NewMapCache()
		require.NoError(t, c.Set("key", []byte("first"), nil))
		require.NoError(t, c.Set("key", []byte("second"), nil))

		b, err := c.Get("key", nil)
		require.NoError(t, err)
		assert.Equal(t, []byte("second"), b)
	})

	t.Run("empty key is supported", func(t *testing.T) {
		c := NewMapCache()
		require.NoError(t, c.Set("", []byte("value"), nil))

		b, err := c.Get("", nil)
		require.NoError(t, err)
		assert.Equal(t, []byte("value"), b)
	})

	t.Run("expired value returns the default and is evicted", func(t *testing.T) {
		c := NewMapCache()
		c.ttl = -time.Hour
		require.NoError(t, c.Set("key", []byte("value"), nil))

		b, err := c.Get("key", []byte("fallback"))
		require.NoError(t, err)
		assert.Equal(t, []byte("fallback"), b)

		_, ok := c.m.Get("key")
		assert.False(t, ok)
	})

	t.Run("other keys are unaffected", func(t *testing.T) {
		c := NewMapCache()
		require.NoError(t, c.Set("a", []byte("1"), nil))
		require.NoError(t, c.Set("b", []byte("2"), nil))

		_, err := c.Get("missing", nil)
		require.NoError(t, err)

		b, err := c.Get("b", nil)
		require.NoError(t, err)
		assert.Equal(t, []byte("2"), b)
	})
}

func TestMapCache_Set(t *testing.T) {
	t.Run("stores the value", func(t *testing.T) {
		c := NewMapCache()
		require.NoError(t, c.Set("key", []byte("value"), nil))

		item, ok := c.m.Get("key")
		require.True(t, ok)
		assert.Equal(t, []byte("value"), item.Data)
		assert.WithinDuration(t, time.Now().Add(5*time.Minute), item.Expiration, time.Minute)
	})

	t.Run("nil and empty set options behave the same", func(t *testing.T) {
		c := NewMapCache()
		require.NoError(t, c.Set("a", []byte("1"), nil))
		require.NoError(t, c.Set("b", []byte("1"), &SetOptions{}))

		a, err := c.Get("a", nil)
		require.NoError(t, err)
		b, err := c.Get("b", nil)
		require.NoError(t, err)
		assert.Equal(t, a, b)
	})

	t.Run("nil and empty values are stored", func(t *testing.T) {
		c := NewMapCache()
		require.NoError(t, c.Set("nil", nil, nil))
		require.NoError(t, c.Set("empty", []byte{}, nil))

		b, err := c.Get("nil", nil)
		require.NoError(t, err)
		assert.Nil(t, b)

		b, err = c.Get("empty", nil)
		require.NoError(t, err)
		assert.Empty(t, b)
	})

	t.Run("keeps keys isolated", func(t *testing.T) {
		c := NewMapCache()
		for _, k := range []string{"a", "b", "c"} {
			require.NoError(t, c.Set(k, []byte(k), nil))
		}

		for _, k := range []string{"a", "b", "c"} {
			b, err := c.Get(k, nil)
			require.NoError(t, err)
			assert.Equal(t, []byte(k), b)
		}
	})
}

func TestMapCache_GetOrCreate(t *testing.T) {
	t.Run("creates and caches the value", func(t *testing.T) {
		c := NewMapCache()
		calls := 0

		b, err := c.GetOrCreate("key", func() []byte {
			calls++
			return []byte("created")
		})
		require.NoError(t, err)
		assert.Equal(t, []byte("created"), b)
		assert.Equal(t, 1, calls)

		b, err = c.GetOrCreate("key", func() []byte {
			calls++
			return []byte("other")
		})
		require.NoError(t, err)
		assert.Equal(t, []byte("created"), b)
		assert.Equal(t, 1, calls)

		item, ok := c.m.Get("key")
		require.True(t, ok)
		assert.Equal(t, []byte("created"), item.Data)
	})

	t.Run("returns an existing unexpired value", func(t *testing.T) {
		c := NewMapCache()
		require.NoError(t, c.Set("key", []byte("value"), nil))

		b, err := c.GetOrCreate("key", func() []byte {
			t.Fatal("factory must not be called for a cached value")
			return nil
		})
		require.NoError(t, err)
		assert.Equal(t, []byte("value"), b)
	})

	t.Run("recreates an expired value", func(t *testing.T) {
		c := NewMapCache()
		c.ttl = -time.Hour
		require.NoError(t, c.Set("key", []byte("stale"), nil))

		b, err := c.GetOrCreate("key", func() []byte { return []byte("fresh") })
		require.NoError(t, err)
		assert.Equal(t, []byte("fresh"), b)
	})

	t.Run("factory may return nil", func(t *testing.T) {
		c := NewMapCache()

		b, err := c.GetOrCreate("key", func() []byte { return nil })
		require.NoError(t, err)
		assert.Nil(t, b)
	})

	t.Run("recreates after remove", func(t *testing.T) {
		c := NewMapCache()
		require.NoError(t, c.Remove("key"))

		b, err := c.GetOrCreate("key", func() []byte { return []byte("created") })
		require.NoError(t, err)
		assert.Equal(t, []byte("created"), b)
	})
}

func TestMapCache_Remove(t *testing.T) {
	t.Run("removes an existing key", func(t *testing.T) {
		c := NewMapCache()
		require.NoError(t, c.Set("key", []byte("value"), nil))
		require.NoError(t, c.Remove("key"))

		_, ok := c.m.Get("key")
		assert.False(t, ok)

		b, err := c.Get("key", []byte("fallback"))
		require.NoError(t, err)
		assert.Equal(t, []byte("fallback"), b)
	})

	t.Run("removing a missing key succeeds", func(t *testing.T) {
		c := NewMapCache()
		assert.NoError(t, c.Remove("missing"))
		assert.NoError(t, c.Remove("missing"))
	})

	t.Run("leaves other keys in place", func(t *testing.T) {
		c := NewMapCache()
		require.NoError(t, c.Set("a", []byte("1"), nil))
		require.NoError(t, c.Set("b", []byte("2"), nil))
		require.NoError(t, c.Remove("a"))

		b, err := c.Get("b", nil)
		require.NoError(t, err)
		assert.Equal(t, []byte("2"), b)
	})
}

func TestMapCache_Concurrent(t *testing.T) {
	c := NewMapCache()

	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			key := fmt.Sprintf("key-%d", i)
			assert.NoError(t, c.Set(key, []byte("value"), nil))

			b, err := c.Get(key, nil)
			assert.NoError(t, err)
			assert.Equal(t, []byte("value"), b)

			shared, err := c.GetOrCreate("shared", func() []byte { return []byte("shared") })
			assert.NoError(t, err)
			assert.Equal(t, []byte("shared"), shared)

			assert.NoError(t, c.Remove(key))
		}(i)
	}
	wg.Wait()
}
