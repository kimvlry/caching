package decorators

import (
	"fmt"
	"github.com/kimvlry/caching/cache"
)

type filteredCache[K comparable, V any] struct {
	source cache.IterableCache[K, V]
	pred   func(V) bool
}

func WithFilter[K comparable, V any](
	wrappee cache.IterableCache[K, V],
	pred func(V) bool,
) cache.IterableCache[K, V] {

	return &filteredCache[K, V]{
		source: wrappee,
		pred:   pred,
	}
}

func (f *filteredCache[K, V]) Get(key K) (V, error) {
	val, err := f.source.Get(key)
	if err != nil {
		return val, err
	}

	if !f.pred(val) {
		var zero V
		return zero, fmt.Errorf("key %v filtered out", key)
	}

	return val, nil
}

func (f *filteredCache[K, V]) Set(key K, value V) error {
	if !f.pred(value) {
		return fmt.Errorf("value for key %v does not pass filter", key)
	}
	return f.source.Set(key, value)
}

func (f *filteredCache[K, V]) Delete(key K) error {
	return f.source.Delete(key)
}

func (f *filteredCache[K, V]) Clear() {
	f.source.Clear()
}

func (f *filteredCache[K, V]) Range(fn func(k K, v V) bool) {
	f.source.Range(func(k K, v V) bool {
		if f.pred(v) {
			return fn(k, v)
		}
		return true
	})
}
