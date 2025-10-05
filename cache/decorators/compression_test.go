package decorators

import (
	"encoding/json"
	"github.com/kimvlry/caching/cache"
	"github.com/kimvlry/caching/cache/strategies"
	"testing"
)

type JSONSerializer[V any] struct{}

func (s JSONSerializer[V]) Marshal(v V) ([]byte, error) {
	return json.Marshal(v)
}

func (s JSONSerializer[V]) Unmarshal(data []byte) (V, error) {
	var v V
	err := json.Unmarshal(data, &v)
	return v, err
}

type TestData struct {
	ID   int
	Name string
	Tags []string
}

func TestCompressionDecorator_SetAndGet(t *testing.T) {
	baseCache := strategies.NewLRUCache[string, []byte](10)
	compCache := WithCompression(baseCache, JSONSerializer[TestData]{})

	testData := TestData{
		ID:   42,
		Name: "Test Item",
		Tags: []string{"tag1", "tag2", "tag3"},
	}

	err := compCache.Set("key1", testData)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	retrieved, err := compCache.Get("key1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.ID != testData.ID {
		t.Errorf("ID mismatch: got %d, want %d", retrieved.ID, testData.ID)
	}
	if retrieved.Name != testData.Name {
		t.Errorf("Name mismatch: got %s, want %s", retrieved.Name, testData.Name)
	}
	if len(retrieved.Tags) != len(testData.Tags) {
		t.Errorf("Tags length mismatch: got %d, want %d", len(retrieved.Tags), len(testData.Tags))
	}
}

func TestCompressionDecorator_GetNonExistent(t *testing.T) {
	baseCache := strategies.NewLRUCache[string, []byte](10)
	compCache := WithCompression(baseCache, JSONSerializer[TestData]{})

	_, err := compCache.Get("nonexistent")

	if err == nil {
		t.Error("Expected error for non-existent key, got nil")
	}
}

func TestCompressionDecorator_Delete(t *testing.T) {
	baseCache := strategies.NewLRUCache[string, []byte](10)
	compCache := WithCompression(baseCache, JSONSerializer[TestData]{})

	testData := TestData{ID: 1, Name: "Test"}
	_ = compCache.Set("key1", testData)

	err := compCache.Delete("key1")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = compCache.Get("key1")

	if err == nil {
		t.Error("Expected error after delete, got nil")
	}
}

func TestCompressionDecorator_Clear(t *testing.T) {
	baseCache := strategies.NewLRUCache[string, []byte](10)
	compCache := WithCompression(baseCache, JSONSerializer[TestData]{})

	_ = compCache.Set("key1", TestData{ID: 1, Name: "Test1"})
	_ = compCache.Set("key2", TestData{ID: 2, Name: "Test2"})
	_ = compCache.Set("key3", TestData{ID: 3, Name: "Test3"})

	compCache.Clear()

	for _, key := range []string{"key1", "key2", "key3"} {
		if _, err := compCache.Get(key); err == nil {
			t.Errorf("Expected error for key %s after Clear, got nil", key)
		}
	}
}

func TestCompressionDecorator_Events(t *testing.T) {
	baseCache := strategies.NewLRUCache[string, []byte](10)
	compCache := WithCompression(baseCache, JSONSerializer[TestData]{})

	var readBytesSize, compressedBytesSize int
	compCache.OnEvent(func(event cache.Event[string, TestData]) {
		switch event.Type {
		case cache.EventTypeReadBytes:
			readBytesSize = event.Size
		case cache.EventTypeCompressBytes:
			compressedBytesSize = event.Size
		}
	})

	testData := TestData{
		ID:   42,
		Name: "Long string to ensure compression: " + string(make([]byte, 1000)),
		Tags: []string{"tag1", "tag2", "tag3"},
	}

	err := compCache.Set("key1", testData)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	if readBytesSize == 0 {
		t.Error("ReadBytes event was not emitted")
	}
	if compressedBytesSize == 0 {
		t.Error("CompressBytes event was not emitted")
	}
	if compressedBytesSize >= readBytesSize {
		t.Logf("Warning: Compressed size (%d) >= raw size (%d). Data may not be compressible.",
			compressedBytesSize, readBytesSize)
	}
}

func TestCompressionDecorator_CompressionActuallyWorks(t *testing.T) {
	baseCache := strategies.NewLRUCache[string, []byte](10)
	compCache := WithCompression(baseCache, JSONSerializer[TestData]{})

	largeData := TestData{
		ID:   1,
		Name: string(make([]byte, 10000)),
		Tags: make([]string, 100),
	}
	for i := range largeData.Tags {
		largeData.Tags[i] = "same_tag"
	}

	var rawSize, compressedSize int
	compCache.OnEvent(func(event cache.Event[string, TestData]) {
		switch event.Type {
		case cache.EventTypeReadBytes:
			rawSize = event.Size
		case cache.EventTypeCompressBytes:
			compressedSize = event.Size
		}
	})

	err := compCache.Set("large", largeData)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	compressionRatio := float64(compressedSize) / float64(rawSize)
	t.Logf("Compression: %d bytes -> %d bytes (ratio: %.2f%%)",
		rawSize, compressedSize, compressionRatio*100)

	if compressionRatio > 0.5 {
		t.Errorf("Compression ratio too high: %.2f%% (expected < 50%% for repetitive data)",
			compressionRatio*100)
	}
}

func TestCompressionDecorator_InvalidJSON(t *testing.T) {
	baseCache := strategies.NewLRUCache[string, []byte](10)
	_ = baseCache.Set("invalid", []byte("not compressed data"))
	compCache := WithCompression(baseCache, JSONSerializer[TestData]{})

	_, err := compCache.Get("invalid")

	if err == nil {
		t.Error("Expected error for invalid compressed data, got nil")
	}
}

func TestCompressionDecorator_MultipleItems(t *testing.T) {
	baseCache := strategies.NewLRUCache[string, []byte](10)
	compCache := WithCompression(baseCache, JSONSerializer[TestData]{})

	items := map[string]TestData{
		"item1": {ID: 1, Name: "First", Tags: []string{"a"}},
		"item2": {ID: 2, Name: "Second", Tags: []string{"b", "c"}},
		"item3": {ID: 3, Name: "Third", Tags: []string{"d", "e", "f"}},
	}

	for key, data := range items {
		if err := compCache.Set(key, data); err != nil {
			t.Fatalf("Set failed for key %s: %v", key, err)
		}
	}

	for key, expected := range items {
		retrieved, err := compCache.Get(key)
		if err != nil {
			t.Fatalf("Get failed for key %s: %v", key, err)
		}
		if retrieved.ID != expected.ID {
			t.Errorf("Key %s: ID mismatch: got %d, want %d", key, retrieved.ID, expected.ID)
		}
		if retrieved.Name != expected.Name {
			t.Errorf("Key %s: Name mismatch: got %s, want %s", key, retrieved.Name, expected.Name)
		}
	}
}
