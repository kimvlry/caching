package cache

// emptyCache is a non-functional implementation for testing
type emptyCache[K comparable, V any] struct{}

func (e *emptyCache[K, V]) Get(key K) (V, error) {
	var zero V
	return zero, ErrKeyNotFound
}

func (e *emptyCache[K, V]) Set(key K, value V) error {
	return ErrCacheFull
}

func (e *emptyCache[K, V]) Delete(key K) error {
	return ErrKeyNotFound
}

func (e *emptyCache[K, V]) Clear() {
	// Do nothing
}
