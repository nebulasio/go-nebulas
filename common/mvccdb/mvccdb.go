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
	"github.com/nebulasio/go-nebulas/storage"
)

type value struct {
	content []byte
	old     int32
	new     int32
	flag    bool
}

type transaction struct {
	logs map[string]*value
}

// MVCCDB schema
type MVCCDB struct {
	// txid - (key - value)
	transactions map[string]*transaction
	// txid - flag (valid or not)
	status map[string]bool

	storage storage.Storage
}

// DB schema
type DB struct {
	txid   string
	mvccdb *MVCCDB
}

// NewMVCCDB create a new change log
func NewMVCCDB(storage storage.Storage) (*MVCCDB, error) {
	return &MVCCDB{
		transactions: make(map[string]*transaction),
		status:       make(map[string]bool),
		storage:      storage,
	}, nil
}

// Begin a transaction
func (cl *MVCCDB) Begin() error { return nil }

// Commit the transaction to storage
func (cl *MVCCDB) Commit() error { return nil }

// RollBack the transaction
func (cl *MVCCDB) RollBack() error { return nil }

// Prepare a nested transaction
func (cl *MVCCDB) Prepare(txid string) (*DB, error) { return nil, nil }

// Update the nested transaction
func (cl *MVCCDB) Update(txid string) error { return nil }

// Check whether the nested transaction conflicts
func (cl *MVCCDB) Check(txid string) (bool, error) { return false, nil }

// Get value
func (db *DB) Get(key []byte) ([]byte, error) { return nil, nil }

// Put value
func (db *DB) Put(key []byte, val []byte) error { return nil }

// Del value
func (db *DB) Del(key []byte) error { return nil }
