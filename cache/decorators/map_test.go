package decorators

import (
	"github.com/kimvlry/caching/cache/decorators/common"
	"testing"

	"github.com/kimvlry/caching/cache"
	"github.com/kimvlry/caching/cache/strategies"
)

func TestWithMap_SetStoresOriginalValue(t *testing.T) {
	baseCache := strategies.NewLRUCache[string, int](10)

	mapped := WithMap(
		baseCache,
		func(v int) int { return v * 10 },
		func() cache.IterableCache[string, int] {
			return strategies.NewLRUCache[string, int](10)
		},
	)

	err := mapped.Set("key", 5)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	val, err := mapped.Get("key")
	if err != nil || val != 5 {
		t.Errorf("Mapped cache should have value 5 (Set is raw), got %d", val)
	}

	_, err = baseCache.Get("key")
	if err == nil {
		t.Error("Base cache should not be mutated")
	}
}

func TestWithMap_WithFilter(t *testing.T) {
	baseCache := strategies.NewLRUCache[string, int](10)
	for i := 1; i <= 10; i++ {
		_ = baseCache.Set(string(rune('a'+i-1)), i)
	}

	composed := WithMap[string, int](
		WithFilter(
			baseCache,
			func(v int) bool { return v%2 == 0 },
			func() cache.IterableCache[string, int] {
				return strategies.NewLRUCache[string, int](10)
			},
		),
		func(v int) int { return v * 2 },
		func() cache.IterableCache[string, int] {
			return strategies.NewLRUCache[string, int](10)
		},
	)

	testCases := []struct {
		key         string
		expected    int
		shouldExist bool
	}{
		{"a", 0, false},
		{"b", 4, true},
		{"c", 0, false},
		{"d", 8, true},
	}

	for _, tc := range testCases {
		val, err := composed.Get(tc.key)
		exists := err == nil

		if exists != tc.shouldExist {
			t.Errorf("%s: expected exists=%v, got exists=%v", tc.key, tc.shouldExist, exists)
			continue
		}

		if exists && val != tc.expected {
			t.Errorf("%s: expected %d, got %d", tc.key, tc.expected, val)
		}
	}
}

func TestWithMap_Snapshot(t *testing.T) {
	baseCache := strategies.NewLRUCache[string, int](10)
	_ = baseCache.Set("a", 1)
	_ = baseCache.Set("b", 2)
	_ = baseCache.Set("c", 3)

	mapped := WithMap(
		baseCache,
		func(v int) int { return v * 10 },
		func() cache.IterableCache[string, int] {
			return strategies.NewLRUCache[string, int](10)
		},
	)

	snapshot := common.Snapshot[string, int](
		mapped,
		func() cache.IterableCache[string, int] {
			return strategies.NewLRUCache[string, int](10)
		},
	)

	testCases := []struct {
		key      string
		expected int
	}{
		{"a", 10},
		{"b", 20},
		{"c", 30},
	}

	for _, tc := range testCases {
		val, err := snapshot.Get(tc.key)
		if err != nil {
			t.Errorf("Get(%s) failed: %v", tc.key, err)
			continue
		}
		if val != tc.expected {
			t.Errorf("Get(%s): expected %d, got %d", tc.key, tc.expected, val)
		}
	}

	_ = baseCache.Set("a", 999)
	val, _ := snapshot.Get("a")
	if val != 10 {
		t.Error("Snapshot should be independent of source cache")
	}
}
