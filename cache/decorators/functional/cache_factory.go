package functional

import "github.com/kimvlry/caching/cache"

type CacheFactory[K comparable, V any] func() cache.Cache[K, V]
