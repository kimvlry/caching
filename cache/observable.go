package cache

type ObservableCache[K comparable, V any] interface {
	OnEvent(callback func(event Event[K, V]))
}

type EventType string

const (
	EventTypeHit      EventType = "hit"
	EventTypeMiss     EventType = "miss"
	EventTypeEviction EventType = "eviction"

	EventTypeReadBytes     EventType = "write raw bytes"
	EventTypeCompressBytes EventType = "compress bytes"
)

type Event[K comparable, V any] struct {
	Type  EventType
	Key   K
	Value V
	Size  int
}
