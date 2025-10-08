package strategies

import (
	"container/list"
	"github.com/kimvlry/caching/cache"
	"github.com/kimvlry/caching/cache/strategies/common"
)

// lruCache implements a Least Recently Used cache
type lruCache[K comparable, V any] struct {
	capacity int
	data     map[K]*list.Element
	keys     *list.List

	eventCallbacks []func(cache.Event[K, V])
}

func newLruCache[K comparable, V any](capacity int) cache.IterableCache[K, V] {
	return &lruCache[K, V]{
		capacity: capacity,
		data:     make(map[K]*list.Element, capacity),
		keys:     list.New(),
	}
}

func (l *lruCache[K, V]) Get(key K) (V, error) {
	if element, exists := l.data[key]; exists {
		l.keys.MoveToBack(element)
		return element.Value.(*entry[K, V]).value, nil
	}

	var zero V
	return zero, common.ErrKeyNotFound
}

func (l *lruCache[K, V]) Set(key K, value V) error {
	if elem, exists := l.data[key]; exists {
		elem.Value.(*entry[K, V]).value = value
		l.keys.MoveToBack(elem)
		return nil
	}

	if len(l.data) >= l.capacity {
		oldest := l.keys.Front()
		if oldest != nil {
			evicted := oldest.Value.(*entry[K, V])
			delete(l.data, evicted.key)
			l.keys.Remove(oldest)
			l.emit(cache.Event[K, V]{
				Type:  cache.EventTypeEviction,
				Key:   evicted.key,
				Value: evicted.value,
			})
		}
	}

	e := &entry[K, V]{key: key, value: value}
	elem := l.keys.PushBack(e)
	l.data[key] = elem
	return nil
}

func (l *lruCache[K, V]) Delete(key K) error {
	if elem, exists := l.data[key]; exists {
		l.keys.Remove(elem)
		delete(l.data, key)
		return nil
	}
	return common.ErrKeyNotFound
}

func (l *lruCache[K, V]) Clear() {
	l.data = make(map[K]*list.Element, l.capacity)
	l.keys = list.New()
}

func (l *lruCache[K, V]) Range(fn func(K, V) bool) {
	for elem := l.keys.Front(); elem != nil; elem = elem.Next() {
		e := elem.Value.(*entry[K, V])
		if !fn(e.key, e.value) {
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
