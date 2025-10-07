package decorators

import (
	"strings"
	"testing"

	"github.com/kimvlry/caching/cache"
	"github.com/kimvlry/caching/cache/strategies"
)

func TestWithFilter_BasicFiltering(t *testing.T) {
	baseCache := strategies.NewLRUCache[string, int](10)
	_ = baseCache.Set("key1", 5)
	_ = baseCache.Set("key2", 15)
	_ = baseCache.Set("key3", 25)
	_ = baseCache.Set("key4", 10)

	filtered := WithFilter(
		baseCache,
		func(v int) bool { return v > 10 },
		func() cache.IterableCache[string, int] {
			return strategies.NewLRUCache[string, int](10)
		},
	)

	if _, err := filtered.Get("key1"); err == nil {
		t.Error("key1 (value=5) should be filtered out")
	}

	val2, err := filtered.Get("key2")
	if err != nil || val2 != 15 {
		t.Errorf("key2 should be present with value 15, got %d, err=%v", val2, err)
	}

	val3, err := filtered.Get("key3")
	if err != nil || val3 != 25 {
		t.Errorf("key3 should be present with value 25, got %d, err=%v", val3, err)
	}

	if _, err := filtered.Get("key4"); err == nil {
		t.Error("key4 (value=10) should be filtered out")
	}

	for k, expected := range map[string]int{"key1": 5, "key2": 15, "key3": 25, "key4": 10} {
		val, _ := baseCache.Get(k)
		if val != expected {
			t.Errorf("Base cache mutated! key=%v, val=%v", k, val)
		}
	}
}

func TestWithFilter_FilterAll(t *testing.T) {
	baseCache := strategies.NewLRUCache[string, int](10)
	_ = baseCache.Set("key1", 1)
	_ = baseCache.Set("key2", 2)
	_ = baseCache.Set("key3", 3)

	filtered := WithFilter(
		baseCache,
		func(v int) bool { return v > 100 },
		func() cache.IterableCache[string, int] {
			return strategies.NewLRUCache[string, int](10)
		},
	)

	for _, key := range []string{"key1", "key2", "key3"} {
		if _, err := filtered.Get(key); err == nil {
			t.Errorf("key %s should be filtered out", key)
		}
	}

	for i, key := range []string{"key1", "key2", "key3"} {
		val, _ := baseCache.Get(key)
		if val != i+1 {
			t.Errorf("Base cache mutated! key=%s, val=%d", key, val)
		}
	}
}

func TestWithFilter_FilterNone(t *testing.T) {
	baseCache := strategies.NewLRUCache[string, int](10)
	_ = baseCache.Set("key1", 1)
	_ = baseCache.Set("key2", 2)
	_ = baseCache.Set("key3", 3)

	filtered := WithFilter(
		baseCache,
		func(v int) bool { return true },
		func() cache.IterableCache[string, int] {
			return strategies.NewLRUCache[string, int](10)
		},
	)

	testCases := []struct {
		key      string
		expected int
	}{
		{"key1", 1},
		{"key2", 2},
		{"key3", 3},
	}

	for _, tc := range testCases {
		val, err := filtered.Get(tc.key)
		if err != nil {
			t.Errorf("Get(%s) failed: %v", tc.key, err)
			continue
		}
		if val != tc.expected {
			t.Errorf("Get(%s): expected %d, got %d", tc.key, tc.expected, val)
		}
	}

	for _, tc := range testCases {
		val, _ := baseCache.Get(tc.key)
		if val != tc.expected {
			t.Errorf("Base cache mutated! key=%s, val=%d", tc.key, val)
		}
	}
}

func TestWithFilter_ImmutabilitySourceCache(t *testing.T) {
	baseCache := strategies.NewLRUCache[string, int](10)
	_ = baseCache.Set("key1", 5)
	_ = baseCache.Set("key2", 15)
	_ = baseCache.Set("key3", 25)

	filtered := WithFilter(
		baseCache,
		func(v int) bool { return v > 10 },
		func() cache.IterableCache[string, int] {
			return strategies.NewLRUCache[string, int](10)
		},
	)

	_ = filtered.Set("key5", 50)

	for k, expected := range map[string]int{"key1": 5, "key2": 15, "key3": 25} {
		val, _ := baseCache.Get(k)
		if val != expected {
			t.Errorf("Base cache mutated! key=%s, val=%d", k, val)
		}
	}
}

func TestWithFilter_And_Map_Chained(t *testing.T) {
	type Product struct {
		ID    int
		Name  string
		Price int
	}

	baseCache := strategies.NewLRUCache[string, Product](10)
	_ = baseCache.Set("p1", Product{1, "A", 50})
	_ = baseCache.Set("p2", Product{2, "B", 150})
	_ = baseCache.Set("p3", Product{3, "C", 200})

	filtered := WithFilter(
		baseCache,
		func(p Product) bool { return p.Price > 100 },
		func() cache.IterableCache[string, Product] {
			return strategies.NewLRUCache[string, Product](10)
		},
	)

	mapped := WithMap(
		filtered,
		func(p Product) Product {
			p.Price += 50
			return p
		},
		func() cache.IterableCache[string, Product] {
			return strategies.NewLRUCache[string, Product](10)
		},
	)

	if _, err := mapped.Get("p1"); err == nil {
		t.Error("p1 should be filtered out")
	}
	p2, _ := mapped.Get("p2")
	if p2.Price != 200 {
		t.Errorf("p2 price expected 200, got %d", p2.Price)
	}
	p3, _ := mapped.Get("p3")
	if p3.Price != 250 {
		t.Errorf("p3 price expected 250, got %d", p3.Price)
	}

	expectedProducts := map[string]Product{
		"p1": {ID: 1, Name: "A", Price: 50},
		"p2": {ID: 2, Name: "B", Price: 150},
		"p3": {ID: 3, Name: "C", Price: 200},
	}

	for k, expected := range expectedProducts {
		val, _ := baseCache.Get(k)
		if val != expected {
			t.Errorf("Base cache mutated! key=%s, val=%+v, expected=%+v", k, val, expected)
		}
	}

}

func TestWithFilter_StringFiltering(t *testing.T) {
	stringCache := strategies.NewLRUCache[string, string](10)
	_ = stringCache.Set("user:1", "alice@example.com")
	_ = stringCache.Set("user:2", "bob@gmail.com")
	_ = stringCache.Set("user:3", "charlie@example.com")
	_ = stringCache.Set("user:4", "david@yahoo.com")

	filtered := WithFilter(
		stringCache,
		func(email string) bool {
			return strings.HasSuffix(email, "@example.com")
		},
		func() cache.IterableCache[string, string] {
			return strategies.NewLRUCache[string, string](10)
		},
	)

	testCases := []struct {
		key         string
		shouldExist bool
	}{
		{"user:1", true},
		{"user:2", false},
		{"user:3", true},
		{"user:4", false},
	}

	for _, tc := range testCases {
		_, err := filtered.Get(tc.key)
		exists := err == nil

		if exists != tc.shouldExist {
			t.Errorf("%s: expected exists=%v, got exists=%v", tc.key, tc.shouldExist, exists)
		}
	}
}
