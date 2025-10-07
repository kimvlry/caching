package main

import (
	"fmt"
	"github.com/kimvlry/caching/cache/decorators"
	"github.com/kimvlry/caching/cache/strategies"
	"log"
	"log/slog"
	"os"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	fmt.Println("--- Metrics Decorator ---")

	metricsCache :=
		decorators.WithDebugLogging(
			decorators.WithMetrics(
				strategies.NewLRUCache[string, int](5)),
			logger,
		)

	_ = metricsCache.Set("key1", 100)
	_ = metricsCache.Set("key2", 200)
	_ = metricsCache.Set("key3", 300)

	_, _ = metricsCache.Get("key1") // hit
	_, _ = metricsCache.Get("key2") // hit
	_, _ = metricsCache.Get("key1") // hit

	_, _ = metricsCache.Get("key999") // miss
	_, _ = metricsCache.Get("key888") // miss

	asAwareCache, ok := any(metricsCache).(decorators.AwareCache[string, int])
	if !ok {
		log.Fatalf("type assertion failed")
	}
	fmt.Println(asAwareCache.GetHits())
}
