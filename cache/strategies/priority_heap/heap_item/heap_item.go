package heap_item

// Item - interface defining priority_heap pq_item
type Item[K comparable, V any] interface {
	GetPriority() int64
	SetPriority(int64)
	GetIndex() int
	SetIndex(index int)
	GetKey() K
	SetKey(k K)
	GetValue() V
	SetValue(V)
}
