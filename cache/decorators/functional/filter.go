package functional

import (
	"caching-labwork/cache"
	"fmt"
)

func WithFilter[K comparable, V any](
	wrappee cache.IterableCache[K, V],
	pred func(V) bool,
	factory CacheFactory[K, V],
) cache.Cache[K, V] {

	res := factory()
	wrappee.Range(func(k K, v V) bool {
		if pred(v) {
			if err := res.Set(k, v); err != nil {
				panic(fmt.Sprintf("WithFilter: unexpected error setting key: %v: %v: ", k, err.Error()))
			}
		}
		return true
	})
	return res
}
