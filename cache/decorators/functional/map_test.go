package functional

import (
	"github.com/kimvlry/caching/cache/decorators/common"
	"strings"
	"testing"

	"github.com/kimvlry/caching/cache"
	"github.com/kimvlry/caching/cache/strategies"
)

func TestWithMap_BasicMapping(t *testing.T) {
	baseCache := strategies.NewLRUCache[string, int](10)
	baseCache.Set("a", 1)
	baseCache.Set("b", 2)
	baseCache.Set("c", 3)

	mapped := WithMap[string, int](
		baseCache,
		func(v int) int { return v * 10 },
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
}

func TestWithMap_StringTransformation(t *testing.T) {
	baseCache := strategies.NewLRUCache[string, string](10)
	baseCache.Set("key1", "hello")
	baseCache.Set("key2", "world")
	baseCache.Set("key3", "test")

	mapped := WithMap[string, string](
		baseCache,
		func(v string) string { return strings.ToUpper(v) },
	)

	testCases := []struct {
		key      string
		expected string
	}{
		{"key1", "HELLO"},
		{"key2", "WORLD"},
		{"key3", "TEST"},
	}

	for _, tc := range testCases {
		val, err := mapped.Get(tc.key)
		if err != nil {
			t.Errorf("Get(%s) failed: %v", tc.key, err)
			continue
		}
		if val != tc.expected {
			t.Errorf("Get(%s): expected %s, got %s", tc.key, tc.expected, val)
		}
	}
}

func TestWithMap_ComplexTransformation(t *testing.T) {
	type Product struct {
		Name  string
		Price int
	}

	baseCache := strategies.NewLRUCache[string, Product](10)
	baseCache.Set("p1", Product{Name: "Laptop", Price: 1000})
	baseCache.Set("p2", Product{Name: "Mouse", Price: 20})

	mapped := WithMap[string, Product](
		baseCache,
		func(p Product) Product {
			return Product{
				Name:  p.Name,
				Price: int(float64(p.Price) * 0.9),
			}
		},
	)

	p1, err := mapped.Get("p1")
	if err != nil {
		t.Fatalf("Get(p1) failed: %v", err)
	}
	if p1.Price != 900 {
		t.Errorf("Expected p1.Price=900, got %d", p1.Price)
	}

	p2, err := mapped.Get("p2")
	if err != nil {
		t.Fatalf("Get(p2) failed: %v", err)
	}
	if p2.Price != 18 {
		t.Errorf("Expected p2.Price=18, got %d", p2.Price)
	}
}

func TestWithMap_ImmutabilitySourceCache(t *testing.T) {
	baseCache := strategies.NewLRUCache[string, int](10)
	baseCache.Set("key1", 5)
	baseCache.Set("key2", 10)

	_ = WithMap[string, int](
		baseCache,
		func(v int) int { return v * 2 },
	)

	val1, err := baseCache.Get("key1")
	if err != nil || val1 != 5 {
		t.Error("Source cache was mutated! key1 should still be 5")
	}

	val2, err := baseCache.Get("key2")
	if err != nil || val2 != 10 {
		t.Error("Source cache was mutated! key2 should still be 10")
	}
}

func TestWithMap_Range(t *testing.T) {
	baseCache := strategies.NewLRUCache[string, int](10)
	baseCache.Set("a", 1)
	baseCache.Set("b", 2)
	baseCache.Set("c", 3)

	mapped := WithMap[string, int](
		baseCache,
		func(v int) int { return v * 100 },
	)

	collected := make(map[string]int)
	mapped.Range(func(k string, v int) bool {
		collected[k] = v
		return true
	})

	expected := map[string]int{
		"a": 100,
		"b": 200,
		"c": 300,
	}

	if len(collected) != len(expected) {
		t.Errorf("Expected %d items, got %d", len(expected), len(collected))
	}

	for k, v := range expected {
		if collected[k] != v {
			t.Errorf("Expected %s=%d, got %d", k, v, collected[k])
		}
	}
}

func TestWithMap_ChainedMappers(t *testing.T) {
	baseCache := strategies.NewLRUCache[string, int](10)
	baseCache.Set("num", 10)

	mapped := WithMap[string, int](
		WithMap[string, int](
			WithMap[string, int](
				baseCache,
				func(v int) int { return v + 5 },
			),
			func(v int) int { return v * 2 },
		),
		func(v int) int { return v - 10 },
	)

	val, err := mapped.Get("num")
	if err != nil {
		t.Fatalf("Get(num) failed: %v", err)
	}
	if val != 20 {
		t.Errorf("Expected 20, got %d", val)
	}
}

func TestWithMap_EmptyCache(t *testing.T) {
	baseCache := strategies.NewLRUCache[string, int](10)

	mapped := WithMap[string, int](
		baseCache,
		func(v int) int { return v * 2 },
	)

	_, err := mapped.Get("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent key")
	}
}

func TestWithMap_SetStoresOriginalValue(t *testing.T) {
	baseCache := strategies.NewLRUCache[string, int](10)
	mapped := WithMap[string, int](
		baseCache,
		func(v int) int { return v * 10 },
	)

	err := mapped.Set("key", 5)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	originalVal, err := baseCache.Get("key")
	if err != nil || originalVal != 5 {
		t.Errorf("Original cache should have value 5, got %d", originalVal)
	}

	mappedVal, err := mapped.Get("key")
	if err != nil || mappedVal != 50 {
		t.Errorf("Mapped cache should have value 50, got %d", mappedVal)
	}
}

func TestWithMap_WithFilter(t *testing.T) {
	baseCache := strategies.NewLRUCache[string, int](10)
	for i := 1; i <= 10; i++ {
		baseCache.Set(string(rune('a'+i-1)), i)
	}

	composed := WithMap[string, int](
		WithFilter[string, int](
			baseCache,
			func(v int) bool { return v%2 == 0 },
		),
		func(v int) int { return v * 2 },
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
	baseCache.Set("a", 1)
	baseCache.Set("b", 2)
	baseCache.Set("c", 3)

	mapped := WithMap[string, int](
		baseCache,
		func(v int) int { return v * 10 },
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

	baseCache.Set("a", 999)
	val, _ := snapshot.Get("a")
	if val != 10 {
		t.Error("Snapshot should be independent of source cache")
	}
}
