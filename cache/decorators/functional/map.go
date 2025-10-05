package functional

import (
	"github.com/kimvlry/caching/cache"
)

type mappedCache[K comparable, V any] struct {
	source cache.IterableCache[K, V]
	mapper func(V) V
}

func WithMap[K comparable, V any](
	wrappee cache.IterableCache[K, V],
	mapper func(V) V,
) cache.IterableCache[K, V] {
	return &mappedCache[K, V]{
		source: wrappee,
		mapper: mapper,
	}
}

func (m *mappedCache[K, V]) Get(key K) (V, error) {
	val, err := m.source.Get(key)
	if err != nil {
		return val, err
	}
	return m.mapper(val), nil
}

func (m *mappedCache[K, V]) Set(key K, value V) error {
	return m.source.Set(key, value)
}

func (m *mappedCache[K, V]) Delete(key K) error {
	return m.source.Delete(key)
}

func (m *mappedCache[K, V]) Clear() {
	m.source.Clear()
}

func (m *mappedCache[K, V]) Range(fn func(k K, v V) bool) {
	m.source.Range(func(k K, v V) bool {
		return fn(k, m.mapper(v))
	})
}
