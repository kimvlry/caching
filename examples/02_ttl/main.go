package main

import (
	"fmt"
	"github.com/kimvlry/caching/cache/decorators"
	"github.com/kimvlry/caching/cache/strategies"
	"time"
)

type UserSession struct {
	UserID    string
	Username  string
	LoginTime time.Time
}

func main() {
	baseTTL := strategies.NewTtlCache[string, UserSession](100, 10*time.Second)()
	cache := decorators.WithMetrics(baseTTL)

	session1 := UserSession{UserID: "user123", Username: "alice", LoginTime: time.Now()}
	session2 := UserSession{UserID: "user456", Username: "bob", LoginTime: time.Now()}
	session3 := UserSession{UserID: "user789", Username: "charlie", LoginTime: time.Now()}

	_ = cache.Set("sess_long", session1)
	fmt.Println("✓ Added long-lived session (10s NewTtlCache)")

	if ttlCache, ok := cache.(strategies.TTLCache[string, UserSession]); ok {
		_ = ttlCache.SetWithTTL("sess_short", session2, 3*time.Second)
		fmt.Println("✓ Added short-lived session (3s NewTtlCache)")
	}

	if ttlCache, ok := cache.(strategies.TTLCache[string, UserSession]); ok {
		_ = ttlCache.SetWithTTL("sess_verylong", session3, 20*time.Second)
		fmt.Println("✓ Added very-long-lived session (20s NewTtlCache)")
	}

	fmt.Println("\n--- Checking immediately ---")
	checkSession(cache, "sess_long")
	checkSession(cache, "sess_short")
	checkSession(cache, "sess_verylong")

	fmt.Printf("\nStats: Hits: %d, Misses: %d, Hit Rate: %.2f%%\n",
		cache.GetHits(), cache.GetMisses(), cache.HitRate()*100)

	fmt.Println("\n--- Waiting 4 seconds (short session expires) ---")
	time.Sleep(4 * time.Second)

	checkSession(cache, "sess_long")
	checkSession(cache, "sess_short")
	checkSession(cache, "sess_verylong")

	fmt.Printf("\nStats: Hits: %d, Misses: %d, Evictions: %d, Hit Rate: %.2f%%\n",
		cache.GetHits(), cache.GetMisses(), cache.GetEvictions(), cache.HitRate()*100)

	fmt.Println("\n--- Waiting 7 more seconds (long session expires) ---")
	time.Sleep(7 * time.Second)

	checkSession(cache, "sess_long")
	checkSession(cache, "sess_short")
	checkSession(cache, "sess_verylong")

	fmt.Printf("\nFinal Stats: Hits: %d, Misses: %d, Evictions: %d, Hit Rate: %.2f%%\n",
		cache.GetHits(), cache.GetMisses(), cache.GetEvictions(), cache.HitRate()*100)
}

func checkSession(c interface {
	Get(string) (UserSession, error)
}, key string) {
	if sess, err := c.Get(key); err == nil {
		fmt.Printf("  ✓ %s: %s (user: %s)\n", key, sess.UserID, sess.Username)
	} else {
		fmt.Printf("  ✗ %s: expired or not found\n", key)
	}
}
