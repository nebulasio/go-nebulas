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

package changelog

import (
	"github.com/nebulasio/go-nebulas/storage"
)

type value struct {
	content []byte
	version int32
}

type transaction struct {
	logs map[string]*value
}

// ChangeLog schema
type ChangeLog struct {
	// txid - (key - value)
	transactions map[string]*transaction
	// txid - flag (valid or not)
	status map[string]bool

	storage storage.Storage
}

// NewChangeLog create a new change log
func NewChangeLog(storage storage.Storage) (*ChangeLog, error) {
	return &ChangeLog{
		transactions: make(map[string]*transaction),
		status:       make(map[string]bool),
		storage:      storage,
	}, nil
}

// Begin a change transaction
func (cl *ChangeLog) Begin() error { return nil }

// Commit the change transaction to storage
func (cl *ChangeLog) Commit() error { return nil }

// RollBack the change transaction
func (cl *ChangeLog) RollBack() error { return nil }

// Prepare a nested transaction
func (cl *ChangeLog) Prepare(txid []byte) error { return nil }

// Update the nested transaction
func (cl *ChangeLog) Update(txid []byte) error { return nil }

// Check whether the nested transaction conflicts
func (cl *ChangeLog) Check(txid []byte) (bool, error) { return false, nil }

// Get value from changelog in given nested transaction
func (cl *ChangeLog) Get(txid []byte, key []byte) ([]byte, error) { return nil, nil }

// Put value to changelog in given nested transaction
func (cl *ChangeLog) Put(txid []byte, key []byte, val []byte) error { return nil }
