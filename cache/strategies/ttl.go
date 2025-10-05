package strategies

import (
	"container/heap"
	"github.com/kimvlry/caching/cache"
	"github.com/kimvlry/caching/cache/strategies/common"
	"github.com/kimvlry/caching/cache/strategies/priority_heap"
	"github.com/kimvlry/caching/cache/strategies/priority_heap/heap_item"
	"sync"
	"time"
)

// TODO: set individual ttl for each item

type ttlCache[K comparable, V any] struct {
	capacity int
	ttl      time.Duration
	data     map[K]*heap_item.TTLHeapItem[K, V]
	keys     *priority_heap.MinHeap[K, V]
	mutex    sync.Mutex

	eventCallbacks []func(cache.Event[K, V])
}

func NewTTLCache[K comparable, V any](capacity int, ttl time.Duration) cache.Cache[K, V] {
	return &ttlCache[K, V]{
		capacity: capacity,
		ttl:      ttl,
		data:     make(map[K]*heap_item.TTLHeapItem[K, V], capacity),
		keys:     priority_heap.NewMinHeap[K, V](),
	}
}

func (t *ttlCache[K, V]) Set(key K, value V) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if item, exists := t.data[key]; exists {
		item.Value = value
		item.SetPriority(int(time.Now().Add(t.ttl).UnixNano()))
		heap.Fix(t.keys, item.GetIndex())
		return nil
	}

	if len(t.data) >= t.capacity {
		evicted := heap.Pop(t.keys).(*heap_item.TTLHeapItem[K, V])
		delete(t.data, evicted.Key)

		t.emit(cache.Event[K, V]{
			Type:  cache.EventTypeEviction,
			Key:   evicted.Key,
			Value: evicted.Value,
		})
	}

	item := heap_item.NewTTLHeapItem(key, value, t.ttl)
	t.data[key] = item
	heap.Push(t.keys, item)
	return nil
}

func (t *ttlCache[K, V]) Get(key K) (V, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	item, exists := t.data[key]
	if !exists {
		var zero V
		return zero, common.ErrKeyNotFound
	}

	if time.Now().After(item.ExpiresAt) {
		heap.Remove(t.keys, item.GetIndex())
		delete(t.data, key)
		var zero V
		return zero, common.ErrKeyNotFound
	}

	return item.Value, nil
}

func (t *ttlCache[K, V]) Delete(key K) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	item, exists := t.data[key]
	if !exists {
		return common.ErrKeyNotFound
	}
	heap.Remove(t.keys, item.GetIndex())
	delete(t.data, key)
	return nil
}

func (t *ttlCache[K, V]) Clear() {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.data = make(map[K]*heap_item.TTLHeapItem[K, V])
	t.keys = priority_heap.NewMinHeap[K, V]()
}

func (t *ttlCache[K, V]) OnEvent(callback func(event cache.Event[K, V])) {
	t.eventCallbacks = append(t.eventCallbacks, callback)
}

func (t *ttlCache[K, V]) emit(event cache.Event[K, V]) {
	for _, callback := range t.eventCallbacks {
		callback(event)
	}
}

func (t *ttlCache[K, V]) Range(f func(K, V) bool) {
	for k, v := range t.data {
		if !f(k, v.Value) {
			break
		}
	}
}

func (t *ttlCache[K, V]) StartEvictor(interval time.Duration) {
	go func() {
		for {
			time.Sleep(interval)
			t.mutex.Lock()
			for {
				top := t.keys.Peek()
				if top == nil {
					break
				}
				item := top.(*heap_item.TTLHeapItem[K, V])
				if time.Now().Before(item.ExpiresAt) {
					break
				}
				heap.Pop(t.keys)
				delete(t.data, item.Key)

				t.emit(cache.Event[K, V]{
					Type:  cache.EventTypeEviction,
					Key:   item.Key,
					Value: item.Value,
				})
			}
			t.mutex.Unlock()
		}
	}()
}
