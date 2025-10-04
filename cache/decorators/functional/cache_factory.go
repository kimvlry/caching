package functional

import (
	"caching-labwork/cache"
)

type CacheFactory[K comparable, V any] func() cache.Cache[K, V]
