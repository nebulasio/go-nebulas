package storage

import (
	"strconv"
	"sync"
	"time"

	"github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/tecbot/gorocksdb"
)

// RocksStorage the nodes in trie.
type RocksStorage struct {
	db          *gorocksdb.DB
	enableBatch bool
	mutex       sync.Mutex
	batchOpts   map[string]*batchOpt

	ro            *gorocksdb.ReadOptions
	wo            *gorocksdb.WriteOptions
	cfHandlersMap map[string]*gorocksdb.ColumnFamilyHandle

	cache *gorocksdb.Cache
}

// NewRocksStorage init a storage
func NewRocksStorage(path string) (*RocksStorage, error) {

	filter := gorocksdb.NewBloomFilter(10)
	bbto := gorocksdb.NewDefaultBlockBasedTableOptions()
	bbto.SetFilterPolicy(filter)

	cache := gorocksdb.NewLRUCache(512 << 20)
	bbto.SetBlockCache(cache)
	opts := gorocksdb.NewDefaultOptions()
	opts.SetBlockBasedTableFactory(bbto)
	opts.SetCreateIfMissing(true)
	opts.SetMaxOpenFiles(500)
	opts.SetWriteBufferSize(64 * opt.MiB) //Default: 4MB
	opts.IncreaseParallelism(4)           //flush and compaction thread
	opts.SetKeepLogFileNum(1)

	db, err := gorocksdb.OpenDb(opts, path)
	if err != nil {
		return nil, err
	}

	storage := &RocksStorage{
		db:          db,
		cache:       cache,
		enableBatch: false,
		batchOpts:   make(map[string]*batchOpt),
		ro:          gorocksdb.NewDefaultReadOptions(),
		wo:          gorocksdb.NewDefaultWriteOptions(),
	}

	//go RecordMetrics(storage)

	return storage, nil
}

// NewRocksStorage init a storage with column Families
func NewRocksStorageWithCF(path string, columnFamiliesNames []string) (*RocksStorage, error) {

	filter := gorocksdb.NewBloomFilter(10)
	bbto := gorocksdb.NewDefaultBlockBasedTableOptions()
	bbto.SetFilterPolicy(filter)

	cache := gorocksdb.NewLRUCache(512 << 20)
	bbto.SetBlockCache(cache)
	opts := gorocksdb.NewDefaultOptions()
	opts.SetBlockBasedTableFactory(bbto)
	opts.SetCreateIfMissing(true)
	opts.SetCreateIfMissingColumnFamilies(true)
	opts.SetMaxOpenFiles(500)
	opts.SetWriteBufferSize(64 * opt.MiB) //Default: 4MB
	opts.IncreaseParallelism(4)           //flush and compaction thread
	opts.SetKeepLogFileNum(1)

	optArray := make([]*gorocksdb.Options, len(columnFamiliesNames))
	for i := 0; i < len(columnFamiliesNames); i++ {
		optArray[i] = opts
	}
	db, cfh, err := gorocksdb.OpenDbColumnFamilies(opts, path, columnFamiliesNames, optArray)
	if err != nil {
		return nil, err
	}

	storage := &RocksStorage{
		db:            db,
		cache:         cache,
		enableBatch:   false,
		batchOpts:     make(map[string]*batchOpt),
		ro:            gorocksdb.NewDefaultReadOptions(),
		wo:            gorocksdb.NewDefaultWriteOptions(),
		cfHandlersMap: make(map[string]*gorocksdb.ColumnFamilyHandle),
	}
	for i := 0; i < len(columnFamiliesNames); i++ {
		storage.cfHandlersMap[columnFamiliesNames[i]] = cfh[i]
	}

	return storage, nil
}

// Get return value to the key in Storage with column Families
func (storage *RocksStorage) GetCF(cfName string, key []byte) ([]byte, error) {
	if handler, ok := storage.cfHandlersMap[cfName]; ok {
		value, err := storage.db.GetCF(storage.ro, handler, key)
		defer value.Free()

		if err != nil {
			return nil, err
		}

		if value == nil {
			return nil, ErrKeyNotFound
		}

		dst := value.Data()
		data := make([]byte, len(dst))
		copy(data, dst)
		return data, err
	} else {
		return nil, ErrCFNotFound
	}
}

// Put put the key-value entry to Storage with column Families
func (storage *RocksStorage) PutCF(cfName string, key []byte, value []byte) error {
	if handler, ok := storage.cfHandlersMap[cfName]; ok {
		if storage.enableBatch {
			storage.mutex.Lock()
			defer storage.mutex.Unlock()

			storage.batchOpts[byteutils.Hex(key)] = &batchOpt{
				cfName:  cfName,
				key:     key,
				value:   value,
				deleted: false,
			}

			return nil
		}

		return storage.db.PutCF(storage.wo, handler, key, value)
	} else {
		return ErrCFNotFound
	}
}

// Del delete the key in Storage with column Families
func (storage *RocksStorage) DelCF(cfName string, key []byte) error {
	if handler, ok := storage.cfHandlersMap[cfName]; ok {
		if storage.enableBatch {
			storage.mutex.Lock()
			defer storage.mutex.Unlock()

			storage.batchOpts[byteutils.Hex(key)] = &batchOpt{
				cfName:  cfName,
				key:     key,
				deleted: true,
			}

			return nil
		}
		return storage.db.DeleteCF(storage.wo, handler, key)
	} else {
		return ErrCFNotFound
	}
}

// Get return value to the key in Storage
func (storage *RocksStorage) Get(key []byte) ([]byte, error) {

	value, err := storage.db.GetBytes(storage.ro, key)

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

	return storage.db.Put(storage.wo, key, value)
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
	return storage.db.Delete(storage.wo, key)
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

	if !storage.enableBatch {
		return nil
	}

	startAt := time.Now().UnixNano()

	wb := gorocksdb.NewWriteBatch()
	defer wb.Destroy()

	bl := len(storage.batchOpts)

	for _, opt := range storage.batchOpts {
		if opt.deleted {
			wb.Delete(opt.key)
		} else {
			wb.Put(opt.key, opt.value)
		}
	}
	storage.batchOpts = make(map[string]*batchOpt)

	err := storage.db.Write(storage.wo, wb)

	endAt := time.Now().UnixNano()
	metricsRocksdbFlushTime.Update(endAt - startAt)
	metricsRocksdbFlushLen.Update(int64(bl))

	return err
}

// DisableBatch disable batch write.
func (storage *RocksStorage) DisableBatch() {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()
	storage.batchOpts = make(map[string]*batchOpt)

	storage.enableBatch = false
}

// RecordMetrics record rocksdb metrics
func RecordMetrics(storage *RocksStorage) {
	metricsUpdateChan := time.NewTicker(5 * time.Second).C

	for {
		select {
		case <-metricsUpdateChan:

			readersMemStr := storage.db.GetProperty("rocksdb.estimate-table-readers-mem")
			allMemTablesStr := storage.db.GetProperty("rocksdb.cur-size-all-mem-tables")
			cacheSize := storage.cache.GetUsage()
			pinnedSize := storage.cache.GetPinnedUsage()

			readersMem, err := strconv.Atoi(readersMemStr)
			if err != nil {
				break
			}
			allMemTables, err := strconv.Atoi(allMemTablesStr)
			if err != nil {
				break
			}

			metricsBlocksdbAllMemTables.Update(int64(allMemTables))
			metricsBlocksdbTableReaderMem.Update(int64(readersMem))
			metricsBlocksdbCacheSize.Update(int64(cacheSize))
			metricsBlocksdbCachePinnedSize.Update(int64(pinnedSize))
		}
	}
}
