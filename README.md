# Cache Library

A flexible, generic caching library in Go that supports multiple eviction strategies,
decorators, and functional programming patterns.

## Features

### Cache Eviction Strategies

* **LRU (Least Recently Used)** - Evicts the least recently accessed items
* **FIFO (First In, First Out)** - Evicts the oldest items first
* **LFU (Least Frequently Used)** - Evicts the least frequently accessed items
* **TTL (Time To Live)** - Automatically expires entries based on time
* **ARC (Adaptive Replacement Cache)** - Adaptive strategy combining LRU and LFU principles

### Basic Decorators

* **Metrics** - Tracks hits, misses, evictions, and hit rate
* **Logging** - Provides debug logging for all cache operations
* **Compression** - Automatically compresses data using gzip with JSON serialization
* **Bloom Filter**

### Functional Decorators

* **Map** - Transforms cache values (immutable, produces a new cache)
* **Filter** - Filters cache entries by predicate (immutable)
* **Reduce** - Aggregates cache values into a single result (returns a value, not a new cache)

Functional decorators (`Map`, `Filter`) always produce new caches and never modify the source cache.

### Advanced Features

* **Generic types** - Fully type-safe with Go generics
* **Observable caches** - Event-driven architecture for metrics and observers
* **Decorator composition** - Chain multiple decorators seamlessly
* **Factory pattern** - Flexible cache creation through closures
* **Thread-safe metrics** - Atomic operations for accurate tracking

## Quick Start

### Installation

```bash
go get github.com/kimvlry/caching
```

### Creating Cache Instances


```go
import (
    "github.com/kimvlry/caching/strategies"
    "time"
)
...
// example: 
// FIFO Cache
fifoCache := strategies.NewFifoCache[string, int]()

// TTL Cache (5-minute expiration)
ttlCache := cache.NewTtlCache[string, int](100, 5*time.Minute)()
```

Note that cache creators are functions returning other functions, allowing them to be used
not only for initialization but also as parameters for functional decorators:

```go
import (
    "github.com/kimvlry/caching/cache/strategies"
    "github.com/kimvlry/caching/cache/decorators"
)

...

filteredAndMapped := decorators.WithFilter(
	decorators.WithMap(
		base,
		discounter,
		strategies.NewLruCache[string, int](100),
	),
	filter,
	strategies.NewFifoCache[string, int](100) ,
)
```
You can also create an output cache using a different strategy from the original like in the example above.

## 🤹🏻‍♀️ Decorators

### Metrics Decorator

**Implementation note: Observer Pattern**

Observable caches (all the strategies do implement this interface) emit events that can be used for metrics tracking:
it requires type assertion
```go
// Metrics decorator subscribes to emitted events
if obs, ok := cache.(ObservableCache[K, V]); ok {
    obs.OnEvent(func(event CacheEvent[K, V]) {
        // Handle hits, misses, and evictions (search examples in `examples` dir or tests)
    })
}
```
### Decorators composition
```go
import (
    "github.com/kimvlry/caching/cache/strategies"
    "github.com/kimvlry/caching/cache/decorators"
)

...

baseCache := strategies.NewLruCache[int, int](10)()

// Chain multiple decorators
cache := decorators.WithDebugLogging(
    decorators.WithMetrics(
        baseCache, 
    ),
    logger,
)

// Use the composed cache
cache.Set("key1", 42)
cache.Get("key1")
```

### More Examples

Various usage examples are provided in the `examples` directory as ready-to-run scripts.

## Running Tests

### Run all tests

```bash
go test ./...
```

### Run tests with coverage

```bash
go test -cover ./...
```

### Run tests for a specific package

```bash
go test ./cache/decorators
go test ./cache/decorators/functional
```

### 🚧 TODO

* Benchmarks
* Full thread-safety support
