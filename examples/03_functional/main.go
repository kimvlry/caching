package main

import (
	"fmt"
	"github.com/kimvlry/caching/cache"
	"github.com/kimvlry/caching/cache/decorators"
	"github.com/kimvlry/caching/cache/strategies"
)

type Product struct {
	Name  string
	Price int
}

func main() {
	discounter := func(p Product) Product {
		return Product{Name: p.Name, Price: int(float64(p.Price) * 0.5)}
	}
	filter := func(p Product) bool {
		return p.Price > 50
	}

	baseCache := decorators.WithMetrics(
		decorators.WithFilter(
			decorators.WithMap[string, Product](
				strategies.NewLRUCache[string, Product](4), discounter,
			),
			filter,
		),
	)

	_ = baseCache.Set("p1", Product{Name: "Laptop", Price: 1000})
	_ = baseCache.Set("p2", Product{Name: "Mouse", Price: 20})
	_ = baseCache.Set("p3", Product{Name: "Monitor", Price: 300})
	_ = baseCache.Set("p4", Product{Name: "Keyboard", Price: 80})

	fmt.Println("Fetching products:")
	printProduct(baseCache, "p1") // Hit
	printProduct(baseCache, "p2") // Miss (doesn't pass filter)
	printProduct(baseCache, "p3") // Hit
	printProduct(baseCache, "p4") // Hit
	printProduct(baseCache, "p5") // Miss (doesn't exist)
	printProduct(baseCache, "p1") // Hit

	fmt.Printf("\nStats:\n")
	fmt.Printf("  Hits: %d\n", baseCache.GetHits())
	fmt.Printf("  Misses: %d\n", baseCache.GetMisses())
	fmt.Printf("  Hit Rate: %.2f%%\n", baseCache.HitRate()*100)

	if asIterable, ok := baseCache.(cache.IterableCache[string, Product]); ok {
		totalSumCalculator := func(acc int, p Product) int {
			return acc + p.Price
		}
		total := decorators.WithReduce[string, Product, int](asIterable, 0, totalSumCalculator)
		fmt.Printf("\nTotal value: $%d\n", total)
	}
}

func printProduct(c cache.Cache[string, Product], key string) {
	if p, err := c.Get(key); err == nil {
		fmt.Printf("%s: %s - $%d\n", key, p.Name, p.Price)
	} else {
		fmt.Printf("%s: not found or filtered\n", key)
	}
}
