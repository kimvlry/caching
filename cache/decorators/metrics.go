package decorators

import (
	"caching-labwork/cache"
	"caching-labwork/cache/decorators/metrics"
	"caching-labwork/cache/strategies/common"
	"errors"
)

// TODO: collect evictions

type MetricsDecorator[K comparable, V any] struct {
	cacheWrappee cache.Cache[K, V]
	Collector    metrics.Collector
}

func WithMetrics[K comparable, V any](cache cache.Cache[K, V]) *MetricsDecorator[K, V] {
	return &MetricsDecorator[K, V]{
		cacheWrappee: cache,
	}
}

func (w MetricsDecorator[K, V]) Get(key K) (V, error) {
	val, err := w.cacheWrappee.Get(key)
	if errors.Is(err, common.ErrKeyNotFound) {
		w.Collector.RecordMiss()
	}
	if err == nil {
		w.Collector.RecordHit()
	}
	return val, err
}

func (w MetricsDecorator[K, V]) Set(key K, value V) error {
	return w.cacheWrappee.Set(key, value)
}

func (w MetricsDecorator[K, V]) Delete(key K) error {
	return w.cacheWrappee.Delete(key)
}

func (w MetricsDecorator[K, V]) Clear() {
	w.cacheWrappee.Clear()
}
