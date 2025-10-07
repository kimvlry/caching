package decorators

import (
	"testing"
	"time"

	"github.com/kimvlry/caching/cache/strategies"
)

func TestWithMap_SetStoresOriginalValue(t *testing.T) {
	baseCache := strategies.NewLruCache[string, int](10)()

	mapped := WithMap(
		baseCache,
		func(v int) int { return v * 10 },
		strategies.NewLruCache[string, int](10),
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
	baseCache := strategies.NewLruCache[string, int](10)()
	for i := 1; i <= 10; i++ {
		_ = baseCache.Set(string(rune('a'+i-1)), i)
	}

	composed := WithMap(
		WithFilter(
			baseCache,
			func(v int) bool { return v%2 == 0 },
			strategies.NewLruCache[string, int](10),
		),
		func(v int) int { return v * 2 },
		strategies.NewLruCache[string, int](10),
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

func TestWithMap_DifferentEvictionStrategy_LFU(t *testing.T) {
	baseCache := strategies.NewLruCache[string, int](10)()
	_ = baseCache.Set("a", 1)
	_ = baseCache.Set("b", 2)
	_ = baseCache.Set("c", 3)

	mapped := WithMap(
		baseCache,
		func(v int) int { return v * 10 },
		strategies.NewLfuCache[string, int](10),
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
		val, err := mapped.Get(tc.key)
		if err != nil {
			t.Errorf("Get(%s) failed: %v", tc.key, err)
			continue
		}
		if val != tc.expected {
			t.Errorf("Get(%s): expected %d, got %d", tc.key, tc.expected, val)
		}
	}

	for k, expected := range map[string]int{"a": 1, "b": 2, "c": 3} {
		val, _ := baseCache.Get(k)
		if val != expected {
			t.Errorf("Base cache mutated! key=%s, val=%d", k, val)
		}
	}
}

func TestWithMap_DifferentEvictionStrategy_TTL(t *testing.T) {
	baseCache := strategies.NewLruCache[string, int](10)()
	_ = baseCache.Set("x", 5)
	_ = baseCache.Set("y", 10)
	_ = baseCache.Set("z", 15)

	mapped := WithMap(
		baseCache,
		func(v int) int { return v + 100 },
		strategies.NewTtlCache[string, int](10, 150*time.Millisecond),
	)

	val, err := mapped.Get("x")
	if err != nil || val != 105 {
		t.Errorf("x expected 105, got %d", val)
	}

	val, err = mapped.Get("y")
	if err != nil || val != 110 {
		t.Errorf("y expected 110, got %d", val)
	}

	time.Sleep(200 * time.Millisecond)

	_, err = mapped.Get("x")
	if err == nil {
		t.Error("x should be expired")
	}

	for k, expected := range map[string]int{"x": 5, "y": 10, "z": 15} {
		val, _ := baseCache.Get(k)
		if val != expected {
			t.Errorf("Base cache mutated! key=%s, val=%d", k, val)
		}
	}
}

func TestWithMap_ChainedWithDifferentStrategies(t *testing.T) {
	baseCache := strategies.NewLruCache[string, int](10)()
	for i := 1; i <= 5; i++ {
		_ = baseCache.Set(string(rune('a'+i-1)), i*10)
	}

	composed := WithMap(
		WithFilter(
			baseCache,
			func(v int) bool { return v >= 30 },
			strategies.NewLfuCache[string, int](10),
		),
		func(v int) int { return v / 10 },
		strategies.NewTtlCache[string, int](10, 500*time.Millisecond),
	)

	testCases := []struct {
		key         string
		expected    int
		shouldExist bool
	}{
		{"a", 0, false}, // 10 < 30
		{"b", 0, false}, // 20 < 30
		{"c", 3, true},  // 30 / 10 = 3
		{"d", 4, true},  // 40 / 10 = 4
		{"e", 5, true},  // 50 / 10 = 5
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

	for i := 1; i <= 5; i++ {
		key := string(rune('a' + i - 1))
		val, _ := baseCache.Get(key)
		if val != i*10 {
			t.Errorf("Base cache mutated! key=%s, val=%d", key, val)
		}
	}
}
