package decorators

import (
	"bytes"
	"caching-labwork/cache"
	"caching-labwork/cache/cache_decorators/metrics"
	"compress/gzip"
	"encoding/json"
	"io"
)

type CompressionDecorator[K comparable, V any] struct {
	cacheWrappee cache.Cache[K, []byte]
	collector    metrics.Collector
}

func WithCompression[K comparable, V any](wrappee cache.Cache[K, []byte]) *CompressionDecorator[K, V] {
	return &CompressionDecorator[K, V]{
		cacheWrappee: wrappee,
	}
}

func (w *CompressionDecorator[K, V]) Get(key K) (V, error) {
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

func (w *CompressionDecorator[K, V]) Set(key K, value V) error {
	rawBytes, err := json.Marshal(value)
	if err != nil {
		return err
	}
	compressedBytes, err := compressRaw(rawBytes)
	if err != nil {
		return err
	}

	w.collector.RecordCompression(len(rawBytes), len(compressedBytes))
	return w.cacheWrappee.Set(key, compressedBytes)
}

func (w *CompressionDecorator[K, V]) Delete(key K) error {
	return w.cacheWrappee.Delete(key)
}

func (w *CompressionDecorator[K, V]) Clear() {
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
