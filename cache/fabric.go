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
// TODO: Implement this function
func NewLRUCache[K comparable, V any](capacity int) Cache[K, V] {
	// Students should implement this
	return &emptyCache[K, V]{}
}

// NewLFUCache creates a new LFU (Least Frequently Used) cache
// TODO: Implement this function
func NewLFUCache[K comparable, V any](capacity int) Cache[K, V] {
	// Students should implement this
	return &emptyCache[K, V]{}
}

// NewTTLCache creates a new TTL (Time To Live) cache
// TODO: Implement this function
func NewTTLCache[K comparable, V any](capacity int, ttl time.Duration) Cache[K, V] {
	// Students should implement this
	return &emptyCache[K, V]{}
}

// NewARCCache creates a new ARC (Adaptive Replacement Cache)
// TODO: Implement this function (Advanced task)
func NewARCCache[K comparable, V any](capacity int) Cache[K, V] {
	// Students should implement this
	return &emptyCache[K, V]{}
}
