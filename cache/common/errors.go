package common

import "errors"

// Common errors
var (
	ErrKeyNotFound = errors.New("key not found")
	ErrCacheFull   = errors.New("cache is full")
)
