package cache

type IterableCache[K comparable, V any] interface {
	Cache[K, V]
	Range(func(k K, v V) bool)
}
