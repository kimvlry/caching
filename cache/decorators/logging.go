package decorators

import (
	"caching-labwork/cache"
	"fmt"
	"log/slog"
)

type LoggingDecorator[K comparable, V any] struct {
	cacheWrappee cache.Cache[K, V]
	logger       *slog.Logger
}

func WithDebugLogging[K comparable, V any](cache cache.Cache[K, V], logger *slog.Logger) *LoggingDecorator[K, V] {
	logger = logger.With("cache", fmt.Sprintf("%T", cache))
	return &LoggingDecorator[K, V]{
		cacheWrappee: cache,
		logger:       logger,
	}
}

func (w *LoggingDecorator[K, V]) Get(key K) (V, error) {
	w.logger.Debug("Get method called", "key", key)
	val, err := w.cacheWrappee.Get(key)
	if err != nil {
		w.logger.Warn("Get method returned an error", "key", key, "err", err)
	}
	w.logger.Debug("Get method returned a value", "key", key, "val", val)
	return val, nil
}

func (w *LoggingDecorator[K, V]) Set(key K, value V) error {
	w.logger.Debug("Set method called", "key", key)
	err := w.cacheWrappee.Set(key, value)
	if err != nil {
		w.logger.Warn("Set method returned an error", "key", key, "err", err)
	}
	w.logger.Debug("Set method returned a value", "key", key, "val", value)
	return nil
}

func (w *LoggingDecorator[K, V]) Delete(key K) error {
	w.logger.Debug("Delete method called", "key", key)
	err := w.cacheWrappee.Delete(key)
	if err != nil {
		w.logger.Warn("Delete method returned an error", "key", key, "err", err)
	}
	w.logger.Debug("Delete method returned a value", "key", key, "val", key)
	return nil
}

func (w *LoggingDecorator[K, V]) Clear() {
	w.logger.Debug("Clear method called")
	w.cacheWrappee.Clear()
}
