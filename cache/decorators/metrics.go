package decorators

import (
	"errors"
	"github.com/kimvlry/caching/cache"
	"github.com/kimvlry/caching/cache/strategies/common"
	"sync/atomic"
)

type metricsDecorator[K comparable, V any] struct {
	cacheWrappee cache.Cache[K, V]

	hits   atomic.Int64
	misses atomic.Int64
	evicts atomic.Int64

	rawBytesNum        atomic.Int64
	compressedBytesNum atomic.Int64
}

type AwareCache[K comparable, V any] interface {
	cache.Cache[K, V]
	HitRate() float64
	GetHits() int64
	GetMisses() int64
	GetEvictions() int64
}

func WithMetrics[K comparable, V any](wrappee cache.Cache[K, V]) AwareCache[K, V] {
	decorator := &metricsDecorator[K, V]{
		cacheWrappee: wrappee,
	}
	if observable, ok := any(wrappee).(cache.ObservableCache[K, V]); ok {
		observable.OnEvent(func(event cache.Event[K, V]) {
			switch event.Type {
			case cache.EventTypeEviction:
				decorator.evicts.Add(1)
			case cache.EventTypeReadBytes:
				decorator.rawBytesNum.Add(int64(event.Size))
			case cache.EventTypeCompressBytes:
				decorator.compressedBytesNum.Add(int64(event.Size))
			}
		})
	}

	return decorator
}

func (m *metricsDecorator[K, V]) HitRate() float64 {
	hits := m.hits.Load()
	misses := m.misses.Load()
	total := hits + misses
	if total == 0 {
		return 0.0
	}
	return float64(hits) / float64(total)
}

func (m *metricsDecorator[K, V]) GetHits() int64 {
	return m.hits.Load()
}

func (m *metricsDecorator[K, V]) GetMisses() int64 {
	return m.misses.Load()
}

func (m *metricsDecorator[K, V]) GetEvictions() int64 {
	return m.evicts.Load()
}

func (w *metricsDecorator[K, V]) Get(key K) (V, error) {
	v, err := w.cacheWrappee.Get(key)
	if err == nil {
		w.hits.Add(1)
	}
	if errors.Is(err, common.ErrKeyNotFound) {
		w.misses.Add(1)
	}
	return v, err
}

func (w *metricsDecorator[K, V]) Set(key K, value V) error {
	return w.cacheWrappee.Set(key, value)
}

func (w *metricsDecorator[K, V]) Delete(key K) error {
	err := w.cacheWrappee.Delete(key)
	if errors.Is(err, common.ErrKeyNotFound) {
		w.misses.Add(1)
	}
	return err
}

func (w *metricsDecorator[K, V]) Clear() {
	w.cacheWrappee.Clear()
}

func (m *metricsDecorator[K, V]) Range(fn func(K, V) bool) {
	if iterable, ok := any(m.cacheWrappee).(cache.IterableCache[K, V]); ok {
		iterable.Range(fn)
	}
}
