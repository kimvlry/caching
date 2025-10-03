package heap_item

type MinHeapItem[K comparable, V any] struct {
	Key       K
	Value     V
	Frequency int
	index     int
}

func NewPriorityHeapItem[K comparable, V any](key K, value V, freq int) *MinHeapItem[K, V] {
	return &MinHeapItem[K, V]{
		Key:       key,
		Value:     value,
		Frequency: freq,
		index:     -1,
	}
}

func (i *MinHeapItem[K, V]) GetPriority() int {
	return i.Frequency
}

func (i *MinHeapItem[K, V]) SetPriority(p int) {
	i.Frequency = p
}

func (i *MinHeapItem[K, V]) GetIndex() int {
	return i.index
}

func (i *MinHeapItem[K, V]) SetIndex(index int) {
	i.index = index
}
