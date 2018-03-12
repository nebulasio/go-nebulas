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

	"github.com/nebulasio/go-nebulas/util/logging"

	"github.com/nebulasio/go-nebulas/storage"
)

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
	isDirtyDB                  bool
	isPreparedDBClosed         bool
	preparedDBs                map[interface{}]*MVCCDB
	isTrieSameKeyCompatibility bool // The `isTrieSameKeyCompatibility` is used to prevent conflict in continuous changes with same key/value.
}

// NewMVCCDB create and return new MVCCDB. The `trieSameKeyCompatibility` is used to prevent conflict in continuous changes with same key/value.
func NewMVCCDB(storage storage.Storage, trieSameKeyCompatibility bool) (*MVCCDB, error) {
	db := &MVCCDB{
		tid:                        nil,
		storage:                    storage,
		stagingTable:               nil,
		parentDB:                   nil,
		isInTransaction:            false,
		isPreparedDB:               false,
		isDirtyDB:                  false,
		isPreparedDBClosed:         false,
		preparedDBs:                make(map[interface{}]*MVCCDB),
		isTrieSameKeyCompatibility: trieSameKeyCompatibility,
	}

	db.tid = storage // as a placeholder.
	db.stagingTable = NewStagingTable(storage, db.tid, trieSameKeyCompatibility)

	return db, nil
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

	if db.IsPreparedDBDirty() {
		return ErrPreparedDBIsDirty
	}

	var retErr error

	// commit.
	db.stagingTable.Lock()
	logging.CLog().Info("MVCCDB Commit ", len(db.stagingTable.GetVersionizedValues()))

	// enable batch.
	db.storage.EnableBatch()

	var err error
	for _, value := range db.stagingTable.GetVersionizedValues() {
		// skip default value loaded from storage.
		if value.isDefault() {
			continue
		}

		if !value.dirty {
			continue
		}

		if value.deleted {
			err = db.delFromStorage(value.key)
		} else {
			err = db.putToStorage(value.key, value.val)
		}

		if err != nil && retErr == nil {
			retErr = err
		}

		// logging.CLog().Infof("MVCCDB.COMMIT: %s %s %d", byteutils.Hex(value.key), byteutils.Hex(hash.Sha3256(value.val)), value.version)
	}

	// flush and disable batch.
	err = db.storage.Flush()
	if err != nil && retErr == nil {
		retErr = err
	}

	db.storage.DisableBatch()

	// unlock.
	db.stagingTable.Unlock()

	if retErr != nil {
		return retErr
	}

	// reset.
	for _, pdb := range db.preparedDBs {
		pdb.Reset()
	}
	db.preparedDBs = make(map[interface{}]*MVCCDB)

	// done.
	db.isInTransaction = false
	db.isDirtyDB = false

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

	// reset.
	for _, pdb := range db.preparedDBs {
		pdb.Reset()
	}
	db.preparedDBs = make(map[interface{}]*MVCCDB)

	db.stagingTable.Purge()

	// done.
	db.isInTransaction = false
	db.isDirtyDB = false

	return nil
}

// Get value
func (db *MVCCDB) Get(key []byte) ([]byte, error) {
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
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if db.isPreparedDB && db.isPreparedDBClosed {
		return ErrPreparedDBIsClosed
	}

	if !db.isInTransaction {
		return db.putToStorage(key, val)
	}

	_, err := db.stagingTable.Put(key, val)
	if err == nil {
		db.isDirtyDB = true
	}
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
	if err == nil {
		db.isDirtyDB = true
	}
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
		isDirtyDB:                  false,
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
		db.isDirtyDB = false

		// cleanup.
		db.stagingTable.Purge()
	}

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

	db.isDirtyDB = false

	return nil
}

// Close close prepared DB.
func (db *MVCCDB) Close() error {

	if !db.isInTransaction {
		return ErrTransactionNotStarted
	}

	if !db.isPreparedDB {
		return ErrDisallowedCallingInNoPreparedDB
	}

	if db.isPreparedDBClosed {
		return ErrPreparedDBIsClosed
	}

	err := db.stagingTable.Close()
	if err != nil {
		return err
	}

	db.parentDB.mutex.Lock()
	defer db.parentDB.mutex.Unlock()

	delete(db.parentDB.preparedDBs, db.tid)
	db.isPreparedDBClosed = true

	return nil
}

// IsPreparedDBDirty is prepared db dirty
func (db *MVCCDB) IsPreparedDBDirty() bool {
	for _, pdb := range db.preparedDBs {
		if pdb.IsPreparedDBDirty() {
			return true
		}
	}

	return false
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
