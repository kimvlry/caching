package strategies

import (
	"github.com/kimvlry/caching/cache"
	"github.com/kimvlry/caching/cache/strategies/common"
)

// lruCache implements a Least Recently Used cache
type lruCache[K comparable, V any] struct {
	capacity int
	data     map[K]V
	keys     []K

	eventCallbacks []func(cache.Event[K, V])
}

func NewLRUCache[K comparable, V any](capacity int) cache.Cache[K, V] {
	return &lruCache[K, V]{
		capacity: capacity,
		data:     make(map[K]V, capacity),
		keys:     make([]K, 0, capacity),
	}
}

func (l *lruCache[K, V]) Get(key K) (V, error) {
	if value, exists := l.data[key]; exists {
		l.moveKeyToEnd(key)
		return value, nil
	}

	var zero V
	return zero, common.ErrKeyNotFound
}

func (l *lruCache[K, V]) Set(key K, value V) error {
	if _, exists := l.data[key]; exists {
		l.moveKeyToEnd(key)
		l.data[key] = value
		return nil
	}

	if len(l.data) >= l.capacity {
		leastRecentKey := l.keys[0]
		delete(l.data, leastRecentKey)
		l.keys = l.keys[1:]

		l.emit(cache.Event[K, V]{
			Type:  cache.EventTypeEviction,
			Key:   leastRecentKey,
			Value: value,
		})
	}

	l.keys = append(l.keys, key)
	l.data[key] = value
	return nil
}

func (l *lruCache[K, V]) Delete(key K) error {
	if _, exists := l.data[key]; !exists {
		return common.ErrKeyNotFound
	}

	delete(l.data, key)

	for i, k := range l.keys {
		if k == key {
			l.keys = append(l.keys[:i], l.keys[i+1:]...)
			break
		}
	}
	return nil
}

func (l *lruCache[K, V]) Clear() {
	l.data = make(map[K]V, l.capacity)
	l.keys = make([]K, 0)
}

func (l *lruCache[K, V]) Range(fn func(K, V) bool) {
	for k, v := range l.data {
		if !fn(k, v) {
			break
		}
	}
}

func (l *lruCache[K, V]) OnEvent(callback func(event cache.Event[K, V])) {
	l.eventCallbacks = append(l.eventCallbacks, callback)
}

func (l *lruCache[K, V]) emit(event cache.Event[K, V]) {
	for _, callback := range l.eventCallbacks {
		callback(event)
	}
}

func (l *lruCache[K, V]) moveKeyToEnd(key K) {
	for i, k := range l.keys {
		if k == key {
			l.keys = append(l.keys[:i], l.keys[i+1:]...)
			break
		}
	}
	l.keys = append(l.keys, key)
}
