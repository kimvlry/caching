// examples/03_functional/main.go
package main

import (
	"caching-labwork/cache/strategies"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"caching-labwork/cache"
	"caching-labwork/cache/decorators"
	"caching-labwork/cache/decorators/functional"
)

func main() {
	fmt.Println("=== Example 3: Functional Decorators & Advanced Composition ===\n")

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Пример 1: Map - трансформация значений
	fmt.Println("--- Functional Map ---")
	intCache := strategies.NewLRUCache[string, int](10)
	intCache.Set("price:apple", 100)
	intCache.Set("price:banana", 50)
	intCache.Set("price:orange", 75)

	// Map: int -> string (добавляем валюту)
	stringCache := functional.Map(
		intCache,
		func(price int) string {
			return fmt.Sprintf("$%.2f", float64(price)/100.0)
		},
		func() cache.Cache[string, string] {
			return strategies.NewLRUCache[string, string](10)
		},
	)

	fmt.Println("После Map (int -> string с форматированием):")
	if val, err := stringCache.Get("price:apple"); err == nil {
		fmt.Printf("  price:apple = %s\n", val)
	}
	if val, err := stringCache.Get("price:banana"); err == nil {
		fmt.Printf("  price:banana = %s\n", val)
	}

	// Пример 2: Filter - фильтрация элементов
	fmt.Println("\n--- Functional Filter ---")
	productCache := strategies.NewLRUCache[string, Product](10)
	productCache.Set("prod:1", Product{ID: 1, Name: "Laptop", Price: 1000, InStock: true})
	productCache.Set("prod:2", Product{ID: 2, Name: "Mouse", Price: 25, InStock: false})
	productCache.Set("prod:3", Product{ID: 3, Name: "Keyboard", Price: 75, InStock: true})
	productCache.Set("prod:4", Product{ID: 4, Name: "Monitor", Price: 300, InStock: true})

	// Filter: только товары в наличии
	inStockCache := functional.Filter(
		productCache,
		func(p Product) bool {
			return p.InStock
		},
		func() cache.Cache[string, Product] {
			return strategies.NewLRUCache[string, Product](10)
		},
	)

	fmt.Println("После Filter (только InStock=true):")
	if val, err := inStockCache.Get("prod:1"); err == nil {
		fmt.Printf("  prod:1: %s ($%d) - InStock=%v\n", val.Name, val.Price, val.InStock)
	}
	if _, err := inStockCache.Get("prod:2"); err != nil {
		fmt.Println("  prod:2: не найден (отфильтрован)")
	}
	if val, err := inStockCache.Get("prod:3"); err == nil {
		fmt.Printf("  prod:3: %s ($%d) - InStock=%v\n", val.Name, val.Price, val.InStock)
	}

	// Пример 3: Reduce - агрегация значений
	fmt.Println("\n--- Functional Reduce ---")
	priceCache := strategies.NewLRUCache[string, int](10)
	priceCache.Set("item:1", 100)
	priceCache.Set("item:2", 250)
	priceCache.Set("item:3", 75)
	priceCache.Set("item:4", 300)

	// Reduce: сумма всех цен
	totalPrice := functional.Reduce(
		priceCache,
		0,
		func(acc int, price int) int {
			return acc + price
		},
	)

	fmt.Printf("Общая сумма всех товаров: $%d\n", totalPrice)

	// Reduce: найти максимальную цену
	maxPrice := functional.Reduce(
		priceCache,
		0,
		func(max int, price int) int {
			if price > max {
				return price
			}
			return max
		},
	)

	fmt.Printf("Максимальная цена: $%d\n", maxPrice)

	// Пример 4: Композиция Filter + Map
	fmt.Println("\n--- Composition: Filter + Map ---")
	userCache := strategies.NewLRUCache[string, User](10)
	userCache.Set("user:1", User{ID: 1, Name: "Alice", Age: 25, Active: true})
	userCache.Set("user:2", User{ID: 2, Name: "Bob", Age: 17, Active: true})
	userCache.Set("user:3", User{ID: 3, Name: "Charlie", Age: 30, Active: false})
	userCache.Set("user:4", User{ID: 4, Name: "Diana", Age: 22, Active: true})

	// Шаг 1: Filter - только активные и совершеннолетние
	adultActiveUsers := functional.Filter(
		userCache,
		func(u User) bool {
			return u.Active && u.Age >= 18
		},
		func() cache.Cache[string, User] {
			return cache.NewLRUCache[string, User](10)
		},
	)

	// Шаг 2: Map - извлекаем только имена
	userNames := functional.Map(
		adultActiveUsers,
		func(u User) string {
			return u.Name
		},
		func() cache.Cache[string, string] {
			return cache.NewLRUCache[string, string](10)
		},
	)

	fmt.Println("После Filter (активные, 18+) + Map (имена):")
	if name, err := userNames.Get("user:1"); err == nil {
		fmt.Printf("  user:1: %s\n", name)
	}
	if _, err := userNames.Get("user:2"); err != nil {
		fmt.Println("  user:2: не найден (несовершеннолетний)")
	}
	if _, err := userNames.Get("user:3"); err != nil {
		fmt.Println("  user:3: не найден (неактивный)")
	}
	if name, err := userNames.Get("user:4"); err == nil {
		fmt.Printf("  user:4: %s\n", name)
	}

	// Пример 5: Сложная композиция с декораторами
	fmt.Println("\n--- Advanced Composition: Functional + Decorators ---")

	// Базовый кэш с продуктами
	baseProducts := strategies.NewLRUCache[string, Product](20)
	baseProducts.Set("p:1", Product{ID: 1, Name: "Gaming Laptop", Price: 1500, InStock: true})
	baseProducts.Set("p:2", Product{ID: 2, Name: "Office Mouse", Price: 20, InStock: true})
	baseProducts.Set("p:3", Product{ID: 3, Name: "Mechanical Keyboard", Price: 150, InStock: false})
	baseProducts.Set("p:4", Product{ID: 4, Name: "4K Monitor", Price: 500, InStock: true})
	baseProducts.Set("p:5", Product{ID: 5, Name: "Webcam", Price: 80, InStock: true})

	// Композиция:
	// 1. Filter: только в наличии и цена >= 100
	// 2. Map: преобразуем в DTO с дополнительной информацией
	// 3. Decorators: добавляем метрики и логирование

	filteredProducts := functional.Filter(
		baseProducts,
		func(p Product) bool {
			return p.InStock && p.Price >= 100
		},
		func() cache.Cache[string, Product] {
			return cache.NewLRUCache[string, Product](10)
		},
	)

	productDTOs := functional.Map(
		filteredProducts,
		func(p Product) ProductDTO {
			return ProductDTO{
				ID:          p.ID,
				DisplayName: strings.ToUpper(p.Name),
				PriceStr:    fmt.Sprintf("$%.2f", float64(p.Price)),
				Available:   p.InStock,
			}
		},
		func() cache.Cache[string, ProductDTO] {
			return cache.NewTTLCache[string, ProductDTO](10, 5*time.Minute)
		},
	)

	// Добавляем декораторы
	finalCache := decorators.WithDebugLogging(
		decorators.WithMetrics(productDTOs),
		logger,
	)

	fmt.Println("Финальный кэш (Filter + Map + Metrics + Logging):")

	// Проверяем что попало в финальный кэш
	if dto, err := finalCache.Get("p:1"); err == nil {
		fmt.Printf("  ✓ p:1: %s - %s\n", dto.DisplayName, dto.PriceStr)
	}
	if _, err := finalCache.Get("p:2"); err != nil {
		fmt.Println("  ✗ p:2: отфильтрован (цена < 100)")
	}
	if _, err := finalCache.Get("p:3"); err != nil {
		fmt.Println("  ✗ p:3: отфильтрован (нет в наличии)")
	}
	if dto, err := finalCache.Get("p:4"); err == nil {
		fmt.Printf("  ✓ p:4: %s - %s\n", dto.DisplayName, dto.PriceStr)
	}

	// Пример 6: Pipeline с несколькими преобразованиями
	fmt.Println("\n--- Data Processing Pipeline ---")

	rawData := cache.NewLRUCache[string, string](20)
	rawData.Set("data:1", "100")
	rawData.Set("data:2", "invalid")
	rawData.Set("data:3", "250")
	rawData.Set("data:4", "75")
	rawData.Set("data:5", "not_a_number")

	// Pipeline: string -> int -> filtered -> doubled

	// Шаг 1: Parse strings to ints (с фильтрацией невалидных)
	intData := cache.NewLRUCache[string, int](20)
	rawData.Range(func(k string, v string) bool {
		if num, err := strconv.Atoi(v); err == nil {
			intData.Set(k, num)
		}
		return true
	})

	// Шаг 2: Filter numbers > 50
	filtered := functional.Filter(
		intData,
		func(n int) bool { return n > 50 },
		func() cache.Cache[string, int] {
			return cache.NewLRUCache[string, int](20)
		},
	)

	// Шаг 3: Map - удвоить значения
	doubled := functional.Map(
		filtered,
		func(n int) int { return n * 2 },
		func() cache.Cache[string, int] {
			return cache.NewLRUCache[string, int](20)
		},
	)

	// Шаг 4: Reduce - сумма
	sum := functional.Reduce(doubled, 0, func(acc, val int) int {
		return acc + val
	})

	fmt.Println("Pipeline результаты:")
	fmt.Printf("  Исходные данные: 5 элементов\n")
	fmt.Printf("  После парсинга: 3 валидных числа\n")
	fmt.Printf("  После фильтра (>50): 2 элемента\n")
	fmt.Printf("  После удвоения и суммирования: %d\n", sum)

	// Пример 7: Real-world scenario - API Response Cache
	fmt.Println("\n--- Real World: API Response Cache ---")

	// Кэш с сырыми API ответами
	apiResponseCache := cache.NewTTLCache[string, APIResponse](100, 1*time.Minute)

	apiResponseCache.Set("api:/users/1", APIResponse{
		StatusCode: 200,
		Body:       `{"id":1,"name":"Alice","role":"admin"}`,
		Timestamp:  time.Now(),
	})
	apiResponseCache.Set("api:/users/2", APIResponse{
		StatusCode: 404,
		Body:       `{"error":"not found"}`,
		Timestamp:  time.Now(),
	})
	apiResponseCache.Set("api:/users/3", APIResponse{
		StatusCode: 200,
		Body:       `{"id":3,"name":"Charlie","role":"user"}`,
		Timestamp:  time.Now(),
	})

	// Pipeline: фильтруем успешные ответы и добавляем метрики
	successfulResponses := functional.Filter(
		apiResponseCache,
		func(resp APIResponse) bool {
			return resp.StatusCode >= 200 && resp.StatusCode < 300
		},
		func() cache.Cache[string, APIResponse] {
			return cache.NewTTLCache[string, APIResponse](100, 1*time.Minute)
		},
	)

	cachedAPI := decorators.WithMetrics(successfulResponses)

	fmt.Println("API Cache со статистикой:")
	cachedAPI.Get("api:/users/1")   // hit
	cachedAPI.Get("api:/users/2")   // miss (был отфильтрован)
	cachedAPI.Get("api:/users/3")   // hit
	cachedAPI.Get("api:/users/999") // miss

	fmt.Printf("  API Hits: %d\n", cachedAPI.GetHits())
	fmt.Printf("  API Misses: %d\n", cachedAPI.GetMisses())
	fmt.Printf("  Cache Hit Rate: %.2f%%\n", cachedAPI.HitRate()*100)

	fmt.Println("\n=== Пример 3 завершен ===")
}

// Data structures
type Product struct {
	ID      int
	Name    string
	Price   int
	InStock bool
}

type ProductDTO struct {
	ID          int
	DisplayName string
	PriceStr    string
	Available   bool
}

type User struct {
	ID     int
	Name   string
	Age    int
	Active bool
}

type APIResponse struct {
	StatusCode int
	Body       string
	Timestamp  time.Time
}
