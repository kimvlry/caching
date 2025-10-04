package strategies

import (
	"caching-labwork/cache/fabric"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTTLCache tests the TTL cache implementation
func TestTTLCache(t *testing.T) {
	c := fabric.NewTTLCache[string, int](3, 100*time.Millisecond)

	// Test basic operations
	err := c.Set("a", 1)
	require.NoError(t, err)
	err = c.Set("b", 2)
	require.NoError(t, err)

	val, err := c.Get("a")
	require.NoError(t, err)
	assert.Equal(t, 1, val)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should not find expired entries
	_, err = c.Get("a")
	assert.Error(t, err)

	_, err = c.Get("b")
	assert.Error(t, err)

	// Test that new entries work after expiration
	err = c.Set("c", 3)
	require.NoError(t, err)
	val, err = c.Get("c")
	require.NoError(t, err)
	assert.Equal(t, 3, val)
}
