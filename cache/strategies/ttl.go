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

type TTLCache[K comparable, V any] interface {
	cache.Cache[K, V]
	cache.IterableCache[K, V]
	SetWithTTL(K, V, time.Duration) error
	GetDefaultTTL() time.Duration
}

type ttlCache[K comparable, V any] struct {
	capacity   int
	defaultTTL time.Duration
	data       map[K]heap_item.Item[K, V]
	keys       *priority_heap.MinHeap[K, V]
	mutex      sync.Mutex

	eventCallbacks []func(cache.Event[K, V])
	stopEvictor    chan struct{}
	evictorOnce    sync.Once
}

func newTtlCache[K comparable, V any](capacity int, defaultTTL time.Duration) cache.IterableCache[K, V] {
	c := &ttlCache[K, V]{
		capacity:   capacity,
		defaultTTL: defaultTTL,
		data:       make(map[K]heap_item.Item[K, V], capacity),
		keys:       priority_heap.NewMinHeap[K, V](),
	}
	c.startEvictor(defaultTTL / 2)
	return c
}

func (t *ttlCache[K, V]) GetDefaultTTL() time.Duration {
	return t.defaultTTL
}

func (t *ttlCache[K, V]) SetWithTTL(key K, value V, ttl time.Duration) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if item, exists := t.data[key]; exists {
		newExpiresAt := time.Now().Add(ttl)
		item.SetPriority(newExpiresAt.UnixNano())
		item.SetValue(value)
		heap.Fix(t.keys, item.GetIndex())
		return nil
	}

	if len(t.data) >= t.capacity {
		popped := heap.Pop(t.keys)
		if popped != nil {
			evicted := popped.(heap_item.Item[K, V])
			delete(t.data, evicted.GetKey())
			t.emit(cache.Event[K, V]{
				Type:  cache.EventTypeEviction,
				Key:   evicted.GetKey(),
				Value: evicted.GetValue(),
			})
		}
	}

	item := heap_item.NewTTLHeapItem(key, value, ttl)
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

	expiresAt := time.Unix(0, item.GetPriority())
	if time.Now().After(expiresAt) {
		if item.GetIndex() >= 0 {
			heap.Remove(t.keys, item.GetIndex())
		}
		delete(t.data, key)
		t.emit(cache.Event[K, V]{
			Type:  cache.EventTypeEviction,
			Key:   item.GetKey(),
			Value: item.GetValue(),
		})
		var zero V
		return zero, common.ErrKeyNotFound
	}

	return item.GetValue(), nil
}

func (t *ttlCache[K, V]) Delete(key K) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	item, exists := t.data[key]
	if !exists {
		return common.ErrKeyNotFound
	}

	if item.GetIndex() >= 0 {
		heap.Remove(t.keys, item.GetIndex())
	}
	delete(t.data, key)
	return nil
}

func (t *ttlCache[K, V]) Clear() {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.data = make(map[K]heap_item.Item[K, V])
	t.keys = priority_heap.NewMinHeap[K, V]()
}

func (t *ttlCache[K, V]) OnEvent(callback func(event cache.Event[K, V])) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.eventCallbacks = append(t.eventCallbacks, callback)
}

func (t *ttlCache[K, V]) emit(event cache.Event[K, V]) {
	for _, callback := range t.eventCallbacks {
		callback(event)
	}
}

func (t *ttlCache[K, V]) Range(f func(K, V) bool) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	now := time.Now()
	for k, item := range t.data {
		expiresAt := time.Unix(0, item.GetPriority())
		if now.After(expiresAt) {
			continue
		}
		if !f(k, item.GetValue()) {
			break
		}
	}
}

func (t *ttlCache[K, V]) startEvictor(interval time.Duration) {
	t.evictorOnce.Do(func() {
		t.stopEvictor = make(chan struct{})
		go func() {
			ticker := time.NewTicker(interval)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					t.evictExpired()
				case <-t.stopEvictor:
					return
				}
			}
		}()
	})
}

func (t *ttlCache[K, V]) evictExpired() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	now := time.Now()
	for {
		top := t.keys.Peek()
		if top == nil {
			break
		}

		expiresAt := time.Unix(0, top.GetPriority())
		if now.Before(expiresAt) {
			break
		}

		item := heap.Pop(t.keys).(heap_item.Item[K, V])
		delete(t.data, item.GetKey())
		t.emit(cache.Event[K, V]{
			Type:  cache.EventTypeEviction,
			Key:   item.GetKey(),
			Value: item.GetValue(),
		})
	}
}

func (t *ttlCache[K, V]) Stop() {
	if t.stopEvictor != nil {
		close(t.stopEvictor)
		t.stopEvictor = nil
	}
}

func (t *ttlCache[K, V]) Set(key K, value V) error {
	return t.SetWithTTL(key, value, t.defaultTTL)
}
