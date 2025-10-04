package decorators

import (
	"caching-labwork/cache"
	"caching-labwork/cache/strategies/common"
	"errors"
	"sync/atomic"
)

type MetricsDecorator[K comparable, V any] struct {
	cacheWrappee cache.Cache[K, V]

	hits   atomic.Int64
	misses atomic.Int64
	evicts atomic.Int64

	rawBytesNum        atomic.Int64
	compressedBytesNum atomic.Int64
}

func WithMetrics[K comparable, V any](wrappee cache.Cache[K, V]) *MetricsDecorator[K, V] {
	decorator := &MetricsDecorator[K, V]{
		cacheWrappee: wrappee,
	}
	if observable, ok := wrappee.(cache.ObservableCache[K, V]); ok {
		observable.OnEvent(func(event cache.Event[K, V]) {
			switch event.Type {
			case cache.EventTypeHit:
				decorator.hits.Add(1)
			case cache.EventTypeMiss:
				decorator.misses.Add(1)
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

func (m *MetricsDecorator[K, V]) HitRate() float64 {
	hits := m.hits.Load()
	misses := m.misses.Load()
	total := hits + misses
	if total == 0 {
		return 0.0
	}
	return float64(hits) / float64(total)
}

func (m *MetricsDecorator[K, V]) CompressionRate() float64 {
	raw := m.rawBytesNum.Load()
	compressed := m.compressedBytesNum.Load()
	if raw == 0.0 {
		return 0.0
	}
	return float64(compressed) / float64(raw)
}

func (m *MetricsDecorator[K, V]) GetHits() int64 {
	return m.hits.Load()
}

func (m *MetricsDecorator[K, V]) GetMisses() int64 {
	return m.misses.Load()
}

func (m *MetricsDecorator[K, V]) GetEvictions() int64 {
	return m.evicts.Load()
}

func (w *MetricsDecorator[K, V]) Get(key K) (V, error) {
	v, err := w.cacheWrappee.Get(key)
	if err == nil {
		w.hits.Add(1)
	}
	if errors.Is(err, common.ErrKeyNotFound) {
		w.misses.Add(1)
	}
	return v, err
}

func (w *MetricsDecorator[K, V]) Set(key K, value V) error {
	return w.cacheWrappee.Set(key, value)
}

func (w *MetricsDecorator[K, V]) Delete(key K) error {
	err := w.cacheWrappee.Delete(key)
	if errors.Is(err, common.ErrKeyNotFound) {
		w.misses.Add(1)
	}
	return err
}

func (w *MetricsDecorator[K, V]) Clear() {
	w.cacheWrappee.Clear()
}
