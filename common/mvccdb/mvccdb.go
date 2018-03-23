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

package mvccdb

import (
	"errors"
	"sync"

	"github.com/nebulasio/go-nebulas/metrics"
	"github.com/nebulasio/go-nebulas/storage"
)

// Errors
var (
	ErrUnsupportedNestedTransaction    = errors.New("unsupported nested transaction")
	ErrTransactionNotStarted           = errors.New("transaction is not started")
	ErrDisallowedCallingInNoPreparedDB = errors.New("disallowed calling in No-Prepared MVCCDB")
	ErrTidIsNil                        = errors.New("tid is nil")
	ErrUnsupportedBeginInPreparedDB    = errors.New("unsupported begin transaction in prepared MVCCDB")
	ErrUnsupportedCommitInPreparedDB   = errors.New("unsupported commit transaction in prepared MVCCDB")
	ErrUnsupportedRollBackInPreparedDB = errors.New("unsupported rollback transaction in prepared MVCCDB")
	ErrPreparedDBIsDirty               = errors.New("prepared MVCCDB is dirty")
	ErrPreparedDBIsClosed              = errors.New("prepared MVCCDB is closed")
	ErrTidIsExist                      = errors.New("tid is exist")
)

func init() {
	metrics.EnableMetrics()
}

/* How to use MVCCDB
It should support three situations as following,
1. directly Get/Put/Del.
2. begin - Get/Put/Del - commit/rollback.
3. begin - prepare - Get/Put/Del - update - commit/rollback
*/

// MVCCDB the data with MVCC supporting.
type MVCCDB struct {
	tid                        interface{}
	storage                    storage.Storage
	stagingTable               *StagingTable
	mutex                      sync.Mutex
	parentDB                   *MVCCDB
	isInTransaction            bool
	isPreparedDB               bool
	isPreparedDBClosed         bool
	preparedDBs                map[interface{}]*MVCCDB
	isTrieSameKeyCompatibility bool // The `isTrieSameKeyCompatibility` is used to prevent conflict in continuous changes with same key/value.
}

// NewMVCCDB create and return new MVCCDB. The `trieSameKeyCompatibility` is used to prevent conflict in continuous changes with same key/value.
func NewMVCCDB(storage storage.Storage, trieSameKeyCompatibility bool) (*MVCCDB, error) {
	tid := new(int)

	db := &MVCCDB{
		tid:                        tid,
		storage:                    storage,
		stagingTable:               NewStagingTable(storage, tid, trieSameKeyCompatibility),
		parentDB:                   nil,
		isInTransaction:            false,
		isPreparedDB:               false,
		isPreparedDBClosed:         false,
		preparedDBs:                make(map[interface{}]*MVCCDB),
		isTrieSameKeyCompatibility: trieSameKeyCompatibility,
	}

	return db, nil
}

// SetStrictGlobalVersionCheck set strict global version check flag.
func (db *MVCCDB) SetStrictGlobalVersionCheck(flag bool) {
	db.stagingTable.SetStrictGlobalVersionCheck(flag)
}

// Begin begin a transaction.
func (db *MVCCDB) Begin() error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if db.isInTransaction {
		return ErrUnsupportedNestedTransaction
	}

	if db.isPreparedDB {
		return ErrUnsupportedBeginInPreparedDB
	}

	db.isInTransaction = true

	return nil
}

// Commit commit changes to storage.
func (db *MVCCDB) Commit() error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if !db.isInTransaction {
		return ErrTransactionNotStarted
	}

	if db.isPreparedDB {
		return ErrUnsupportedCommitInPreparedDB
	}

	// commit.
	db.stagingTable.Lock()

	// enable batch.
	db.storage.EnableBatch()

	var err error
	for _, value := range db.stagingTable.getVersionizedValues() {
		// skip default value loaded from storage.
		if value.isDefault() {
			continue
		}

		if !value.dirty {
			continue
		}

		if value.deleted {
			// The delete opt of Trie is not `delete`, just flag to delete.
			// So no delete when isTrieSameKeyCompatibility == true.
			if db.isTrieSameKeyCompatibility == false {
				err = db.delFromStorage(value.key)
			}
		} else {
			err = db.putToStorage(value.key, value.val)
		}

		if err != nil {
			db.storage.DisableBatch()
			db.stagingTable.Unlock()
			return err
		}
	}

	// flush and disable batch.
	err = db.storage.Flush()
	if err != nil {
		db.storage.DisableBatch()
		db.stagingTable.Unlock()
		return err
	}

	db.storage.DisableBatch()

	// unlock.
	db.stagingTable.Unlock()

	// Close child prepareDBs.
	for _, pdb := range db.preparedDBs {
		pdb.doClose()
	}
	db.preparedDBs = make(map[interface{}]*MVCCDB)
	db.stagingTable.Purge()

	// done.
	db.isInTransaction = false

	return nil
}

// RollBack the transaction
func (db *MVCCDB) RollBack() error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if !db.isInTransaction {
		return ErrTransactionNotStarted
	}

	if db.isPreparedDB {
		return ErrUnsupportedRollBackInPreparedDB
	}

	// Close.
	for _, pdb := range db.preparedDBs {
		pdb.doClose()
	}
	db.preparedDBs = make(map[interface{}]*MVCCDB)
	db.stagingTable.Purge()

	// done.
	db.isInTransaction = false

	return nil
}

