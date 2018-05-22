package cache

import "errors"

// Operator replay operator
type Operator int

// const
const (
	Add Operator = 1 + iota
	Delete
)

// var
var (
	ErrInvalidEncodeArguments = errors.New("invalid encode arguments")
	ErrInvalidCacheKey        = errors.New("invalid cache key")
	ErrInvalidCacheValue      = errors.New("invalid cache value")
)

// ExportableEntry ..
type ExportableEntry struct {
	K interface{}
	V interface{}
}

// ReplayLog ..
type ReplayLog struct {
	Op      Operator
	Operand *ExportableEntry
}

// Cache cache interface definition
// cache key must be string-compatible or comparable type
type Cache interface {
	Entries() interface{}
	Keys() interface{}
	Set(interface{}, interface{}) error
	Get(interface{}) interface{}
	Take(interface{}) (interface{}, error)
	Size() int
	EncodeEntry(interface{}, interface{}) (*ExportableEntry, error)
	DecodeEntry(*ExportableEntry) (interface{}, interface{}, error)
}

// PersistableCacheConfig config of cache
type PersistableCacheConfig struct {
	// dump interval in milliseconds
	IntervalMs int64
	// storage apth
	SaveDir string
	// cache size
	CacheSize int
}
