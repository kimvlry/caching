package decorators

import (
	"fmt"
	"github.com/kimvlry/caching/cache"
	"log/slog"
)

type loggingDecorator[K comparable, V any] struct {
	cacheWrappee cache.Cache[K, V]
	logger       *slog.Logger
}

func WithDebugLogging[K comparable, V any](cache cache.Cache[K, V], logger *slog.Logger) cache.Cache[K, V] {
	logger = logger.With("cache", fmt.Sprintf("%T", cache))
	return &loggingDecorator[K, V]{
		cacheWrappee: cache,
		logger:       logger,
	}
}

func (w *loggingDecorator[K, V]) Get(key K) (V, error) {
	w.logger.Debug("Get method called", "key", key)
	val, err := w.cacheWrappee.Get(key)
	if err != nil {
		w.logger.Warn("Get method returned an error", "key", key, "err", err)
	}
	w.logger.Debug("Get method returned a value", "key", key, "val", val)
	return val, err
}

func (w *loggingDecorator[K, V]) Set(key K, value V) error {
	w.logger.Debug("Set method called", "key", key)
	err := w.cacheWrappee.Set(key, value)
	if err != nil {
		w.logger.Warn("Set method returned an error", "key", key, "err", err)
	}
	w.logger.Debug("Set method returned a value", "key", key, "val", value)
	return err
}

func (w *loggingDecorator[K, V]) Delete(key K) error {
	w.logger.Debug("Delete method called", "key", key)
	err := w.cacheWrappee.Delete(key)
	if err != nil {
		w.logger.Warn("Delete method returned an error", "key", key, "err", err)
	}
	w.logger.Debug("Delete method returned a value", "key", key, "val", key)
	return err
}

func (w *loggingDecorator[K, V]) Clear() {
	w.logger.Debug("Clear method called")
	w.cacheWrappee.Clear()
}
