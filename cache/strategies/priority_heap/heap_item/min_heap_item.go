package heap_item

type minHeapItem[K comparable, V any] struct {
	Key       K
	Value     V
	Frequency int64
	index     int
}

func NewPriorityHeapItem[K comparable, V any](key K, value V, freq int64) Item[K, V] {
	return &minHeapItem[K, V]{
		Key:       key,
		Value:     value,
		Frequency: freq,
		index:     -1,
	}
}

func (i *minHeapItem[K, V]) GetPriority() int64 {
	return i.Frequency
}

func (i *minHeapItem[K, V]) SetPriority(p int64) {
	i.Frequency = p
}

func (i *minHeapItem[K, V]) GetIndex() int {
	return i.index
}

func (i *minHeapItem[K, V]) SetIndex(index int) {
	i.index = index
}

func (i *minHeapItem[K, V]) GetKey() K {
	return i.Key
}

func (i *minHeapItem[K, V]) SetKey(k K) {
	i.Key = k
}

func (i *minHeapItem[K, V]) GetValue() V {
	return i.Value
}

func (i *minHeapItem[K, V]) SetValue(v V) {
	i.Value = v
}
