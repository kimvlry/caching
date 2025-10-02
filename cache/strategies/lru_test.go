package strategies_test

import (
	"testing"

	"caching-labwork/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLRUCache tests the LRU cache implementation
func TestLRUCache(t *testing.T) {
	c := cache.NewLRUCache[string, int](3)

	// Test basic operations
	err := c.Set("a", 1)
	require.NoError(t, err)
	err = c.Set("b", 2)
	require.NoError(t, err)
	err = c.Set("c", 3)
	require.NoError(t, err)

	// Access "a" to make it most recently used
	val, err := c.Get("a")
	require.NoError(t, err)
	assert.Equal(t, 1, val)

	// Add "d" - should evict "b" (least recently used)
	err = c.Set("d", 4)
	require.NoError(t, err)

	_, err = c.Get("b")
	assert.Error(t, err)

	val, err = c.Get("a")
	require.NoError(t, err)
	assert.Equal(t, 1, val)

	val, err = c.Get("c")
	require.NoError(t, err)
	assert.Equal(t, 3, val)

	val, err = c.Get("d")
	require.NoError(t, err)
	assert.Equal(t, 4, val)
}
