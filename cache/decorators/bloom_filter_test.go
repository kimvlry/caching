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
	base := strategies.NewLfuCache[string, int](10)()
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
	base := strategies.NewLfuCache[string, int](10)()
	cacheWithBF := WithBloomFilter(base, 10, 0.01)

	_ = cacheWithBF.Set("x", 10)
	_ = cacheWithBF.Set("y", 20)

	cacheWithBF.Clear()

	_, err := cacheWithBF.Get("x")
	assert.True(t, errors.Is(err, common.ErrKeyNotFound))
}

func TestBloomFilter_CombinedDecorators(t *testing.T) {
	base := strategies.NewLfuCache[string, Product](10)()
	_ = base.Set("p1", Product{Name: "Laptop", Price: 1000})
	_ = base.Set("p2", Product{Name: "Mouse", Price: 20})
	_ = base.Set("p3", Product{Name: "Monitor", Price: 300})
	_ = base.Set("p4", Product{Name: "Keyboard", Price: 80})

	chain := WithMetrics(
		WithFilter(
			WithMap[string, Product](
				WithBloomFilter(base, 100, 0.01),
				discounter,
				func() cache.IterableCache[string, Product] {
					return strategies.NewLfuCache[string, Product](10)()
				},
			),
			expensiveOnly,
			func() cache.IterableCache[string, Product] {
				return strategies.NewLfuCache[string, Product](10)()
			},
		),
	)

	v1, err := chain.Get("p1")
	require.NoError(t, err)
	assert.Equal(t, 900, v1.Price)

	_, err = chain.Get("p2")
	assert.Error(t, err)

	v3, err := chain.Get("p3")
	require.NoError(t, err)
	assert.Equal(t, 270, v3.Price)

	v4, err := chain.Get("p4")
	require.NoError(t, err)
	assert.Equal(t, 72, v4.Price)

	mAware := chain.(AwareCache[string, Product])
	assert.GreaterOrEqual(t, mAware.GetHits(), int64(3))
	assert.GreaterOrEqual(t, mAware.GetMisses(), int64(1))
}

func TestBloomFilter_WithReduceAndMetrics(t *testing.T) {
	base := strategies.NewLfuCache[string, Product](10)()
	_ = base.Set("a", Product{"ItemA", 100})
	_ = base.Set("b", Product{"ItemB", 200})
	_ = base.Set("c", Product{"ItemC", 50})

	bf := WithBloomFilter(base, 1000, 0.01)

	iterableBF, ok := any(bf).(cache.IterableCache[string, Product])
	require.True(t, ok)

	mapped := WithMap[string, Product](
		iterableBF,
		discounter,
		func() cache.IterableCache[string, Product] {
			return strategies.NewLfuCache[string, Product](10)()
		},
	)

	filtered := WithFilter(
		mapped,
		expensiveOnly,
		func() cache.IterableCache[string, Product] {
			return strategies.NewLfuCache[string, Product](10)()
		},
	)

	chain := WithMetrics(filtered)

	total := WithReduce[string, Product, int](
		filtered,
		0,
		priceReducer,
	)

	assert.Equal(t, 270, total)

	mAware := chain.(AwareCache[string, Product])
	fmt.Printf("Hits: %d, Misses: %d, HitRate: %.2f%%\n",
		mAware.GetHits(), mAware.GetMisses(), mAware.HitRate()*100)
}

// TODO: benchmarks
