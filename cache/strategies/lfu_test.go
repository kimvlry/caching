package strategies_test

import (
	"github.com/kimvlry/caching/cache/strategies"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLFUCache tests the LFU cache implementation
func TestLFUCache(t *testing.T) {
	c := strategies.NewLFUCache[string, int](3)

	// Test basic operations
	err := c.Set("a", 1)
	require.NoError(t, err)
	err = c.Set("b", 2)
	require.NoError(t, err)
	err = c.Set("c", 3)
	require.NoError(t, err)

	// Access "a" multiple times to increase its frequency
	_, err = c.Get("a")
	require.NoError(t, err)
	_, err = c.Get("a")
	require.NoError(t, err)
	_, err = c.Get("a")
	require.NoError(t, err)

	// Access "b" once
	_, err = c.Get("b")
	require.NoError(t, err)

	// Add "d" - should evict "c" (least frequently used)
	err = c.Set("d", 4)
	require.NoError(t, err)

	_, err = c.Get("c")
	assert.Error(t, err)

	val, err := c.Get("a")
	require.NoError(t, err)
	assert.Equal(t, 1, val)

	val, err = c.Get("b")
	require.NoError(t, err)
	assert.Equal(t, 2, val)

	val, err = c.Get("d")
	require.NoError(t, err)
	assert.Equal(t, 4, val)
}
