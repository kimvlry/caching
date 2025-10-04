package strategies

// ARCCache implements an Adaptive Replacement Cache
type ARCCache[K comparable, V any] struct {
	// TODO: Add necessary fields for ARC implementation
}

func (A ARCCache[K, V]) Get(key K) (V, error) {
	//TODO implement me
	panic("implement me")
}

func (A ARCCache[K, V]) Set(key K, value V) error {
	//TODO implement me
	panic("implement me")
}

func (A ARCCache[K, V]) Delete(key K) error {
	//TODO implement me
	panic("implement me")
}

func (A ARCCache[K, V]) Clear() {
	//TODO implement me
	panic("implement me")
}

func NewARCCache[K comparable, V any](capacity int) *ARCCache[K, V] {
	return nil
}
