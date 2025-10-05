package decorators

import (
	"bytes"
	"github.com/kimvlry/caching/cache/strategies"
	"log/slog"
	"strings"
	"testing"
)

func TestLoggingDecorator_Get(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	baseCache := strategies.NewLRUCache[string, int](10)
	baseCache.Set("key1", 42)

	loggingCache := WithDebugLogging(baseCache, logger)

	val, err := loggingCache.Get("key1")

	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val != 42 {
		t.Errorf("Value mismatch: got %d, want 42", val)
	}

	logOutput := buf.String()
	if !strings.Contains(logOutput, "Get method called") {
		t.Error("Log should contain 'Get method called'")
	}
	if !strings.Contains(logOutput, "key1") {
		t.Error("Log should contain the key 'key1'")
	}
	if !strings.Contains(logOutput, "Get method returned a value") {
		t.Error("Log should contain 'Get method returned a value'")
	}
}

func TestLoggingDecorator_GetNonExistent(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	baseCache := strategies.NewLRUCache[string, int](10)
	loggingCache := WithDebugLogging(baseCache, logger)

	_, err := loggingCache.Get("nonexistent")

	if err != nil {
		t.Logf("Expected error: %v", err)
	}

	logOutput := buf.String()
	if !strings.Contains(logOutput, "Get method called") {
		t.Error("Log should contain 'Get method called'")
	}
	if !strings.Contains(logOutput, "Get method returned an error") {
		t.Error("Log should contain 'Get method returned an error'")
	}
}

func TestLoggingDecorator_Set(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	baseCache := strategies.NewLRUCache[string, string](10)
	loggingCache := WithDebugLogging(baseCache, logger)

	err := loggingCache.Set("key1", "value1")

	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	logOutput := buf.String()
	if !strings.Contains(logOutput, "Set method called") {
		t.Error("Log should contain 'Set method called'")
	}
	if !strings.Contains(logOutput, "key1") {
		t.Error("Log should contain the key 'key1'")
	}
	if !strings.Contains(logOutput, "Set method returned a value") {
		t.Error("Log should contain 'Set method returned a value'")
	}
}

func TestLoggingDecorator_Delete(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	baseCache := strategies.NewLRUCache[string, int](10)
	baseCache.Set("key1", 42)

	loggingCache := WithDebugLogging(baseCache, logger)
	buf.Reset()

	err := loggingCache.Delete("key1")

	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	logOutput := buf.String()
	if !strings.Contains(logOutput, "Delete method called") {
		t.Error("Log should contain 'Delete method called'")
	}
	if !strings.Contains(logOutput, "key1") {
		t.Error("Log should contain the key 'key1'")
	}
}

func TestLoggingDecorator_DeleteNonExistent(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	baseCache := strategies.NewLRUCache[string, int](10)
	loggingCache := WithDebugLogging(baseCache, logger)

	err := loggingCache.Delete("nonexistent")

	if err != nil {
		t.Logf("Delete returned: %v", err)
	}

	logOutput := buf.String()
	if !strings.Contains(logOutput, "Delete method called") {
		t.Error("Log should contain 'Delete method called'")
	}
}

func TestLoggingDecorator_Clear(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	baseCache := strategies.NewLRUCache[string, int](10)
	baseCache.Set("key1", 1)
	baseCache.Set("key2", 2)

	loggingCache := WithDebugLogging(baseCache, logger)
	buf.Reset()

	loggingCache.Clear()

	logOutput := buf.String()
	if !strings.Contains(logOutput, "Clear method called") {
		t.Error("Log should contain 'Clear method called'")
	}

	if _, err := loggingCache.Get("key1"); err == nil {
		t.Error("Cache should be empty after Clear")
	}
}

func TestLoggingDecorator_CacheTypeInLog(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	baseCache := strategies.NewLRUCache[string, int](10)
	loggingCache := WithDebugLogging(baseCache, logger)

	loggingCache.Get("key1")

	logOutput := buf.String()
	if !strings.Contains(logOutput, "cache=") {
		t.Error("Log should contain cache type information")
	}
}

func TestLoggingDecorator_LogLevelDebug(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	baseCache := strategies.NewLRUCache[string, int](10)
	loggingCache := WithDebugLogging(baseCache, logger)

	loggingCache.Set("key1", 42)
	loggingCache.Get("key1")

	logOutput := buf.String()
	debugCount := strings.Count(logOutput, "level=DEBUG")
	if debugCount < 2 {
		t.Errorf("Expected at least 2 DEBUG logs, got %d", debugCount)
	}
}

func TestLoggingDecorator_LogLevelWarn(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelWarn,
	}))

	baseCache := strategies.NewLRUCache[string, int](10)
	loggingCache := WithDebugLogging(baseCache, logger)

	loggingCache.Get("nonexistent")
	loggingCache.Set("key1", 42)

	logOutput := buf.String()
	if !strings.Contains(logOutput, "level=WARN") {
		t.Error("Log should contain WARN level")
	}
	if strings.Contains(logOutput, "Set method called") {
		t.Error("DEBUG logs should not appear with WARN level")
	}
}

func TestLoggingDecorator_MultipleOperations(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	baseCache := strategies.NewLRUCache[string, int](10)
	loggingCache := WithDebugLogging(baseCache, logger)

	loggingCache.Set("key1", 1)
	loggingCache.Set("key2", 2)
	loggingCache.Get("key1")
	loggingCache.Get("key2")
	loggingCache.Delete("key1")

	logOutput := buf.String()

	setCount := strings.Count(logOutput, "Set method called")
	if setCount != 2 {
		t.Errorf("Expected 2 Set operations in log, got %d", setCount)
	}

	getCount := strings.Count(logOutput, "Get method called")
	if getCount != 2 {
		t.Errorf("Expected 2 Get operations in log, got %d", getCount)
	}

	deleteCount := strings.Count(logOutput, "Delete method called")
	if deleteCount != 1 {
		t.Errorf("Expected 1 Delete operation in log, got %d", deleteCount)
	}
}

func TestLoggingDecorator_CompositionWithMetrics(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	baseCache := strategies.NewLRUCache[string, int](10)
	metricsCache := WithMetrics(baseCache)
	loggingCache := WithDebugLogging(metricsCache, logger)

	loggingCache.Set("key1", 42)
	loggingCache.Get("key1")
	loggingCache.Get("nonexistent")

	logOutput := buf.String()
	if !strings.Contains(logOutput, "Get method called") {
		t.Error("Logging should work in composition")
	}

	if metricsCache.GetHits() != 1 {
		t.Errorf("Expected 1 hit, got %d", metricsCache.GetHits())
	}
	if metricsCache.GetMisses() != 1 {
		t.Errorf("Expected 1 miss, got %d", metricsCache.GetMisses())
	}
}
