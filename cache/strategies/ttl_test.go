package strategies

import (
	"github.com/kimvlry/caching/cache"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTTLCache tests the NewTtlCache cache implementation with default NewTtlCache
func TestTTLCache(t *testing.T) {
	c := NewTtlCache[string, int](3, 100*time.Millisecond)()

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

// TestTTLCacheWithCustomTTL tests individual NewTtlCache per item
func TestTTLCacheWithCustomTTL(t *testing.T) {
	c := NewTtlCache[string, string](10, 500*time.Millisecond)()

	// Type assert to TTLCache interface
	ttlCache, ok := c.(TTLCache[string, string])
	require.True(t, ok, "cache should implement TTLCache interface")

	// Add items with different TTLs
	err := ttlCache.Set("default", "uses 500ms")
	require.NoError(t, err)

	err = ttlCache.SetWithTTL("short", "expires in 100ms", 100*time.Millisecond)
	require.NoError(t, err)

	err = ttlCache.SetWithTTL("long", "expires in 1s", 1*time.Second)
	require.NoError(t, err)

	// All should be accessible immediately
	val, err := c.Get("default")
	require.NoError(t, err)
	assert.Equal(t, "uses 500ms", val)

	val, err = c.Get("short")
	require.NoError(t, err)
	assert.Equal(t, "expires in 100ms", val)

	val, err = c.Get("long")
	require.NoError(t, err)
	assert.Equal(t, "expires in 1s", val)

	// Wait 150ms - short should expire
	time.Sleep(150 * time.Millisecond)

	_, err = c.Get("short")
	assert.Error(t, err, "short-lived item should be expired")

	val, err = c.Get("default")
	require.NoError(t, err, "default NewTtlCache item should still be alive")
	assert.Equal(t, "uses 500ms", val)

	val, err = c.Get("long")
	require.NoError(t, err, "long-lived item should still be alive")
	assert.Equal(t, "expires in 1s", val)

	// Wait another 400ms (total 550ms) - default should expire
	time.Sleep(400 * time.Millisecond)

	_, err = c.Get("default")
	assert.Error(t, err, "default NewTtlCache item should be expired")

	val, err = c.Get("long")
	require.NoError(t, err, "long-lived item should still be alive")
	assert.Equal(t, "expires in 1s", val)

	// Wait another 500ms (total 1050ms) - all should expire
	time.Sleep(500 * time.Millisecond)

	_, err = c.Get("long")
	assert.Error(t, err, "long-lived item should be expired")
}

// TestTTLCacheUpdateWithCustomTTL tests updating existing items with new NewTtlCache
func TestTTLCacheUpdateWithCustomTTL(t *testing.T) {
	c := NewTtlCache[string, int](10, 1*time.Second)()
	ttlCache := c.(TTLCache[string, int])

	// Add item with short NewTtlCache
	err := ttlCache.SetWithTTL("key", 1, 100*time.Millisecond)
	require.NoError(t, err)

	// Wait 50ms, then update with longer NewTtlCache
	time.Sleep(50 * time.Millisecond)
	err = ttlCache.SetWithTTL("key", 2, 500*time.Millisecond)
	require.NoError(t, err)

	// Wait another 100ms (total 150ms from original)
	// Original would have expired, but new NewTtlCache keeps it alive
	time.Sleep(100 * time.Millisecond)

	val, err := c.Get("key")
	require.NoError(t, err, "item should still be alive with extended NewTtlCache")
	assert.Equal(t, 2, val, "value should be updated")

	// Wait until new NewTtlCache expires
	time.Sleep(400 * time.Millisecond)

	_, err = c.Get("key")
	assert.Error(t, err, "item should be expired after new NewTtlCache")
}

// TestTTLCacheMixedOperations tests mixing Set and SetWithTTL
func TestTTLCacheMixedOperations(t *testing.T) {
	c := NewTtlCache[string, string](5, 200*time.Millisecond)()
	ttlCache := c.(TTLCache[string, string])

	// Mix of default and custom TTLs
	_ = c.Set("a", "default")
	_ = ttlCache.SetWithTTL("b", "short", 50*time.Millisecond)
	_ = ttlCache.SetWithTTL("c", "long", 400*time.Millisecond)
	_ = c.Set("d", "default2")

	// Check all exist
	for _, key := range []string{"a", "b", "c", "d"} {
		_, err := c.Get(key)
		require.NoError(t, err, "key %s should exist", key)
	}

	// Wait 100ms - only 'b' should expire
	time.Sleep(100 * time.Millisecond)

	_, err := c.Get("b")
	assert.Error(t, err, "b should be expired")

	for _, key := range []string{"a", "c", "d"} {
		_, err := c.Get(key)
		require.NoError(t, err, "key %s should still exist", key)
	}

	// Wait another 150ms (total 250ms) - a and d should expire
	time.Sleep(150 * time.Millisecond)

	_, err = c.Get("a")
	assert.Error(t, err, "a should be expired")

	_, err = c.Get("d")
	assert.Error(t, err, "d should be expired")

	val, err := c.Get("c")
	require.NoError(t, err, "c should still be alive")
	assert.Equal(t, "long", val)
}

// TestTTLCacheEvictionEvents tests that eviction events are fired
func TestTTLCacheEvictionEvents(t *testing.T) {
	c := NewTtlCache[string, int](3, 100*time.Millisecond)()
	ttlCache := c.(TTLCache[string, int])

	evictions := 0
	if observable, ok := c.(interface {
		OnEvent(func(event cache.Event[string, int]))
	}); ok {
		observable.OnEvent(func(event cache.Event[string, int]) {
			if event.Type == cache.EventTypeEviction {
				evictions++
			}
		})
	}

	// Add items with different TTLs
	_ = ttlCache.SetWithTTL("a", 1, 50*time.Millisecond)
	_ = ttlCache.SetWithTTL("b", 2, 150*time.Millisecond)
	_ = ttlCache.SetWithTTL("c", 3, 250*time.Millisecond)

	// Wait for first expiration
	time.Sleep(80 * time.Millisecond)
	_, _ = c.Get("a") // Trigger eviction check

	assert.Equal(t, 1, evictions, "one item should be evicted")

	// Wait for second expiration
	time.Sleep(100 * time.Millisecond)
	_, _ = c.Get("b") // Trigger eviction check

	assert.Equal(t, 2, evictions, "two items should be evicted")

	// Wait for third expiration
	time.Sleep(100 * time.Millisecond)
	_, _ = c.Get("c") // Trigger eviction check

	assert.Equal(t, 3, evictions, "all items should be evicted")
}

// TestTTLCacheCapacityWithCustomTTL tests capacity limits with different TTLs
func TestTTLCacheCapacityWithCustomTTL(t *testing.T) {
	c := NewTtlCache[string, int](3, 1*time.Second)()
	ttlCache := c.(TTLCache[string, int])

	// Fill cache
	_ = ttlCache.SetWithTTL("a", 1, 100*time.Millisecond)
	_ = ttlCache.SetWithTTL("b", 2, 200*time.Millisecond)
	_ = ttlCache.SetWithTTL("c", 3, 300*time.Millisecond)

	// Add fourth item - should evict 'a' (shortest NewTtlCache, earliest expiration)
	_ = ttlCache.SetWithTTL("d", 4, 400*time.Millisecond)

	_, err := c.Get("a")
	assert.Error(t, err, "a should be evicted due to capacity")

	for _, key := range []string{"b", "c", "d"} {
		_, err := c.Get(key)
		require.NoError(t, err, "key %s should exist", key)
	}
}

// TestTTLCacheZeroTTL tests edge case with zero NewTtlCache
func TestTTLCacheZeroTTL(t *testing.T) {
	c := NewTtlCache[string, int](5, 100*time.Millisecond)()
	ttlCache := c.(TTLCache[string, int])

	// Item with zero NewTtlCache should expire immediately
	err := ttlCache.SetWithTTL("instant", 42, 0)
	require.NoError(t, err)

	// Should be expired immediately
	_, err = c.Get("instant")
	assert.Error(t, err, "item with zero NewTtlCache should be immediately expired")
}
