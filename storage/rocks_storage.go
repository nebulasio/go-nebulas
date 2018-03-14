package storage

import (
	"sync"

	"github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/tecbot/gorocksdb"
)

/*
CGO_CFLAGS="-I/Users/fengzi/go/src/github.com/nebulasio/go-nebulas/storage/" \
CGO_LDFLAGS="-L/Users/fengzi/go/src/github.com/nebulasio/go-nebulas/storage/rocksdb -lrocksdb -lstdc++ -lm -lz -lbz2 -lsnappy -llz4 -lzstd" \
  go get github.com/tecbot/gorocksdb
*/
// RocksStorage the nodes in trie.
type RocksStorage struct {
	db          *gorocksdb.DB
	enableBatch bool
	mutex       sync.Mutex
	batchOpts   map[string]*batchOpt
}

// NewRocksStorage init a storage
func NewRocksStorage(path string) (*RocksStorage, error) {

	filter := gorocksdb.NewBloomFilter(10)
	bbto := gorocksdb.NewDefaultBlockBasedTableOptions()
	bbto.SetFilterPolicy(filter)
	bbto.SetBlockCache(gorocksdb.NewLRUCache(3 << 30))
	opts := gorocksdb.NewDefaultOptions()
	opts.SetBlockBasedTableFactory(bbto)
	opts.SetCreateIfMissing(true)

	db, err := gorocksdb.OpenDb(opts, path)
	if err != nil {
		return nil, err
	}

	return &RocksStorage{
		db:          db,
		enableBatch: false,
		batchOpts:   make(map[string]*batchOpt),
	}, nil
}

// Get return value to the key in Storage
func (storage *RocksStorage) Get(key []byte) ([]byte, error) {
	ro := gorocksdb.NewDefaultReadOptions()
	value, err := storage.db.GetBytes(ro, key)

	if err != nil {
		return nil, err
	}

	if value == nil {
		return nil, ErrKeyNotFound
	}

	return value, err
}

// Put put the key-value entry to Storage
func (storage *RocksStorage) Put(key []byte, value []byte) error {
	if storage.enableBatch {
		storage.mutex.Lock()
		defer storage.mutex.Unlock()

		storage.batchOpts[byteutils.Hex(key)] = &batchOpt{
			key:     key,
			value:   value,
			deleted: false,
		}

		return nil
	}
	wo := gorocksdb.NewDefaultWriteOptions()
	return storage.db.Put(wo, key, value)
}

// Del delete the key in Storage.
func (storage *RocksStorage) Del(key []byte) error {
	if storage.enableBatch {
		storage.mutex.Lock()
		defer storage.mutex.Unlock()

		storage.batchOpts[byteutils.Hex(key)] = &batchOpt{
			key:     key,
			deleted: true,
		}

		return nil
	}
	wo := gorocksdb.NewDefaultWriteOptions()
	return storage.db.Delete(wo, key)
}

// Close levelDB
func (storage *RocksStorage) Close() error {
	storage.db.Close()
	return nil
}

// EnableBatch enable batch write.
func (storage *RocksStorage) EnableBatch() {
	storage.enableBatch = true
}

// Flush write and flush pending batch write.
func (storage *RocksStorage) Flush() error {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()

	wo := gorocksdb.NewDefaultWriteOptions()

	wb := gorocksdb.NewWriteBatch()
	defer wb.Clear()
	for _, opt := range storage.batchOpts {
		if opt.deleted {
			wb.Delete(opt.key)
		} else {
			wb.Put(opt.key, opt.value)
		}
	}

	return storage.db.Write(wo, wb)
}

// DisableBatch disable batch write.
func (storage *RocksStorage) DisableBatch() {
	storage.Flush()
	storage.enableBatch = false
}
