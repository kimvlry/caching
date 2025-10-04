package heap_item

// Item - interface defining priority_heap pq_item
type Item[K comparable, V any] interface {
	GetPriority() int
	SetPriority(int)
	GetIndex() int
	SetIndex(int)
}
