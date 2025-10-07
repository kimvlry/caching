package priority_heap

import (
	"container/heap"
	"github.com/kimvlry/caching/cache/strategies/priority_heap/heap_item"
)

type MinHeap[K comparable, V any] struct {
	items []heap_item.Item[K, V]
}

func NewMinHeap[K comparable, V any]() *MinHeap[K, V] {
	h := &MinHeap[K, V]{items: make([]heap_item.Item[K, V], 0)}
	heap.Init(h)
	return h
}

func (h *MinHeap[K, V]) Len() int {
	return len(h.items)
}

func (h *MinHeap[K, V]) Less(i, j int) bool {
	return h.items[i].GetPriority() < h.items[j].GetPriority()
}

func (h *MinHeap[K, V]) Swap(i, j int) {
	h.items[i], h.items[j] = h.items[j], h.items[i]
	h.items[i].SetIndex(i)
	h.items[j].SetIndex(j)
}

func (h *MinHeap[K, V]) Push(x any) {
	item := x.(heap_item.Item[K, V])
	item.SetIndex(len(h.items))
	h.items = append(h.items, item)
}

func (h *MinHeap[K, V]) Pop() any {
	n := len(h.items)
	if n == 0 {
		return nil
	}
	item := h.items[n-1]
	item.SetIndex(-1)
	h.items = h.items[:n-1]
	return item
}

func (h *MinHeap[K, V]) Peek() heap_item.Item[K, V] {
	if len(h.items) == 0 {
		return nil
	}
	return h.items[0]
}
