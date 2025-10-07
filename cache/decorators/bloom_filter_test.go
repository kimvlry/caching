package decorators

import (
	"errors"
	"fmt"
	"testing"

	"github.com/kimvlry/caching/cache"
	"github.com/kimvlry/caching/cache/strategies"
	"github.com/kimvlry/caching/cache/strategies/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Product struct {
	Name  string
	Price int
}

var discounter = func(p Product) Product {
	return Product{
		Name:  p.Name,
		Price: int(float64(p.Price) * 0.9),
	}
}

var expensiveOnly = func(p Product) bool {
	return p.Price > 50
}

var priceReducer = func(acc int, p Product) int {
	return acc + p.Price
}

func TestBloomFilter_BasicUsage(t *testing.T) {
	base := strategies.NewLFUCache[string, int](10)
	cacheWithBF := WithBloomFilter(base, 10, 0.01)

	err := cacheWithBF.Set("a", 1)
	require.NoError(t, err)

	val, err := cacheWithBF.Get("a")
	require.NoError(t, err)
	assert.Equal(t, 1, val)

	_, err = cacheWithBF.Get("missing")
	assert.True(t, errors.Is(err, common.ErrKeyNotFound))
}

func TestBloomFilter_Clear(t *testing.T) {
	base := strategies.NewLFUCache[string, int](10)
	cacheWithBF := WithBloomFilter(base, 10, 0.01)

	_ = cacheWithBF.Set("x", 10)
	_ = cacheWithBF.Set("y", 20)

	cacheWithBF.Clear()

	_, err := cacheWithBF.Get("x")
	assert.True(t, errors.Is(err, common.ErrKeyNotFound))
}

func TestBloomFilter_CombinedDecorators(t *testing.T) {
	chain := WithMetrics(
		WithFilter(
			WithMap[string, Product](
				WithBloomFilter(
					strategies.NewLFUCache[string, Product](10),
					100,
					0.01,
				),
				discounter,
			),
			expensiveOnly,
		),
	)

	_ = chain.Set("p1", Product{Name: "Laptop", Price: 1000})
	_ = chain.Set("p2", Product{Name: "Mouse", Price: 20})
	_ = chain.Set("p3", Product{Name: "Monitor", Price: 300})
	_ = chain.Set("p4", Product{Name: "Keyboard", Price: 80})

	v1, err := chain.Get("p1")
	require.NoError(t, err)
	assert.True(t, v1.Price > 50)

	_, err = chain.Get("p2")
	assert.Error(t, err)

	mAware := chain.(AwareCache[string, Product])
	assert.GreaterOrEqual(t, mAware.GetHits(), int64(1))
	assert.GreaterOrEqual(t, mAware.GetMisses(), int64(1))
}

func TestBloomFilter_WithReduceAndMetrics(t *testing.T) {
	bf := WithBloomFilter(strategies.NewLFUCache[string, Product](10), 1000, 0.01)
	iterableBF, ok := any(bf).(cache.IterableCache[string, Product])
	require.True(t, ok)

	mapped := WithMap[string, Product](iterableBF, discounter)
	filtered := WithFilter(mapped, expensiveOnly)
	chain := WithMetrics(filtered)

	_ = chain.Set("a", Product{"ItemA", 100})
	_ = chain.Set("b", Product{"ItemB", 200})
	_ = chain.Set("c", Product{"ItemC", 50})

	total := WithReduce[string, Product, int](filtered, 0, priceReducer)
	assert.Equal(t, 270, total)

	mAware := chain.(AwareCache[string, Product])
	fmt.Printf("Hits: %d, Misses: %d, HitRate: %.2f%%\n",
		mAware.GetHits(), mAware.GetMisses(), mAware.HitRate()*100)
}

// TODO: benchmarks
