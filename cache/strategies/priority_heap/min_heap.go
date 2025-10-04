package priority_heap

import (
	common "caching-labwork/cache/strategies/priority_heap/heap_item"
	"container/heap"
)

type MinHeap[K comparable, V any] struct {
	items []common.Item[K, V]
}

func NewMinHeap[K comparable, V any]() *MinHeap[K, V] {
	return &MinHeap[K, V]{
		items: []common.Item[K, V]{},
	}
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
	item := x.(common.Item[K, V])
	item.SetIndex(len(h.items))
	h.items = append(h.items, item)
}

func (h *MinHeap[K, V]) Pop() any {
	n := len(h.items)
	item := h.items[n-1]
	item.SetIndex(-1)
	h.items = h.items[:n-1]
	return item
}

func (h *MinHeap[K, V]) Update(item common.Item[K, V], newPriority int) {
	item.SetPriority(newPriority)
	heap.Fix(h, item.GetIndex())
}

func (h *MinHeap[K, V]) Peek() common.Item[K, V] {
	if len(h.items) == 0 {
		return nil
	}
	return h.items[0]
}
