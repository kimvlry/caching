package strategies

import (
	"github.com/kimvlry/caching/cache"
	"github.com/kimvlry/caching/cache/strategies/common"
)

// fifoCache implements a First In, First Out cache
type fifoCache[K comparable, V any] struct {
	capacity int
	data     map[K]V
	keys     []K

	eventCallbacks []func(cache.Event[K, V])
}

func newFifoCache[K comparable, V any](capacity int) cache.IterableCache[K, V] {
	return &fifoCache[K, V]{
		capacity: capacity,
		data:     make(map[K]V, capacity),
		keys:     make([]K, 0, capacity),
	}
}

// Get retrieves a value by key. If key not found, returns zero value and error
func (f *fifoCache[K, V]) Get(key K) (V, error) {
	if value, exists := f.data[key]; exists {
		return value, nil
	}
	var zero V
	return zero, common.ErrKeyNotFound
}

// Set adds or updates a key-value pair. If cache is full, the oldest pq_item gets evicted (first in)
func (f *fifoCache[K, V]) Set(key K, value V) error {
	if _, exists := f.data[key]; exists {
		f.data[key] = value
		return nil
	}

	if len(f.data) >= f.capacity {
		oldestKey := f.keys[0]
		delete(f.data, oldestKey)
		f.keys = f.keys[1:]

		f.emit(cache.Event[K, V]{
			Type:  cache.EventTypeEviction,
			Key:   oldestKey,
			Value: value,
		})
	}

	f.data[key] = value
	f.keys = append(f.keys, key)
	return nil
}

// Delete removes a key-value pair. Returns error if key not found
func (f *fifoCache[K, V]) Delete(key K) error {
	if _, exists := f.data[key]; !exists {
		return common.ErrKeyNotFound
	}

	delete(f.data, key)

	for i, k := range f.keys {
		if k == key {
			f.keys = append(f.keys[:i], f.keys[i+1:]...)
			break
		}
	}

	return nil
}

// Clear removes all key-value pairs
func (f *fifoCache[K, V]) Clear() {
	f.data = make(map[K]V, f.capacity)
	f.keys = make([]K, 0)
}

func (f *fifoCache[K, V]) OnEvent(callback func(event cache.Event[K, V])) {
	f.eventCallbacks = append(f.eventCallbacks, callback)
}

func (f *fifoCache[K, V]) emit(event cache.Event[K, V]) {
	for _, callback := range f.eventCallbacks {
		callback(event)
	}
}

func (f *fifoCache[K, V]) Range(fn func(K, V) bool) {
	for k, v := range f.data {
		if !fn(k, v) {
			break
		}
	}
}
