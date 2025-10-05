package decorators

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"github.com/kimvlry/caching/cache"
	"io"
)

type Serializer[V any] interface {
	Marshal(V) ([]byte, error)
	Unmarshal([]byte) (V, error)
}

type compressionDecorator[K comparable, V any] struct {
	cacheWrappee   cache.Cache[K, []byte]
	serializerWrap Serializer[V]
	eventCallbacks []func(cache.Event[K, V])
}

func WithCompression[K comparable, V any](
	wrappee cache.Cache[K, []byte],
	serializer Serializer[V],
) cache.ObservableCache[K, V] {

	return &compressionDecorator[K, V]{
		cacheWrappee:   wrappee,
		serializerWrap: serializer,
	}
}

func (w *compressionDecorator[K, V]) Get(key K) (V, error) {
	compressed, err := w.cacheWrappee.Get(key)
	if err != nil {
		var zero V
		return zero, err
	}

	raw, err := decompressRaw(compressed)
	if err != nil {
		var zero V
		return zero, err
	}

	var v V
	if err := json.Unmarshal(raw, &v); err != nil {
		var zero V
		return zero, err
	}
	return v, nil
}

func (w *compressionDecorator[K, V]) Set(key K, value V) error {
	rawBytes, err := json.Marshal(value)
	if err != nil {
		return err
	}
	w.emit(cache.Event[K, V]{
		Type: cache.EventTypeReadBytes,
		Key:  key,
		Size: len(rawBytes),
	})

	compressedBytes, err := compressRaw(rawBytes)
	if err != nil {
		return err
	}
	w.emit(cache.Event[K, V]{
		Type: cache.EventTypeCompressBytes,
		Key:  key,
		Size: len(compressedBytes),
	})

	return w.cacheWrappee.Set(key, compressedBytes)
}

func (w *compressionDecorator[K, V]) Delete(key K) error {
	return w.cacheWrappee.Delete(key)
}

func (w *compressionDecorator[K, V]) Clear() {
	w.cacheWrappee.Clear()
}

func compressRaw(raw []byte) ([]byte, error) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, err := gw.Write(raw)
	if err != nil {
		return nil, err
	}
	if err := gw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func decompressRaw(data []byte) (b []byte, err error) {
	gr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer func() {
		err = gr.Close()
	}()
	return io.ReadAll(gr)
}

func (w *compressionDecorator[K, V]) OnEvent(callback func(cache.Event[K, V])) {
	w.eventCallbacks = append(w.eventCallbacks, callback)
}

func (w *compressionDecorator[K, V]) emit(event cache.Event[K, V]) {
	for _, callback := range w.eventCallbacks {
		callback(event)
	}
}
