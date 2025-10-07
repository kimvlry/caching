package heap_item

import "time"

type ttlHeapItem[K comparable, V any] struct {
	Key       K
	Value     V
	ExpiresAt time.Time
	index     int
}

func NewTTLHeapItem[K comparable, V any](key K, value V, ttl time.Duration) Item[K, V] {
	expiresAt := time.Now().Add(ttl)
	return &ttlHeapItem[K, V]{
		Key:       key,
		Value:     value,
		ExpiresAt: expiresAt,
		index:     -1,
	}
}

func (i *ttlHeapItem[K, V]) GetPriority() int64 {
	return i.ExpiresAt.UnixNano()
}

func (i *ttlHeapItem[K, V]) SetPriority(p int64) {
	i.ExpiresAt = time.Unix(0, p)
}

func (i *ttlHeapItem[K, V]) GetIndex() int {
	return i.index
}

func (i *ttlHeapItem[K, V]) SetIndex(idx int) {
	i.index = idx
}

func (i *ttlHeapItem[K, V]) GetKey() K {
	return i.Key
}

func (i *ttlHeapItem[K, V]) SetKey(k K) {
	i.Key = k
}

func (i *ttlHeapItem[K, V]) GetValue() V {
	return i.Value
}

func (i *ttlHeapItem[K, V]) SetValue(v V) {
	i.Value = v
}
