package strategies

import (
	"container/heap"
	"github.com/kimvlry/caching/cache"
	"github.com/kimvlry/caching/cache/strategies/common"
	"github.com/kimvlry/caching/cache/strategies/priority_heap"
	"github.com/kimvlry/caching/cache/strategies/priority_heap/heap_item"
)

type lfuCache[K comparable, V any] struct {
	capacity int
	data     map[K]heap_item.Item[K, V]
	keys     *priority_heap.MinHeap[K, V]

	eventCallbacks []func(cache.Event[K, V])
}

func NewLFUCache[K comparable, V any](capacity int) cache.IterableCache[K, V] {
	return &lfuCache[K, V]{
		capacity: capacity,
		data:     make(map[K]heap_item.Item[K, V], capacity),
		keys:     priority_heap.NewMinHeap[K, V](),
	}
}

func (l *lfuCache[K, V]) Get(key K) (V, error) {
	item, exists := l.data[key]
	if !exists {
		var zero V
		return zero, common.ErrKeyNotFound
	}

	item.SetPriority(item.GetPriority() + 1)
	heap.Fix(l.keys, item.GetIndex())

	return item.GetValue(), nil
}

func (l *lfuCache[K, V]) Set(key K, value V) error {
	if item, exists := l.data[key]; exists {
		item.SetPriority(item.GetPriority() + 1)
		item.SetValue(value)
		heap.Fix(l.keys, item.GetIndex())
		return nil
	}

	if len(l.data) >= l.capacity {
		evicted := heap.Pop(l.keys).(heap_item.Item[K, V])
		delete(l.data, evicted.GetKey())

		l.emit(cache.Event[K, V]{
			Type:  cache.EventTypeEviction,
			Key:   evicted.GetKey(),
			Value: evicted.GetValue(),
		})
	}

	item := heap_item.NewPriorityHeapItem(key, value, 1)
	l.data[key] = item
	heap.Push(l.keys, item)
	return nil
}

func (l *lfuCache[K, V]) Delete(key K) error {
	item, exists := l.data[key]
	if !exists {
		return common.ErrKeyNotFound
	}
	heap.Remove(l.keys, item.GetIndex())
	delete(l.data, key)
	return nil
}

func (l *lfuCache[K, V]) Clear() {
	l.data = make(map[K]heap_item.Item[K, V])
	l.keys = priority_heap.NewMinHeap[K, V]()
}

func (l *lfuCache[K, V]) OnEvent(callback func(event cache.Event[K, V])) {
	l.eventCallbacks = append(l.eventCallbacks, callback)
}

func (l *lfuCache[K, V]) emit(event cache.Event[K, V]) {
	for _, callback := range l.eventCallbacks {
		callback(event)
	}
}

func (l *lfuCache[K, V]) Range(fn func(K, V) bool) {
	for k, item := range l.data {
		if !fn(k, item.GetValue()) {
			break
		}
	}
}
