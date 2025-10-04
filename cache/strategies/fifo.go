package strategies

import (
	"caching-labwork/cache"
	"caching-labwork/cache/strategies/common"
)

// FIFOCache implements a First In, First Out cache
type FIFOCache[K comparable, V any] struct {
	capacity int
	data     map[K]V
	keys     []K

	eventCallbacks []func(cache.Event[K, V])
}

// NewFIFOCache creates a new FIFO cache with the given capacity
func NewFIFOCache[K comparable, V any](capacity int) *FIFOCache[K, V] {
	return &FIFOCache[K, V]{
		capacity: capacity,
		data:     make(map[K]V, capacity),
		keys:     make([]K, 0),
	}
}

// Get retrieves a value by key. If key not found, returns zero value and error
func (F *FIFOCache[K, V]) Get(key K) (V, error) {
	if value, exists := F.data[key]; exists {
		return value, nil
	}
	var zero V
	return zero, common.ErrKeyNotFound
}

// Set adds or updates a key-value pair. If cache is full, the oldest pq_item gets evicted (first in)
func (F *FIFOCache[K, V]) Set(key K, value V) error {
	if _, exists := F.data[key]; exists {
		F.data[key] = value
		return nil
	}

	if len(F.data) >= F.capacity {
		oldestKey := F.keys[0]
		delete(F.data, oldestKey)
		F.keys = F.keys[1:]

		F.emit(cache.Event[K, V]{
			Type:  cache.EventTypeEviction,
			Key:   oldestKey,
			Value: value,
		})
	}

	F.data[key] = value
	F.keys = append(F.keys, key)
	return nil
}

// Delete removes a key-value pair. Returns error if key not found
func (F *FIFOCache[K, V]) Delete(key K) error {
	if _, exists := F.data[key]; !exists {
		return common.ErrKeyNotFound
	}

	delete(F.data, key)

	for i, k := range F.keys {
		if k == key {
			F.keys = append(F.keys[:i], F.keys[i+1:]...)
			break
		}
	}

	return nil
}

// Clear removes all key-value pairs
func (F *FIFOCache[K, V]) Clear() {
	F.data = make(map[K]V, F.capacity)
	F.keys = make([]K, 0)
}

func (L *FIFOCache[K, V]) OnEvent(callback func(event cache.Event[K, V])) {
	L.eventCallbacks = append(L.eventCallbacks, callback)
}

func (L *FIFOCache[K, V]) emit(event cache.Event[K, V]) {
	for _, callback := range L.eventCallbacks {
		callback(event)
	}
}

func (F *FIFOCache[K, V]) Range(f func(K, V) bool) {
	// TODO: implement
}
