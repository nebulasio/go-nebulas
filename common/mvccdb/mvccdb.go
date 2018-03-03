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
)

/* How to use MVCCDB
It should support three situations as following,
1. directly Get/Put/Del.
2. begin - Get/Put/Del - commit/rollback.
3. begin - prepare - Get/Put/Del - update - commit/rollback
*/

// MVCCDB the data with MVCC supporting.
type MVCCDB struct {
	tid             interface{}
	storage         storage.Storage
	stagingTable    *StagingTable
	mutex           sync.Mutex
	rootDB          *MVCCDB
	isInTransaction bool
	isPreparedDB    bool
}

// NewMVCCDB create and return new MVCCDB.
func NewMVCCDB(storage storage.Storage) (*MVCCDB, error) {
	db := &MVCCDB{
		tid:             nil,
		storage:         storage,
		stagingTable:    NewStagingTable(),
		rootDB:          nil,
		isInTransaction: false,
		isPreparedDB:    false,
	}

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

	// commit.
	db.stagingTable.LockFinalVersionValue()
	defer db.stagingTable.UnlockFinalVersionValue()
	for _, value := range db.stagingTable.finalVersionizedValue {
		if !value.initialized {
			continue
		}

		if !value.dirty {
			continue
		}

		if value.deleted {
			db.delFromStorage(value.key)
		} else {
			db.putToStorage(value.key, value.val)
		}
	}

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

	// rollback.
	db.stagingTable.Purge(db.tid)

	// done.
	db.isInTransaction = false

	return nil
}

// Get value
func (db *MVCCDB) Get(key []byte) ([]byte, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if !db.isInTransaction {
		return db.getFromStorage(key)
	}

	value := db.stagingTable.Get(db.tid, key)
	if value == nil {
		// get from storage.
		data, err := db.getFromStorage(key)
		if err != nil {
			return nil, err
		}

		value = db.stagingTable.Set(db.tid, key, data, false, false)
	}

	if value.deleted {
		return nil, storage.ErrKeyNotFound
	}
	return value.val, nil
}

// Put value
func (db *MVCCDB) Put(key []byte, val []byte) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if !db.isInTransaction {
		return db.putToStorage(key, val)
	}

	db.stagingTable.Put(db.tid, key, val)
	return nil
}

// Del value
func (db *MVCCDB) Del(key []byte) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if !db.isInTransaction {
		return db.delFromStorage(key)
	}

	db.stagingTable.Del(db.tid, key)
	return nil
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

	return &MVCCDB{
		tid:             tid,
		storage:         db.storage,
		stagingTable:    db.stagingTable,
		rootDB:          db,
		isInTransaction: true,
		isPreparedDB:    true,
	}, nil
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

	return db.stagingTable.MergeToFinal(db.tid)
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

	db.stagingTable.Purge(db.tid)
	return nil
}

// GetRootDB return the root db.
func (db *MVCCDB) GetRootDB() *MVCCDB {
	return db.rootDB
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
