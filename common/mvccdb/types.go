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

// ChangeLog interface
type ChangeLog interface {
	Prepare(interface{}) (*MVCCDB, error)
	CheckAndUpdate(interface{}) ([]interface{}, error)
	Reset(interface{}) error

	Get([]byte) ([]byte, error)
	Put([]byte, []byte) error
	Del([]byte) error
}

// NewChangeLog create a new change log
func NewChangeLog() (ChangeLog, error) {
	mem, err := storage.NewMemoryStorage()
	if err != nil {
		return nil, err
	}
	return NewMVCCDB(mem)
}

// Storage interface
type Storage interface {
	Begin() error
	Commit() error
	RollBack() error

	Prepare(interface{}) (*MVCCDB, error)
	CheckAndUpdate(interface{}) ([]interface{}, error)
	Reset(interface{}) error

	Get([]byte) ([]byte, error)
	Put([]byte, []byte) error
	Del([]byte) error
}

// NewStorage return a storage with mvcc support
func NewStorage(storage storage.Storage) (Storage, error) {
	return NewMVCCDB(storage)
}
