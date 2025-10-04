package functional

import (
	"caching-labwork/cache"
	"fmt"
)

func WithMap[K comparable, V any](
	wrappee cache.IterableCache[K, V],
	mapper func(V) V,
	factory CacheFactory[K, V],
) cache.Cache[K, V] {

	res := factory()
	wrappee.Range(func(k K, v V) bool {
		if err := res.Set(k, mapper(v)); err != nil {
			panic(fmt.Sprintf("WithMap: unexpected error setting key: %v: %v: ", k, err.Error()))
		}
		return true
	})
	return res
}
