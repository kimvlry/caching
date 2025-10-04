package metrics

type DefaultCollector struct {
	hits            int
	misses          int
	evictions       int
	uncompressedSum int
	compressedSum   int
}

func (c *DefaultCollector) GetHits() int {
	return c.hits
}

func (c *DefaultCollector) GetMisses() int {
	return c.misses
}

func (c *DefaultCollector) GetEvictions() int {
	return c.evictions
}

func (c *DefaultCollector) RecordHit() {
	c.hits++
}

func (c *DefaultCollector) RecordMiss() {
	c.misses++
}

func (c *DefaultCollector) RecordEviction() {
	c.evictions++
}

func (c *DefaultCollector) RecordCompression(uncompressed, compressed int) {
	c.uncompressedSum += uncompressed
	c.compressedSum += compressed
}

func (c *DefaultCollector) CompressionRatio() float64 {
	if c.compressedSum == 0 {
		return 1.0
	}
	return float64(c.uncompressedSum) / float64(c.compressedSum)
}
