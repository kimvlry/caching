package decorators

import (
	"fmt"
	"hash/fnv"

	"github.com/bits-and-blooms/bloom/v3"
	"github.com/kimvlry/caching/cache"
	"github.com/kimvlry/caching/cache/strategies/common"
)

type bloomDecorator[K comparable, V any] struct {
	cacheWrappee cache.IterableCache[K, V]
	filter       *bloom.BloomFilter
	hasher       func(K) []byte
}

// WithBloomFilter creates a Bloom filter decorator for the given cache.
// The Bloom filter is most effective when:
// - Cache miss rate is high (>70%)
// - The underlying cache has expensive lookups (e.g., disk/network access)
// - Memory is abundant (Bloom filter adds memory overhead)
//
// For fast in-memory caches with low miss rates, the overhead may outweigh benefits.
func WithBloomFilter[K comparable, V any](
	wrappee cache.IterableCache[K, V],
	expectedEntries uint,
	falsePositiveRate float64,
) cache.IterableCache[K, V] {

	filter := bloom.NewWithEstimates(expectedEntries, falsePositiveRate)

	var hasher func(K) []byte
	switch any(*new(K)).(type) {
	case string:
		// For strings, direct byte conversion is fastest
		hasher = func(key K) []byte {
			return []byte(any(key).(string))
		}
	default:
		// For other types, use FNV hash for consistent byte representation
		hasher = func(key K) []byte {
			h := fnv.New64a()
			fmt.Fprintf(h, "%v", key)
			return h.Sum(nil)
		}
	}

	decorator := &bloomDecorator[K, V]{
		cacheWrappee: wrappee,
		filter:       filter,
		hasher:       hasher,
	}

	// Synchronize Bloom filter with cache evictions
	// Without this, evicted keys remain in the Bloom filter, causing false positives
	// that make the filter progressively less useful over time

	// TODO: rebuild effectively (not after every eviction??)
	if observable, ok := any(wrappee).(cache.ObservableCache[K, V]); ok {
		observable.OnEvent(func(event cache.Event[K, V]) {
			switch event.Type {
			case cache.EventTypeEviction:
				decorator.rebuildFilter()
			}
		})
	}

	return decorator
}

// rebuildFilter reconstructs the Bloom filter from the current cache contents.
func (b *bloomDecorator[K, V]) rebuildFilter() {
	b.filter.ClearAll()
	b.cacheWrappee.Range(func(key K, _ V) bool {
		b.filter.Add(b.hasher(key))
		return true
	})
}

func (b *bloomDecorator[K, V]) Get(key K) (V, error) {
	if !b.filter.Test(b.hasher(key)) {
		var zero V
		return zero, common.ErrKeyNotFound
	}
	return b.cacheWrappee.Get(key)
}

func (b *bloomDecorator[K, V]) Set(key K, value V) error {
	b.filter.Add(b.hasher(key))
	return b.cacheWrappee.Set(key, value)
}

// Delete removes from cache but not from Bloom filter.
// The key will remain in the filter (causing potential false positives)
// until the next eviction triggers a rebuild via rebuildFilter().
func (b *bloomDecorator[K, V]) Delete(key K) error {
	return b.cacheWrappee.Delete(key)
}

func (b *bloomDecorator[K, V]) Clear() {
	b.cacheWrappee.Clear()
	b.filter.ClearAll()
}

func (b *bloomDecorator[K, V]) Range(fn func(K, V) bool) {
	if iterable, ok := any(b.cacheWrappee).(cache.IterableCache[K, V]); ok {
		iterable.Range(fn)
	}
}
