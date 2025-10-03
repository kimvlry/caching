package decorators

import (
	"caching-labwork/cache"
	"caching-labwork/cache/cache_decorators/metrics"
	"caching-labwork/cache/common"
	"errors"
)

// TODO: collect evictions

type WithMetrics[K comparable, V any] struct {
	cacheWrappee cache.Cache[K, V]
	Collector    metrics.Collector
}

func NewWithMetrics[K comparable, V any](cache cache.Cache[K, V]) *WithMetrics[K, V] {
	return &WithMetrics[K, V]{
		cacheWrappee: cache,
	}
}

func (w WithMetrics[K, V]) Get(key K) (V, error) {
	val, err := w.cacheWrappee.Get(key)
	if errors.Is(err, common.ErrKeyNotFound) {
		w.Collector.RecordMiss()
	}
	if err == nil {
		w.Collector.RecordHit()
	}
	return val, err
}

func (w WithMetrics[K, V]) Set(key K, value V) error {
	return w.cacheWrappee.Set(key, value)
}

func (w WithMetrics[K, V]) Delete(key K) error {
	return w.cacheWrappee.Delete(key)
}

func (w WithMetrics[K, V]) Clear() {
	w.cacheWrappee.Clear()
}
