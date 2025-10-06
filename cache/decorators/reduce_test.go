package decorators

import (
	"fmt"
	"github.com/kimvlry/caching/cache/decorators/common"
	"strings"
	"testing"

	"github.com/kimvlry/caching/cache"
	"github.com/kimvlry/caching/cache/strategies"
)

func TestWithReduce_Sum(t *testing.T) {
	baseCache := strategies.NewLRUCache[string, int](10)
	_ = baseCache.Set("a", 1)
	_ = baseCache.Set("b", 2)
	_ = baseCache.Set("c", 3)
	_ = baseCache.Set("d", 4)
	_ = baseCache.Set("e", 5)

	sum := WithReduce[string, int, int](
		baseCache,
		0,
		func(acc int, value int) int { return acc + value },
	)

	if sum != 15 {
		t.Errorf("Expected sum=15, got %d", sum)
	}
}

func TestWithReduce_Product(t *testing.T) {
	baseCache := strategies.NewLRUCache[string, int](10)
	_ = baseCache.Set("a", 2)
	_ = baseCache.Set("b", 3)
	_ = baseCache.Set("c", 4)

	product := WithReduce[string, int, int](
		baseCache,
		1,
		func(acc int, value int) int { return acc * value },
	)

	if product != 24 {
		t.Errorf("Expected product=24, got %d", product)
	}
}

func TestWithReduce_Max(t *testing.T) {
	baseCache := strategies.NewLRUCache[string, int](10)
	_ = baseCache.Set("a", 5)
	_ = baseCache.Set("b", 12)
	_ = baseCache.Set("c", 3)
	_ = baseCache.Set("d", 8)

	reduced := WithReduce[string, int, int](
		baseCache,
		0,
		func(acc int, value int) int {
			if value > acc {
				return value
			}
			return acc
		},
	)

	if reduced != 12 {
		t.Errorf("Expected reduced=12, got %d", reduced)
	}
}

func TestWithReduce_Min(t *testing.T) {
	baseCache := strategies.NewLRUCache[string, int](10)
	_ = baseCache.Set("a", 5)
	_ = baseCache.Set("b", 12)
	_ = baseCache.Set("c", 3)
	_ = baseCache.Set("d", 8)

	reduced := WithReduce[string, int, int](
		baseCache,
		999999,
		func(acc int, value int) int {
			if value < acc {
				return value
			}
			return acc
		},
	)

	if reduced != 3 {
		t.Errorf("Expected reduced=3, got %d", reduced)
	}
}

func TestWithReduce_StringConcatenation(t *testing.T) {
	baseCache := strategies.NewLRUCache[string, string](10)
	_ = baseCache.Set("1", "Hello")
	_ = baseCache.Set("2", " ")
	_ = baseCache.Set("3", "World")

	result := WithReduce[string, string, string](
		baseCache,
		"",
		func(acc string, value string) string { return acc + value },
	)

	if !strings.Contains(result, "Hello") || !strings.Contains(result, "World") {
		t.Errorf("Result should contain 'Hello' and 'World', got '%s'", result)
	}
}

func TestWithReduce_ComplexAccumulator(t *testing.T) {
	type Stats struct {
		Count int
		Sum   int
		Avg   float64
	}

	baseCache := strategies.NewLRUCache[string, int](10)
	_ = baseCache.Set("a", 10)
	_ = baseCache.Set("b", 20)
	_ = baseCache.Set("c", 30)
	_ = baseCache.Set("d", 40)

	stats := WithReduce[string, int, Stats](
		baseCache,
		Stats{Count: 0, Sum: 0, Avg: 0},
		func(acc Stats, value int) Stats {
			newCount := acc.Count + 1
			newSum := acc.Sum + value
			return Stats{
				Count: newCount,
				Sum:   newSum,
				Avg:   float64(newSum) / float64(newCount),
			}
		},
	)

	if stats.Count != 4 {
		t.Errorf("Expected Count=4, got %d", stats.Count)
	}
	if stats.Sum != 100 {
		t.Errorf("Expected Sum=100, got %d", stats.Sum)
	}
	if stats.Avg != 25.0 {
		t.Errorf("Expected Avg=25.0, got %f", stats.Avg)
	}
}

func TestWithReduce_EmptyCache(t *testing.T) {
	baseCache := strategies.NewLRUCache[string, int](10)

	sum := WithReduce[string, int, int](
		baseCache,
		42,
		func(acc int, value int) int { return acc + value },
	)

	if sum != 42 {
		t.Errorf("Expected initial value 42, got %d", sum)
	}
}

func TestWithReduce_SingleElement(t *testing.T) {
	baseCache := strategies.NewLRUCache[string, int](10)
	_ = baseCache.Set("only", 100)

	result := WithReduce[string, int, int](
		baseCache,
		0,
		func(acc int, value int) int { return acc + value },
	)

	if result != 100 {
		t.Errorf("Expected 100, got %d", result)
	}
}

func TestWithReduce_WithFilteredCache(t *testing.T) {
	baseCache := strategies.NewLRUCache[string, int](10)
	for i := 1; i <= 10; i++ {
		_ = baseCache.Set(fmt.Sprintf("key%d", i), i)
	}

	filtered := WithFilter[string, int](
		baseCache,
		func(v int) bool { return v%2 == 0 },
	)

	sum := WithReduce[string, int, int](
		filtered,
		0,
		func(acc int, value int) int { return acc + value },
	)

	if sum != 30 {
		t.Errorf("Expected sum=30, got %d", sum)
	}
}

