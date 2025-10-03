package strategies

import (
	"caching-labwork/cache/common"
)

// FIFOCache implements a First In, First Out cache
type FIFOCache[K comparable, V any] struct {
	capacity int
	data     map[K]V
	keys     []K
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
func (f *FIFOCache[K, V]) Get(key K) (V, error) {
	if value, exists := f.data[key]; exists {
		return value, nil
	}
	var zero V
	return zero, common.ErrKeyNotFound
}

// Set adds or updates a key-value pair. If cache is full, the oldest pq_item gets evicted (first in)
func (f *FIFOCache[K, V]) Set(key K, value V) error {
	if _, exists := f.data[key]; exists {
		f.data[key] = value
		return nil
	}

	if len(f.data) >= f.capacity {
		oldestKey := f.keys[0]
		delete(f.data, oldestKey)
		f.keys = f.keys[1:]
	}

	f.data[key] = value
	f.keys = append(f.keys, key)
	return nil
}

// Delete removes a key-value pair. Returns error if key not found
func (f *FIFOCache[K, V]) Delete(key K) error {
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
func (f *FIFOCache[K, V]) Clear() {
	f.data = make(map[K]V, f.capacity)
	f.keys = make([]K, 0)
}