// Get value
func (db *MVCCDB) Get(key []byte) ([]byte, error) {
	// s := time.Now().UnixNano()

	db.mutex.Lock()
	defer db.mutex.Unlock()

	if db.isPreparedDB && db.isPreparedDBClosed {
		return nil, ErrPreparedDBIsClosed
	}

	if !db.isInTransaction {
		return db.getFromStorage(key)
	}

	value, err := db.stagingTable.Get(key)
	if err != nil {
		return nil, err
	}

	if value.deleted || value.val == nil {
		return nil, storage.ErrKeyNotFound
	}

	return value.val, nil
}

// Put value
func (db *MVCCDB) Put(key []byte, val []byte) error {
	// s := time.Now().UnixNano()

	db.mutex.Lock()
	defer db.mutex.Unlock()

	if db.isPreparedDB && db.isPreparedDBClosed {
		return ErrPreparedDBIsClosed
	}

	if !db.isInTransaction {
		return db.putToStorage(key, val)
	}

	_, err := db.stagingTable.Put(key, val)
	return err
}

// Del value
func (db *MVCCDB) Del(key []byte) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if db.isPreparedDB && db.isPreparedDBClosed {
		return ErrPreparedDBIsClosed
	}

	if !db.isInTransaction {
		return db.delFromStorage(key)
	}

	_, err := db.stagingTable.Del(key)
	return err
}

// Prepare a nested transaction
func (db *MVCCDB) Prepare(tid interface{}) (*MVCCDB, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if !db.isInTransaction {
		return nil, ErrTransactionNotStarted
	}

	if tid == nil {
		return nil, ErrTidIsNil
	}

	if db.preparedDBs[tid] != nil {
		return nil, ErrTidIsExist
	}

	preparedStagingTable, err := db.stagingTable.Prepare(tid)
	if err != nil {
		return nil, err
	}

	preparedDB := &MVCCDB{
		tid:                        tid,
		storage:                    db.storage,
		stagingTable:               preparedStagingTable,
		parentDB:                   db,
		isInTransaction:            true,
		isPreparedDB:               true,
		isPreparedDBClosed:         false,
		preparedDBs:                make(map[interface{}]*MVCCDB),
		isTrieSameKeyCompatibility: db.isTrieSameKeyCompatibility,
	}

	db.preparedDBs[tid] = preparedDB
	return preparedDB, nil
}

// CheckAndUpdate merge current changes to `FinalVersionizedValues`.
func (db *MVCCDB) CheckAndUpdate() ([]interface{}, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if !db.isInTransaction {
		return nil, ErrTransactionNotStarted
	}

	if !db.isPreparedDB {
		return nil, ErrDisallowedCallingInNoPreparedDB
	}

	if db.isPreparedDBClosed {
		return nil, ErrPreparedDBIsClosed
	}

	ret, err := db.stagingTable.MergeToParent()

	if err == nil {
		// cleanup.
		db.stagingTable.Purge()
	}

	/* 	logging.CLog().Infof("tid %s-%s: GET latency { %6.6f, %6.6f, %6.6f, %6.6f }; PUT latency { %6.6f, %6.6f, %6.6f, %6.6f }}",
	   		db.prefix,
	   		db.tid,
	   		db.getLatency.Percentile(0.10),
	   		db.getLatency.Percentile(0.50),
	   		db.getLatency.Percentile(0.80),
	   		db.getLatency.Percentile(0.90),
	   		db.putLatency.Percentile(0.10),
	   		db.putLatency.Percentile(0.50),
	   		db.putLatency.Percentile(0.80),
	   		db.putLatency.Percentile(0.90),
	   	)
	*/
	return ret, err
}

// Reset the nested transaction
func (db *MVCCDB) Reset() error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if !db.isInTransaction {
		return ErrTransactionNotStarted
	}

	if !db.isPreparedDB {
		return ErrDisallowedCallingInNoPreparedDB
	}

	if db.isPreparedDBClosed {
		return ErrPreparedDBIsClosed
	}

	// reset.
	for _, pdb := range db.preparedDBs {
		pdb.Reset()
	}
	db.preparedDBs = make(map[interface{}]*MVCCDB)

	db.stagingTable.Purge()

	return nil
}

// Close close prepared DB.
func (db *MVCCDB) Close() error {
	db.parentDB.mutex.Lock()
	defer db.parentDB.mutex.Unlock()

	return db.doClose()
}

func (db *MVCCDB) doClose() error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if !db.isInTransaction {
		return ErrTransactionNotStarted
	}

	if !db.isPreparedDB {
		return ErrDisallowedCallingInNoPreparedDB
	}

	if db.isPreparedDBClosed {
		return ErrPreparedDBIsClosed
	}

	// Close.
	for _, pdb := range db.preparedDBs {
		pdb.doClose()
	}
	db.preparedDBs = make(map[interface{}]*MVCCDB)

	err := db.stagingTable.Close()
	if err != nil {
		return err
	}
	db.stagingTable.Purge()

	// detach from parent.
	delete(db.parentDB.preparedDBs, db.tid)
	db.parentDB = nil
	db.isPreparedDBClosed = true

	return nil
}

// GetParentDB return the root db.
func (db *MVCCDB) GetParentDB() *MVCCDB {
	return db.parentDB
}

func (db *MVCCDB) getFromStorage(key []byte) ([]byte, error) {
	return db.storage.Get(key)
}

func (db *MVCCDB) putToStorage(key []byte, val []byte) error {
	return db.storage.Put(key, val)
}

func (db *MVCCDB) delFromStorage(key []byte) error {
	return db.storage.Del(key)
}

// EnableBatch enable batch write.
func (db *MVCCDB) EnableBatch() {
}

// Flush write and flush pending batch write.
func (db *MVCCDB) Flush() error {
	return nil
}

// DisableBatch disable batch write.
func (db *MVCCDB) DisableBatch() {
}
