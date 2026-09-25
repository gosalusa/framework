package cache_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gosalusa.com/cache"
	"gosalusa.com/di"
)

func TestRegister(t *testing.T) {
	ctx := di.TestDependencyProviderContext()
	cache.Register(ctx)

	c, err := di.Resolve[cache.MemoryCache](ctx)
	require.NoError(t, err)
	require.NotNil(t, c)
	assert.IsType(t, &cache.MapCache{}, c)

	again, err := di.Resolve[cache.MemoryCache](ctx)
	require.NoError(t, err)
	assert.Same(t, c, again, "the cache should be registered as a singleton")

	b, err := c.Get("missing", []byte("fallback"))
	require.NoError(t, err)
	assert.Equal(t, []byte("fallback"), b)
}
