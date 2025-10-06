package decorators

import (
	"fmt"
	"github.com/kimvlry/caching/cache/decorators/common"
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
}

func TestWithFilter_FilterAll(t *testing.T) {
	baseCache := strategies.NewLRUCache[string, int](10)
	_ = baseCache.Set("key1", 1)
	_ = baseCache.Set("key2", 2)
	_ = baseCache.Set("key3", 3)

	filtered := WithFilter(
		baseCache,
		func(v int) bool { return v > 100 },
	)

	for _, key := range []string{"key1", "key2", "key3"} {
		if _, err := filtered.Get(key); err == nil {
			t.Errorf("key %s should be filtered out", key)
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
}

func TestWithFilter_EmptyCache(t *testing.T) {
	baseCache := strategies.NewLRUCache[string, int](10)

	filtered := WithFilter(
		baseCache,
		func(v int) bool { return v > 0 },
	)

	_, err := filtered.Get("anykey")
	if err == nil {
		t.Error("Expected error for non-existent key in empty cache")
	}
}

func TestWithFilter_ImmutabilitySourceCache(t *testing.T) {
	baseCache := strategies.NewLRUCache[string, int](10)
	_ = baseCache.Set("key1", 5)
	_ = baseCache.Set("key2", 15)
	_ = baseCache.Set("key3", 25)

	val1, err1 := baseCache.Get("key1")
	if err1 != nil || val1 != 5 {
		t.Error("Source cache was mutated! key1 should still exist with value 5")
	}

	val2, err2 := baseCache.Get("key2")
	if err2 != nil || val2 != 15 {
		t.Error("Source cache was mutated! key2 should still exist with value 15")
	}

	val3, err3 := baseCache.Get("key3")
	if err3 != nil || val3 != 25 {
		t.Error("Source cache was mutated! key3 should still exist with value 25")
	}
}

func TestWithFilter_ComplexPredicate(t *testing.T) {
	type Product struct {
		ID      int
		Name    string
		Price   int
		InStock bool
	}

	productCache := strategies.NewLRUCache[string, Product](10)
	_ = productCache.Set("p1", Product{ID: 1, Name: "Laptop", Price: 1000, InStock: true})
	_ = productCache.Set("p2", Product{ID: 2, Name: "Mouse", Price: 20, InStock: false})
	_ = productCache.Set("p3", Product{ID: 3, Name: "Keyboard", Price: 75, InStock: true})
	_ = productCache.Set("p4", Product{ID: 4, Name: "Monitor", Price: 300, InStock: true})

	filtered := WithFilter(
		productCache,
		func(p Product) bool {
			return p.InStock && p.Price >= 100
		},
	)

	if p, err := filtered.Get("p1"); err != nil {
		t.Error("p1 should pass filter")
	} else if p.Name != "Laptop" {
		t.Error("p1 should be Laptop")
	}

	if _, err := filtered.Get("p2"); err == nil {
		t.Error("p2 should be filtered out (not in stock)")
	}

	if _, err := filtered.Get("p3"); err == nil {
		t.Error("p3 should be filtered out (price too low)")
	}

	if p, err := filtered.Get("p4"); err != nil {
		t.Error("p4 should pass filter")
	} else if p.Name != "Monitor" {
		t.Error("p4 should be Monitor")
	}
}

func TestWithFilter_ChainedFilters(t *testing.T) {
	baseCache := strategies.NewLRUCache[string, int](100)
	for i := 1; i <= 100; i++ {
		_ = baseCache.Set(fmt.Sprintf("num%d", i), i)
	}

	finalFilter := WithFilter(
		WithFilter(
			WithFilter(
				baseCache,
				func(v int) bool { return v%2 == 0 },
			),
			func(v int) bool { return v > 50 },
		),
		func(v int) bool { return v <= 80 },
	)

	expectedValues := []int{52, 54, 56, 58, 60, 62, 64, 66, 68, 70, 72, 74, 76, 78, 80}

	for _, expected := range expectedValues {
		key := fmt.Sprintf("num%d", expected)
		val, err := finalFilter.Get(key)
		if err != nil {
			t.Errorf("%s should be in final filter", key)
			continue
		}
		if val != expected {
			t.Errorf("Expected %d, got %d", expected, val)
		}
	}

	shouldNotExist := []int{1, 50, 51, 81, 100}
	for _, num := range shouldNotExist {
		key := fmt.Sprintf("num%d", num)
		if _, err := finalFilter.Get(key); err == nil {
			t.Errorf("%s should NOT be in final filter", key)
		}
	}
}

func TestWithFilter_DifferentCacheTypes(t *testing.T) {
	lruCache := strategies.NewLRUCache[string, int](10)
	_ = lruCache.Set("a", 1)
	_ = lruCache.Set("b", 2)
	_ = lruCache.Set("c", 3)

	filtered := WithFilter(
		lruCache,
		func(v int) bool { return v > 1 },
	)

	if _, err := filtered.Get("a"); err == nil {
		t.Error("'a' should be filtered out")
	}

	val, err := filtered.Get("b")
	if err != nil || val != 2 {
		t.Errorf("'b' should exist with value 2, got %d, err=%v", val, err)
	}

	val, err = filtered.Get("c")
	if err != nil || val != 3 {
		t.Errorf("'c' should exist with value 3, got %d, err=%v", val, err)
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

func TestWithFilter_LargeDataset(t *testing.T) {
	baseCache := strategies.NewLRUCache[string, int](10000)
	for i := 0; i < 10000; i++ {
		_ = baseCache.Set(fmt.Sprintf("key%d", i), i)
	}

	filtered := WithFilter(
		baseCache,
		func(v int) bool { return v%100 == 0 },
	)

	count := 0
	for i := 0; i < 10000; i += 100 {
		key := fmt.Sprintf("key%d", i)
		if val, err := filtered.Get(key); err == nil {
			count++
			if val != i {
				t.Errorf("key%d: expected %d, got %d", i, i, val)
			}
		}
	}

	if count != 100 {
		t.Errorf("Expected 100 elements in filtered cache, got %d", count)
	}
}

func TestWithFilter_Range(t *testing.T) {
	baseCache := strategies.NewLRUCache[string, int](10)
	_ = baseCache.Set("a", 1)
	_ = baseCache.Set("b", 2)
	_ = baseCache.Set("c", 3)
	_ = baseCache.Set("d", 4)
	_ = baseCache.Set("e", 5)

	filtered := WithFilter(
		baseCache,
		func(v int) bool { return v%2 == 0 },
	)

	collected := make(map[string]int)
	filtered.Range(func(k string, v int) bool {
		collected[k] = v
		return true
	})

	expected := map[string]int{
		"b": 2,
		"d": 4,
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

func TestWithFilter_SetRejectsFilteredValues(t *testing.T) {
	baseCache := strategies.NewLRUCache[string, int](10)
	filtered := WithFilter(
		baseCache,
		func(v int) bool { return v > 10 },
	)

	err := filtered.Set("key1", 5)
	if err == nil {
		t.Error("Set should reject value that doesn't pass filter")
	}

	err = filtered.Set("key2", 15)
	if err != nil {
		t.Errorf("Set should accept value that passes filter: %v", err)
	}

	val, err := filtered.Get("key2")
	if err != nil || val != 15 {
		t.Errorf("Expected key2=15, got %d, err=%v", val, err)
	}
}

func TestSnapshot_MaterializesFilteredCache(t *testing.T) {
	baseCache := strategies.NewLRUCache[string, int](10)
	for i := 1; i <= 10; i++ {
		_ = baseCache.Set(fmt.Sprintf("key%d", i), i)
	}

	filtered := WithFilter(
		baseCache,
		func(v int) bool { return v > 5 },
	)

	snapshot := common.Snapshot(
		filtered,
		func() cache.IterableCache[string, int] {
			return strategies.NewLRUCache[string, int](5)
		},
	)

	for i := 1; i <= 5; i++ {
		key := fmt.Sprintf("key%d", i)
		if _, err := snapshot.Get(key); err == nil {
			t.Errorf("%s should not be in snapshot", key)
		}
	}

	for i := 6; i <= 10; i++ {
		key := fmt.Sprintf("key%d", i)
		val, err := snapshot.Get(key)
		if err != nil {
			t.Errorf("%s should be in snapshot", key)
			continue
		}
		if val != i {
			t.Errorf("Expected %s=%d, got %d", key, i, val)
		}
	}

	_ = baseCache.Set("key6", 100)
	val, _ := snapshot.Get("key6")
	if val != 6 {
		t.Error("Snapshot should be independent of source cache")
	}
}
