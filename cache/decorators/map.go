package decorators

import (
	"fmt"
	"github.com/kimvlry/caching/cache"
	"github.com/kimvlry/caching/cache/strategies"
)

func WithMap[K comparable, V any](
	base cache.IterableCache[K, V],
	mapper func(V) V,
	factory strategies.CacheFactory[K, V],
) cache.IterableCache[K, V] {

	mapped := factory()
	base.Range(func(k K, v V) bool {
		if err := mapped.Set(k, mapper(v)); err != nil {
			panic(fmt.Sprintf("error setting key %v with value %v: %v", k, v, err))
		}
		return true
	})
	return mapped
}
