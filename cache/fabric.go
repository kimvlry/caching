package cache

import (
	"caching-labwork/cache/strategies"
	"time"
)

// NewFIFOCache creates a new FIFO (First In, First Out) cache
func NewFIFOCache[K comparable, V any](capacity int) Cache[K, V] {
	return strategies.NewFIFOCache[K, V](capacity)
}

// NewLRUCache creates a new LRU (Least Recently Used) cache
func NewLRUCache[K comparable, V any](capacity int) Cache[K, V] {
	return strategies.NewLRUCache[K, V](capacity)
}

// NewLFUCache creates a new LFU (Least Frequently Used) cache
func NewLFUCache[K comparable, V any](capacity int) Cache[K, V] {
	return strategies.NewLFUCache[K, V](capacity)
}

// NewTTLCache creates a new TTL (Time To Live) cache
func NewTTLCache[K comparable, V any](capacity int, ttl time.Duration) Cache[K, V] {
	return strategies.NewTTLCache[K, V](capacity, ttl)
}

// NewARCCache creates a new ARC (Adaptive Replacement Cache)
// TODO: Implement this function (Advanced task)
func NewARCCache[K comparable, V any](capacity int) Cache[K, V] {
	// Students should implement this
	return &emptyCache[K, V]{}
}
