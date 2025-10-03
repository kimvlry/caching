package metrics

type Collector interface {
	RecordHit()
	RecordMiss()
	RecordEviction()
	RecordCompression(uncompressedSize int, compressedSize int)

	GetHits() int
	GetMisses() int
	GetEvictions() int
	CompressionRatio() float64
}