func TestWithReduce_WithMappedCache(t *testing.T) {
	baseCache := strategies.NewLRUCache[string, int](10)
	_ = baseCache.Set("a", 1)
	_ = baseCache.Set("b", 2)
	_ = baseCache.Set("c", 3)

	mapped := WithMap[string, int](
		baseCache,
		func(v int) int { return v * 2 },
	)

	sum := WithReduce[string, int, int](
		mapped,
		0,
		func(acc int, value int) int { return acc + value },
	)

	if sum != 12 {
		t.Errorf("Expected sum=12, got %d", sum)
	}
}

func TestWithReduce_CollectToSlice(t *testing.T) {
	baseCache := strategies.NewLRUCache[string, int](10)
	_ = baseCache.Set("a", 1)
	_ = baseCache.Set("b", 2)
	_ = baseCache.Set("c", 3)

	values := WithReduce[string, int, []int](
		baseCache,
		[]int{},
		func(acc []int, value int) []int {
			return append(acc, value)
		},
	)

	if len(values) != 3 {
		t.Errorf("Expected 3 values, got %d", len(values))
	}

	sum := 0
	for _, v := range values {
		sum += v
	}
	if sum != 6 {
		t.Errorf("Expected sum of values = 6, got %d", sum)
	}
}

func TestWithReduce_CollectToMap(t *testing.T) {
	type User struct {
		ID   int
		Name string
	}

	baseCache := strategies.NewLRUCache[string, User](10)
	_ = baseCache.Set("u1", User{ID: 1, Name: "Alice"})
	_ = baseCache.Set("u2", User{ID: 2, Name: "Bob"})
	_ = baseCache.Set("u3", User{ID: 3, Name: "Charlie"})

	userMap := WithReduce[string, User, map[int]string](
		baseCache,
		make(map[int]string),
		func(acc map[int]string, user User) map[int]string {
			acc[user.ID] = user.Name
			return acc
		},
	)

	if len(userMap) != 3 {
		t.Errorf("Expected 3 users, got %d", len(userMap))
	}

	if userMap[1] != "Alice" {
		t.Errorf("Expected user 1 = Alice, got %s", userMap[1])
	}
	if userMap[2] != "Bob" {
		t.Errorf("Expected user 2 = Bob, got %s", userMap[2])
	}
	if userMap[3] != "Charlie" {
		t.Errorf("Expected user 3 = Charlie, got %s", userMap[3])
	}
}

func TestWithReduce_CountOccurrences(t *testing.T) {
	baseCache := strategies.NewLRUCache[string, string](10)
	_ = baseCache.Set("1", "apple")
	_ = baseCache.Set("2", "banana")
	_ = baseCache.Set("3", "apple")
	_ = baseCache.Set("4", "orange")
	_ = baseCache.Set("5", "apple")
	_ = baseCache.Set("6", "banana")

	counts := WithReduce[string, string, map[string]int](
		baseCache,
		make(map[string]int),
		func(acc map[string]int, fruit string) map[string]int {
			acc[fruit]++
			return acc
		},
	)

	if counts["apple"] != 3 {
		t.Errorf("Expected 3 apples, got %d", counts["apple"])
	}
	if counts["banana"] != 2 {
		t.Errorf("Expected 2 bananas, got %d", counts["banana"])
	}
	if counts["orange"] != 1 {
		t.Errorf("Expected 1 orange, got %d", counts["orange"])
	}
}

func TestIntegration_FilterMapReduce(t *testing.T) {
	baseCache := strategies.NewLRUCache[string, int](20)
	for i := 1; i <= 20; i++ {
		_ = baseCache.Set(fmt.Sprintf("num%d", i), i)
	}

	result := WithReduce[string, int, int](
		WithMap[string, int](
			WithFilter[string, int](
				baseCache,
				func(v int) bool { return v > 10 },
			),
			func(v int) int { return v * 2 },
		),
		0,
		func(acc int, value int) int { return acc + value },
	)

	if result != 310 {
		t.Errorf("Expected 310, got %d", result)
	}
}

func TestIntegration_MultipleFiltersWithReduce(t *testing.T) {
	baseCache := strategies.NewLRUCache[string, int](100)
	for i := 1; i <= 100; i++ {
		_ = baseCache.Set(fmt.Sprintf("num%d", i), i)
	}

	result := WithReduce[string, int, int](
		WithFilter[string, int](
			WithFilter[string, int](
				WithFilter[string, int](
					baseCache,
					func(v int) bool { return v%2 == 0 },
				),
				func(v int) bool { return v > 50 },
			),
			func(v int) bool { return v <= 80 },
		),
		0,
		func(acc int, value int) int { return acc + value },
	)

	if result != 990 {
		t.Errorf("Expected 990, got %d", result)
	}
}

func TestIntegration_SnapshotWithReduce(t *testing.T) {
	baseCache := strategies.NewLRUCache[string, int](10)
	for i := 1; i <= 10; i++ {
		_ = baseCache.Set(fmt.Sprintf("key%d", i), i)
	}

	composed := WithMap[string, int](
		WithFilter[string, int](
			baseCache,
			func(v int) bool { return v > 5 },
		),
		func(v int) int { return v * 10 },
	)

	snapshot := common.Snapshot[string, int](
		composed,
		func() cache.IterableCache[string, int] {
			return strategies.NewLRUCache[string, int](5)
		},
	)

	sum := WithReduce[string, int, int](
		snapshot,
		0,
		func(acc int, value int) int { return acc + value },
	)

	if sum != 400 {
		t.Errorf("Expected sum=400, got %d", sum)
	}
}
