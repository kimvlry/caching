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

	base := strategies.NewLRUCache[string, Product](10)

	filteredAndMapped := decorators.WithFilter(
		decorators.WithMap[string, Product](
			base,
			discounter,
			func() cache.IterableCache[string, Product] {
				return strategies.NewLRUCache[string, Product](10)
			},
		),
		filter,
		func() cache.IterableCache[string, Product] {
			return strategies.NewLRUCache[string, Product](10)
		},
	)

	cacheWithMetrics := decorators.WithMetrics(filteredAndMapped)

	setWithMapping := func(key string, p Product) error {
		mapped := discounter(p)
		if !filter(mapped) {
			return nil
		}
		return filteredAndMapped.Set(key, mapped)
	}

	_ = setWithMapping("p1", Product{Name: "Laptop", Price: 1000})
	_ = setWithMapping("p2", Product{Name: "Mouse", Price: 20})
	_ = setWithMapping("p3", Product{Name: "Monitor", Price: 300})
	_ = setWithMapping("p4", Product{Name: "Keyboard", Price: 80})

	fmt.Println("Fetching products:")
	printProduct(cacheWithMetrics, "p1")
	printProduct(cacheWithMetrics, "p2")
	printProduct(cacheWithMetrics, "p3")
	printProduct(cacheWithMetrics, "p4")
	printProduct(cacheWithMetrics, "p5")
	printProduct(cacheWithMetrics, "p1")

	fmt.Printf("\nStats:\n")
	fmt.Printf("  Hits: %d\n", cacheWithMetrics.GetHits())
	fmt.Printf("  Misses: %d\n", cacheWithMetrics.GetMisses())
	fmt.Printf("  Hit Rate: %.2f%%\n", cacheWithMetrics.HitRate()*100)

	if asIterable, ok := cacheWithMetrics.(cache.IterableCache[string, Product]); ok {
		total := decorators.WithReduce[string, Product, int](asIterable, 0, func(acc int, p Product) int {
			return acc + p.Price
		})
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
