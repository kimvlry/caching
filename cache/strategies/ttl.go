package strategies

import (
	"caching-labwork/cache"
	"caching-labwork/cache/strategies/common"
	"caching-labwork/cache/strategies/priority_heap"
	"caching-labwork/cache/strategies/priority_heap/heap_item"
	"container/heap"
	"sync"
	"time"
)

// TODO: set individual ttl for each item

type TTLCache[K comparable, V any] struct {
	capacity int
	ttl      time.Duration
	data     map[K]*heap_item.TTLHeapItem[K, V]
	keys     *priority_heap.MinHeap[K, V]
	mutex    sync.Mutex

	eventCallbacks []func(cache.Event[K, V])
}

func NewTTLCache[K comparable, V any](capacity int, ttl time.Duration) *TTLCache[K, V] {
	return &TTLCache[K, V]{
		capacity: capacity,
		ttl:      ttl,
		data:     make(map[K]*heap_item.TTLHeapItem[K, V], capacity),
		keys:     priority_heap.NewMinHeap[K, V](),
	}
}

func (T *TTLCache[K, V]) Set(key K, value V) error {
	T.mutex.Lock()
	defer T.mutex.Unlock()

	if item, exists := T.data[key]; exists {
		item.Value = value
		item.SetPriority(int(time.Now().Add(T.ttl).UnixNano()))
		heap.Fix(T.keys, item.GetIndex())
		return nil
	}

	if len(T.data) >= T.capacity {
		evicted := heap.Pop(T.keys).(*heap_item.TTLHeapItem[K, V])
		delete(T.data, evicted.Key)

		T.emit(cache.Event[K, V]{
			Type:  cache.EventTypeEviction,
			Key:   evicted.Key,
			Value: evicted.Value,
		})
	}

	item := heap_item.NewTTLHeapItem(key, value, T.ttl)
	T.data[key] = item
	heap.Push(T.keys, item)
	return nil
}

func (T *TTLCache[K, V]) Get(key K) (V, error) {
	T.mutex.Lock()
	defer T.mutex.Unlock()

	item, exists := T.data[key]
	if !exists {
		var zero V
		return zero, common.ErrKeyNotFound
	}

	if time.Now().After(item.ExpiresAt) {
		heap.Remove(T.keys, item.GetIndex())
		delete(T.data, key)
		var zero V
		return zero, common.ErrKeyNotFound
	}

	return item.Value, nil
}

func (T *TTLCache[K, V]) Delete(key K) error {
	T.mutex.Lock()
	defer T.mutex.Unlock()
	item, exists := T.data[key]
	if !exists {
		return common.ErrKeyNotFound
	}
	heap.Remove(T.keys, item.GetIndex())
	delete(T.data, key)
	return nil
}

func (T *TTLCache[K, V]) Clear() {
	T.mutex.Lock()
	defer T.mutex.Unlock()
	T.data = make(map[K]*heap_item.TTLHeapItem[K, V])
	T.keys = priority_heap.NewMinHeap[K, V]()
}

func (L *TTLCache[K, V]) OnEvent(callback func(event cache.Event[K, V])) {
	L.eventCallbacks = append(L.eventCallbacks, callback)
}

func (L *TTLCache[K, V]) emit(event cache.Event[K, V]) {
	for _, callback := range L.eventCallbacks {
		callback(event)
	}
}

func (T *TTLCache[K, V]) Range(f func(K, V) bool) {
	for k, v := range T.data {
		if !f(k, v.Value) {
			break
		}
	}
}

func (T *TTLCache[K, V]) StartEvictor(interval time.Duration) {
	go func() {
		for {
			time.Sleep(interval)
			T.mutex.Lock()
			for {
				top := T.keys.Peek()
				if top == nil {
					break
				}
				item := top.(*heap_item.TTLHeapItem[K, V])
				if time.Now().Before(item.ExpiresAt) {
					break
				}
				heap.Pop(T.keys)
				delete(T.data, item.Key)

				T.emit(cache.Event[K, V]{
					Type:  cache.EventTypeEviction,
					Key:   item.Key,
					Value: item.Value,
				})
			}
			T.mutex.Unlock()
		}
	}()
}
