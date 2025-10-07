package strategies_test

import (
	"github.com/kimvlry/caching/cache/strategies"
	"github.com/kimvlry/caching/cache/strategies/common"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFIFOCache tests the NewFifoCache cache implementation
func TestFIFOCache(t *testing.T) {
	c := strategies.NewFifoCache[string, int](3)()

	// Test basic operations
	err := c.Set("a", 1)
	require.NoError(t, err)

	val, err := c.Get("a")
	require.NoError(t, err)
	assert.Equal(t, 1, val)

	// Test NewFifoCache eviction
	err = c.Set("b", 2)
	require.NoError(t, err)
	err = c.Set("c", 3)
	require.NoError(t, err)
	err = c.Set("d", 4) // This should evict "a"
	require.NoError(t, err)

	_, err = c.Get("a")
	assert.Error(t, err)
	assert.Equal(t, common.ErrKeyNotFound, err)

	val, err = c.Get("b")
	require.NoError(t, err)
	assert.Equal(t, 2, val)

	// Test delete
	err = c.Delete("b")
	require.NoError(t, err)

	_, err = c.Get("b")
	assert.Error(t, err)

	// Test clear
	c.Clear()
	_, err = c.Get("c")
	assert.Error(t, err)
}
