package common

import (
	"fmt"
	"github.com/kimvlry/caching/cache"
)

func Snapshot[K comparable, V any](
	source cache.IterableCache[K, V],
	factory func() cache.IterableCache[K, V],
) cache.IterableCache[K, V] {
	result := factory()

	source.Range(func(k K, v V) bool {
		if err := result.Set(k, v); err != nil {
			panic(fmt.Sprintf("Snapshot: unexpected error setting key %v: %v", k, err))
		}
		return true
	})

	return result
}
