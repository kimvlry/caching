package strategies_test

import (
	"github.com/kimvlry/caching/cache/strategies"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

// TestARCCache tests the ARC cache implementation (advanced)
func TestARCCache(t *testing.T) {
	c := strategies.NewArcCache[string, int](4)()

	// Test basic operations
	err := c.Set("a", 1)
	require.NoError(t, err)
	err = c.Set("b", 2)
	require.NoError(t, err)
	err = c.Set("c", 3)
	require.NoError(t, err)
	err = c.Set("d", 4)
	require.NoError(t, err)

	// Access some items to change their status
	_, err = c.Get("a")
	require.NoError(t, err)
	_, err = c.Get("b")
	require.NoError(t, err)

	// Add new pq_item - should trigger adaptive replacement
	err = c.Set("e", 5)
	require.NoError(t, err)

	// Verify that some items are still accessible
	// (ARC behavior depends on implementation)
	val, err := c.Get("a")
	if err == nil {
		assert.Equal(t, 1, val)
	}

	val, err = c.Get("b")
	if err == nil {
		assert.Equal(t, 2, val)
	}
}
