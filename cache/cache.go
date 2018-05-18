package cache

import "errors"

// Operator ..
type Operator int

// const
const (
	// dump
	Dump Operator = 1 + iota

	// inclog
	Add
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

// SerializableLog ..
type SerializableLog struct {
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
	Encode(interface{}, interface{}) (*ExportableEntry, error)
	Decode(*ExportableEntry) (interface{}, interface{}, error)
}

// PersistableCacheConfig config of cache
type PersistableCacheConfig struct {
	// milliseconds
	IntervalMs int64
	SaveDir    string
	CacheSize  int
}
