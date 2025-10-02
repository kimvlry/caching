package cache_test

import (
	"caching-labwork/cache/common"
	"testing"

	"caching-labwork/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCacheErrors tests error conditions that apply to all cache implementations
func TestCacheErrors(t *testing.T) {
	c := cache.NewFIFOCache[string, int](1)

	// Test getting non-existent key
	_, err := c.Get("nonexistent")
	assert.Error(t, err)
	assert.Equal(t, common.ErrKeyNotFound, err)

	// Test deleting non-existent key
	err = c.Delete("nonexistent")
	assert.Error(t, err)
	assert.Equal(t, common.ErrKeyNotFound, err)

	// Test basic operations work
	err = c.Set("a", 1)
	require.NoError(t, err)
	val, err := c.Get("a")
	require.NoError(t, err)
	assert.Equal(t, 1, val)
}
