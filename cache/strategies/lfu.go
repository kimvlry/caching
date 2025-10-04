package strategies

import (
	"caching-labwork/cache/strategies/common"
	"caching-labwork/cache/strategies/priority_heap"
	"caching-labwork/cache/strategies/priority_heap/heap_item"
	"container/heap"
)

type LFUCache[K comparable, V any] struct {
	capacity int
	data     map[K]*heap_item.MinHeapItem[K, V]
	keys     *priority_heap.MinHeap[K, V]
}

func NewLFUCache[K comparable, V any](capacity int) *LFUCache[K, V] {
	return &LFUCache[K, V]{
		capacity: capacity,
		data:     make(map[K]*heap_item.MinHeapItem[K, V], capacity),
		keys:     priority_heap.NewMinHeap[K, V](),
	}
}

func (c *LFUCache[K, V]) Get(key K) (V, error) {
	item, exists := c.data[key]
	if !exists {
		var zero V
		return zero, common.ErrKeyNotFound
	}
	c.keys.Update(item, item.GetPriority()+1)
	return item.Value, nil
}

func (c *LFUCache[K, V]) Set(key K, value V) error {
	if item, exists := c.data[key]; exists {
		item.Value = value
		c.keys.Update(item, item.GetPriority()+1)
		return nil
	}

	if len(c.data) >= c.capacity {
		evicted := heap.Pop(c.keys).(*heap_item.MinHeapItem[K, V])
		delete(c.data, evicted.Key)
	}

	item := heap_item.NewPriorityHeapItem(key, value, 1)
	c.data[key] = item
	heap.Push(c.keys, item)
	return nil
}

func (c *LFUCache[K, V]) Delete(key K) error {
	item, exists := c.data[key]
	if !exists {
		return common.ErrKeyNotFound
	}
	heap.Remove(c.keys, item.GetIndex())
	delete(c.data, key)
	return nil
}

func (c *LFUCache[K, V]) Clear() {
	c.data = make(map[K]*heap_item.MinHeapItem[K, V])
	c.keys = priority_heap.NewMinHeap[K, V]()
}
