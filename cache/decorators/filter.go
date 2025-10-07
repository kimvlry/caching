package decorators

import (
	"fmt"
	"github.com/kimvlry/caching/cache"
)

func WithFilter[K comparable, V any](
	base cache.IterableCache[K, V],
	pred func(V) bool,
	factory func() cache.IterableCache[K, V],
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
