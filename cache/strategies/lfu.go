package strategies

import (
	"container/heap"
	"github.com/kimvlry/caching/cache"
	"github.com/kimvlry/caching/cache/strategies/common"
	"github.com/kimvlry/caching/cache/strategies/priority_heap"
	"github.com/kimvlry/caching/cache/strategies/priority_heap/heap_item"
)

type LFUCache[K comparable, V any] struct {
	capacity int
	data     map[K]*heap_item.MinHeapItem[K, V]
	keys     *priority_heap.MinHeap[K, V]

	eventCallbacks []func(cache.Event[K, V])
}

func NewLFUCache[K comparable, V any](capacity int) *LFUCache[K, V] {
	return &LFUCache[K, V]{
		capacity: capacity,
		data:     make(map[K]*heap_item.MinHeapItem[K, V], capacity),
		keys:     priority_heap.NewMinHeap[K, V](),
	}
}

func (L *LFUCache[K, V]) Get(key K) (V, error) {
	item, exists := L.data[key]
	if !exists {
		var zero V
		return zero, common.ErrKeyNotFound
	}
	L.keys.Update(item, item.GetPriority()+1)
	return item.Value, nil
}

func (L *LFUCache[K, V]) Set(key K, value V) error {
	if item, exists := L.data[key]; exists {
		item.Value = value
		L.keys.Update(item, item.GetPriority()+1)
		return nil
	}

	if len(L.data) >= L.capacity {
		evicted := heap.Pop(L.keys).(*heap_item.MinHeapItem[K, V])
		delete(L.data, evicted.Key)

		L.emit(cache.Event[K, V]{
			Type:  cache.EventTypeEviction,
			Key:   evicted.Key,
			Value: evicted.Value,
		})
	}

	item := heap_item.NewPriorityHeapItem(key, value, 1)
	L.data[key] = item
	heap.Push(L.keys, item)
	return nil
}

func (L *LFUCache[K, V]) Delete(key K) error {
	item, exists := L.data[key]
	if !exists {
		return common.ErrKeyNotFound
	}
	heap.Remove(L.keys, item.GetIndex())
	delete(L.data, key)
	return nil
}

func (L *LFUCache[K, V]) Clear() {
	L.data = make(map[K]*heap_item.MinHeapItem[K, V])
	L.keys = priority_heap.NewMinHeap[K, V]()
}

func (L *LFUCache[K, V]) OnEvent(callback func(event cache.Event[K, V])) {
	L.eventCallbacks = append(L.eventCallbacks, callback)
}

func (L *LFUCache[K, V]) emit(event cache.Event[K, V]) {
	for _, callback := range L.eventCallbacks {
		callback(event)
	}
}

func (L *LFUCache[K, V]) Range(fn func(K, V) bool) {
	for k, v := range L.data {
		if !fn(k, v.Value) {
			break
		}
	}
}
