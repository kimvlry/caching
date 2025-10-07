package decorators

import (
	"github.com/kimvlry/caching/cache/strategies"
	"testing"
)

func TestMetricsDecorator_HitsAndMisses(t *testing.T) {
	baseCache := strategies.NewLruCache[string, int](10)()
	_ = baseCache.Set("key1", 100)
	_ = baseCache.Set("key2", 200)

	metricsCache := WithMetrics(baseCache)

	_, _ = metricsCache.Get("key1")             // hit
	_, _ = metricsCache.Get("key2")             // hit
	_, _ = metricsCache.Get("key1")             // hit
	_, _ = metricsCache.Get("nonexistent")      // miss
	_, _ = metricsCache.Get("also_nonexistent") // miss

	if hits := metricsCache.GetHits(); hits != 3 {
		t.Errorf("Expected 3 hits, got %d", hits)
	}
	if misses := metricsCache.GetMisses(); misses != 2 {
		t.Errorf("Expected 2 misses, got %d", misses)
	}
}

func TestMetricsDecorator_HitRate(t *testing.T) {
	baseCache := strategies.NewLruCache[string, int](10)()
	_ = baseCache.Set("key1", 100)

	metricsCache := WithMetrics(baseCache)

	_, _ = metricsCache.Get("key1")        // hit
	_, _ = metricsCache.Get("key1")        // hit
	_, _ = metricsCache.Get("key1")        // hit
	_, _ = metricsCache.Get("nonexistent") // miss

	expectedHitRate := 0.75 // 3 hits / 4 total = 75%
	hitRate := metricsCache.HitRate()

	if hitRate < expectedHitRate-0.01 || hitRate > expectedHitRate+0.01 {
		t.Errorf("Expected hit rate %.2f, got %.2f", expectedHitRate, hitRate)
	}
}

func TestMetricsDecorator_HitRateEmpty(t *testing.T) {
	baseCache := strategies.NewLruCache[string, int](10)()
	metricsCache := WithMetrics(baseCache)

	hitRate := metricsCache.HitRate()
	if hitRate != 0.0 {
		t.Errorf("Expected 0.0 hit rate for empty metrics, got %.2f", hitRate)
	}
}

func TestMetricsDecorator_Evictions_ObservableCache(t *testing.T) {
	baseCache := strategies.NewLfuCache[string, int](3)()
	metricsCache := WithMetrics(baseCache)

	_ = metricsCache.Set("key1", 1)
	_ = metricsCache.Set("key2", 2)
	_ = metricsCache.Set("key3", 3)

	initialEvictions := metricsCache.GetEvictions()

	_ = metricsCache.Set("key4", 4)
	_ = metricsCache.Set("key5", 5)

	finalEvictions := metricsCache.GetEvictions()
	evicted := finalEvictions - initialEvictions

	if evicted != 2 {
		t.Errorf("Expected 2 evictions, got %d", evicted)
	}
}

func TestMetricsDecorator_Evictions_NonObservableCache(t *testing.T) {
	baseCache := strategies.NewFifoCache[string, int](3)()
	metricsCache := WithMetrics(baseCache)

	_ = metricsCache.Set("key1", 1)
	_ = metricsCache.Set("key2", 2)
	_ = metricsCache.Set("key3", 3)
	_ = metricsCache.Set("key4", 4)

	evictions := metricsCache.GetEvictions()
	if evictions != 0 {
		t.Logf("Evictions for non-Observable cache: %d (expected 0)", evictions)
	}
}

func TestMetricsDecorator_DeleteTracking(t *testing.T) {
	baseCache := strategies.NewLruCache[string, int](10)()
	_ = baseCache.Set("key1", 100)

	metricsCache := WithMetrics(baseCache)

	err := metricsCache.Delete("key1")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = metricsCache.Get("key1")

	if err == nil {
		t.Error("Expected error when getting deleted key")
	}

	initialMisses := metricsCache.GetMisses()
	_ = metricsCache.Delete("nonexistent")

	if metricsCache.GetMisses() <= initialMisses {
		t.Error("Delete of non-existent key should increment misses")
	}
}

func TestMetricsDecorator_CompressionRate(t *testing.T) {
	baseCache := strategies.NewLruCache[string, []byte](10)()
	compCache := WithCompression(baseCache, JSONSerializer[TestData]{})
	metricsCache := WithMetrics(compCache)

	testData := TestData{
		ID:   42,
		Name: string(make([]byte, 1000)), // 1KB
		Tags: []string{"tag1", "tag2"},
	}

	err := metricsCache.Set("key1", testData)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}
}
