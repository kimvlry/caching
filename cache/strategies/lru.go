package strategies

import (
	"caching-labwork/cache"
	"caching-labwork/cache/strategies/common"
)

// LRUCache implements a Least Recently Used cache
type LRUCache[K comparable, V any] struct {
	capacity int
	data     map[K]V
	keys     []K

	eventCallbacks []func(cache.Event[K, V])
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	return &LRUCache[K, V]{
		capacity: capacity,
		data:     make(map[K]V, capacity),
		keys:     make([]K, capacity),
	}
}

func (L *LRUCache[K, V]) Get(key K) (V, error) {
	if value, exists := L.data[key]; exists {
		L.moveKeyToEnd(key)
		return value, nil
	}

	var zero V
	return zero, common.ErrKeyNotFound
}

func (L *LRUCache[K, V]) Set(key K, value V) error {
	if _, exists := L.data[key]; exists {
		L.moveKeyToEnd(key)
		L.data[key] = value
		return nil
	}

	if len(L.data) >= L.capacity {
		leastRecentKey := L.keys[0]
		delete(L.data, leastRecentKey)
		L.keys = L.keys[1:]

		L.emit(cache.Event[K, V]{
			Type:  cache.EventTypeEviction,
			Key:   leastRecentKey,
			Value: value,
		})
	}

	L.keys = append(L.keys, key)
	L.data[key] = value
	return nil
}

func (L *LRUCache[K, V]) Delete(key K) error {
	if _, exists := L.data[key]; !exists {
		return common.ErrKeyNotFound
	}

	delete(L.data, key)

	for i, k := range L.keys {
		if k == key {
			L.keys = append(L.keys[:i], L.keys[i+1:]...)
			break
		}
	}
	return nil
}

func (L *LRUCache[K, V]) Clear() {
	L.data = make(map[K]V, L.capacity)
	L.keys = make([]K, 0)
}

func (L *LRUCache[K, V]) Range(f func(K, V) bool) {
	// TODO: implement
}

func (L *LRUCache[K, V]) OnEvent(callback func(event cache.Event[K, V])) {
	L.eventCallbacks = append(L.eventCallbacks, callback)
}

func (L *LRUCache[K, V]) emit(event cache.Event[K, V]) {
	for _, callback := range L.eventCallbacks {
		callback(event)
	}
}

func (L *LRUCache[K, V]) moveKeyToEnd(key K) {
	for i, k := range L.keys {
		if k == key {
			L.keys = append(L.keys[:i], L.keys[i+1:]...)
			break
		}
	}
	L.keys = append(L.keys, key)
}
