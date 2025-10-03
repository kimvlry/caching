package strategies

import "caching-labwork/cache/common"

// LRUCache implements a Least Recently Used cache
type LRUCache[K comparable, V any] struct {
	capacity int
	data     map[K]V
	keys     []K
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	return &LRUCache[K, V]{
		capacity: capacity,
		data:     make(map[K]V, capacity),
		keys:     make([]K, capacity),
	}
}

func (l *LRUCache[K, V]) Get(key K) (V, error) {
	if value, exists := l.data[key]; exists {
		l.moveKeyToEnd(key)
		return value, nil
	}
	var zero V
	return zero, common.ErrKeyNotFound
}

func (l *LRUCache[K, V]) Set(key K, value V) error {
	if _, exists := l.data[key]; exists {
		l.moveKeyToEnd(key)
		l.data[key] = value
		return nil
	}

	if len(l.data) >= l.capacity {
		leastRecentKey := l.keys[0]
		delete(l.data, leastRecentKey)
		l.keys = l.keys[1:]
	}

	l.keys = append(l.keys, key)
	l.data[key] = value
	return nil
}

func (l *LRUCache[K, V]) Delete(key K) error {
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

func (l *LRUCache[K, V]) Clear() {
	l.data = make(map[K]V, l.capacity)
	l.keys = make([]K, 0)
}

func (l *LRUCache[K, V]) moveKeyToEnd(key K) {
	for i, k := range l.keys {
		if k == key {
			l.keys = append(l.keys[:i], l.keys[i+1:]...)
			break
		}
	}
	l.keys = append(l.keys, key)
}
