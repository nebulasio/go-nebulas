package storage

import (
	"github.com/nebulasio/go-nebulas/util"
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

	opts *gorocksdb.Options
	columnHandles  map[string]*gorocksdb.ColumnFamilyHandle

	ro *gorocksdb.ReadOptions
	wo *gorocksdb.WriteOptions

	cache *gorocksdb.Cache
}

func createRocksOptions(cache *gorocksdb.Cache) *gorocksdb.Options {
	filter := gorocksdb.NewBloomFilter(10)
	bbto := gorocksdb.NewDefaultBlockBasedTableOptions()
	bbto.SetFilterPolicy(filter)

	bbto.SetBlockCache(cache)
	opts := gorocksdb.NewDefaultOptions()
	opts.SetBlockBasedTableFactory(bbto)
	opts.SetCreateIfMissing(true)
	opts.SetMaxOpenFiles(500)
	opts.SetWriteBufferSize(64 * opt.MiB) //Default: 4MB
	opts.IncreaseParallelism(4)           //flush and compaction thread

	opts.SetKeepLogFileNum(1)
	opts.SetDbLogDir("logs")
	return opts
}

// NewRocksStorage init a storage
func NewRocksStorage(path string) (*RocksStorage, error) {

	cache := gorocksdb.NewLRUCache(512 << 20)
	opts := createRocksOptions(cache)
	db, err := gorocksdb.OpenDb(opts, path)
	if err != nil {
		return nil, err
	}

	storage := &RocksStorage{
		db:          db,
		cache:       cache,
		enableBatch: false,
		opts:		 opts,
		batchOpts:   make(map[string]*batchOpt),
		columnHandles: make(map[string]*gorocksdb.ColumnFamilyHandle),
		ro:          gorocksdb.NewDefaultReadOptions(),
		wo:          gorocksdb.NewDefaultWriteOptions(),
	}

	//go RecordMetrics(storage)

	return storage, nil
}

// NewRocksStorage init a storage with column families
func NewRocksStorageWithCF(path string, cfNames []string) (*RocksStorage, error) {
	cache := gorocksdb.NewLRUCache(512 << 20)
	opts := createRocksOptions(cache)

	//we should create column families first.
	//as the `default` column is can't be find for the normal open, we should close and then reopen.
	// This is strange, but the example is like this.
	if exist,_ := util.FileExists(path); !exist {
		db, err := gorocksdb.OpenDb(opts, path)
		if err != nil {
			return nil, err
		}
		for _, cfName := range cfNames {
			_, err := db.CreateColumnFamily(opts, cfName)
			if err != nil {
				return nil, err
			}
		}
		db.Close()
	}

	cfNames = append([]string{"default"}, cfNames...)
	cfOpts := make([]*gorocksdb.Options, len(cfNames))
	for idx := range cfNames {
		cfOpts[idx] = opts
	}
	db, handles, err := gorocksdb.OpenDbColumnFamilies(opts, path, cfNames, cfOpts)
	if err != nil {
		return nil, err
	}

	handlesMap := make(map[string]*gorocksdb.ColumnFamilyHandle)
	for idx, handle := range handles {
		cfName := cfNames[idx]
		handlesMap[cfName] = handle
	}

	//cfList, err := gorocksdb.ListColumnFamilies(opts, path)
	//
	//if err != nil {
	//	return nil, err
	//}
	//if len(cfList) != len(cfNames) {
	//	return nil, err
	//}

	storage := &RocksStorage{
		db:          db,
		cache:       cache,
		enableBatch: false,
		opts:		 opts,
		batchOpts:   make(map[string]*batchOpt),
		columnHandles: handlesMap,
		ro:          gorocksdb.NewDefaultReadOptions(),
		wo:          gorocksdb.NewDefaultWriteOptions(),
	}

	//go RecordMetrics(storage)

	return storage, nil
}

func (storage *RocksStorage) CreateColumn(name string) error {
	handle, err := storage.db.CreateColumnFamily(storage.opts, name)
	if err != nil {
		return err
	}
	storage.columnHandles[name] = handle
	return nil
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

// GetCF return value to the key in Storage
func (storage *RocksStorage) GetCF(cfName string, key []byte) ([]byte, error) {
	if handle, ok := storage.columnHandles[cfName]; ok {
		value, err := storage.db.GetCF(storage.ro, handle, key)

		if err != nil {
			return nil, err
		}

		if value == nil {
			return nil, ErrKeyNotFound
		}
		defer value.Free()
		return value.Data(), err
	} else {
		return nil, ErrCFNameNotFound
	}
}

// PutCF put the key-value entry to Storage
func (storage *RocksStorage) PutCF(cfName string, key []byte, value []byte) error {
	if handle, ok := storage.columnHandles[cfName]; ok {
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

		return storage.db.PutCF(storage.wo, handle, key, value)
	} else {
		return ErrCFNameNotFound
	}
}

// DelCF delete the key in Storage.
func (storage *RocksStorage) DelCF(cfName string, key []byte) error {
	if handle, ok := storage.columnHandles[cfName]; ok {
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
		return storage.db.DeleteCF(storage.wo, handle, key)
	} else {
		return ErrCFNameNotFound
	}
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
		if len(opt.cfName) > 0 {
			handle := storage.columnHandles[opt.cfName]
			if opt.deleted {
				wb.DeleteCF(handle, opt.key)
			} else {
				wb.PutCF(handle, opt.key, opt.value)
			}
		} else {
			if opt.deleted {
				wb.Delete(opt.key)
			} else {
				wb.Put(opt.key, opt.value)
			}
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
