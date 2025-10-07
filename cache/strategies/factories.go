package strategies

import (
	"github.com/kimvlry/caching/cache"
	"time"
)

// CacheFactory is an alias for cache factory function
type CacheFactory[K comparable, V any] func() cache.IterableCache[K, V]

func NewLruCache[K comparable, V any](capacity int) CacheFactory[K, V] {
	return func() cache.IterableCache[K, V] {
		return newLruCache[K, V](capacity)
	}
}

func NewLfuCache[K comparable, V any](capacity int) CacheFactory[K, V] {
	return func() cache.IterableCache[K, V] {
		return newLfuCache[K, V](capacity)
	}
}

func NewTtlCache[K comparable, V any](capacity int, ttl time.Duration) CacheFactory[K, V] {
	return func() cache.IterableCache[K, V] {
		return newTtlCache[K, V](capacity, ttl)
	}
}

func NewFifoCache[K comparable, V any](capacity int) CacheFactory[K, V] {
	return func() cache.IterableCache[K, V] {
		return newFifoCache[K, V](capacity)
	}
}
