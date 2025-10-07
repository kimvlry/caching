package decorators

import (
	"fmt"
	"github.com/kimvlry/caching/cache"
	"github.com/kimvlry/caching/cache/strategies"
)

func WithFilter[K comparable, V any](
	base cache.IterableCache[K, V],
	pred func(V) bool,
	factory strategies.CacheFactory[K, V],
) cache.IterableCache[K, V] {

	filtered := factory()
	base.Range(func(k K, v V) bool {
		if pred(v) {
			if err := filtered.Set(k, v); err != nil {
				panic(fmt.Sprintf("error setting filtered item: %v", err))
			}
		}
		return true
	})
	return filtered
}
