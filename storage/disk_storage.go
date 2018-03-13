// Copyright (C) 2017 go-nebulas authors
//
// This file is part of the go-nebulas library.
//
// the go-nebulas library is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// the go-nebulas library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-nebulas library.  If not, see <http://www.gnu.org/licenses/>.
//

package storage

import (
	"sync"

	"github.com/nebulasio/go-nebulas/util/byteutils"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

// DiskStorage the nodes in trie.
type DiskStorage struct {
	db          *leveldb.DB
	enableBatch bool
	mutex       sync.Mutex
	batchOpts   map[string]*batchOpt
}

type batchOpt struct {
	key     []byte
	value   []byte
	deleted bool
}

// NewDiskStorage init a storage
func NewDiskStorage(path string) (*DiskStorage, error) {
	db, err := leveldb.OpenFile(path, &opt.Options{
		OpenFilesCacheCapacity: 500,
		BlockCacheCapacity:     8 * opt.MiB,
		BlockSize:              4 * opt.MiB,
		Filter:                 filter.NewBloomFilter(10),
	})
	if err != nil {
		return nil, err
	}
	return &DiskStorage{
		db:          db,
		enableBatch: false,
		batchOpts:   make(map[string]*batchOpt),
	}, nil
}

// Get return value to the key in Storage
func (storage *DiskStorage) Get(key []byte) ([]byte, error) {
	// if storage.enableBatch {
	// 	storage.mutex.Lock()
	// 	defer storage.mutex.Unlock()

	// 	opt := storage.batchOpts[byteutils.Hex(key)]
	// 	if opt != nil {
	// 		if opt.deleted {
	// 			return nil, ErrKeyNotFound
	// 		}
	// 		return opt.value, nil
	// 	}
	// }

	value, err := storage.db.Get(key, nil)
	if err != nil && err == leveldb.ErrNotFound {
		return nil, ErrKeyNotFound
	}

	return value, err
}

// Put put the key-value entry to Storage
func (storage *DiskStorage) Put(key []byte, value []byte) error {
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

	return storage.db.Put(key, value, nil)
}

// Del delete the key in Storage.
func (storage *DiskStorage) Del(key []byte) error {
	if storage.enableBatch {
		storage.mutex.Lock()
		defer storage.mutex.Unlock()

		storage.batchOpts[byteutils.Hex(key)] = &batchOpt{
			key:     key,
			deleted: true,
		}

		return nil
	}

	return storage.db.Delete(key, nil)
}

// Close levelDB
func (storage *DiskStorage) Close() error {
	return storage.db.Close()
}

// EnableBatch enable batch write.
func (storage *DiskStorage) EnableBatch() {
	storage.enableBatch = true
}

// Flush write and flush pending batch write.
func (storage *DiskStorage) Flush() error {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()

	batch := new(leveldb.Batch)
	for _, opt := range storage.batchOpts {
		if opt.deleted {
			batch.Delete(opt.key)
		} else {
			batch.Put(opt.key, opt.value)
		}
	}
	storage.batchOpts = make(map[string]*batchOpt)

	return storage.db.Write(batch, nil)
}

// DisableBatch disable batch write.
func (storage *DiskStorage) DisableBatch() {
	storage.Flush()
	storage.enableBatch = false
}
