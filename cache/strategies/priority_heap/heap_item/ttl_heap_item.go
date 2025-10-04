package heap_item

import "time"

type TTLHeapItem[K comparable, V any] struct {
	Key       K
	Value     V
	ExpiresAt time.Time
	index     int
}

func NewTTLHeapItem[K comparable, V any](key K, value V, ttl time.Duration) *TTLHeapItem[K, V] {
	return &TTLHeapItem[K, V]{
		Key:       key,
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
		index:     -1,
	}
}

func (i *TTLHeapItem[K, V]) GetPriority() int {
	return int(i.ExpiresAt.UnixNano())
}

func (i *TTLHeapItem[K, V]) SetPriority(p int) {
	i.ExpiresAt = time.Unix(0, int64(p))
}

func (i *TTLHeapItem[K, V]) GetIndex() int {
	return i.index
}

func (i *TTLHeapItem[K, V]) SetIndex(index int) {
	i.index = index
}
