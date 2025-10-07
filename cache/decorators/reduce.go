package decorators

import (
	"github.com/kimvlry/caching/cache"
)

func WithReduce[K comparable, V any, R any](
	wrappee cache.IterableCache[K, V],
	initial R,
	reducer func(acc R, value V) R,
) R {
	accumulator := initial
	wrappee.Range(func(k K, v V) bool {
		accumulator = reducer(accumulator, v)
		return true
	})
	return accumulator
}
