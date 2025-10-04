package cache

// Cache defines the interface for all cache implementations
type Cache[K comparable, V any] interface {
	Get(key K) (V, error)
	Set(key K, value V) error
	Delete(key K) error
	Clear()
}

// Individual cache implementations are in separate files:
// - fifo.go: FIFO cache implementation
// - lru.go: LRU cache implementation
// - lfu.go: LFU cache implementation (to be implemented)
// - ttl.go: TTL cache implementation (to be implemented)
// - arc.go: ARC cache implementation (to be implemented)
