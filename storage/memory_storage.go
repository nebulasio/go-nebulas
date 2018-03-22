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
)

// MemoryStorage the nodes in trie.
type MemoryStorage struct {
	data *sync.Map
}

// kv entry
type kv struct{ k, v []byte }

// MemoryBatch do batch task in memory storage
type MemoryBatch struct {
	db      *MemoryStorage
	entries []*kv
}

// NewMemoryStorage init a storage
func NewMemoryStorage() (*MemoryStorage, error) {
	return &MemoryStorage{
		data: new(sync.Map),
	}, nil
}

// Get return value to the key in Storage
func (db *MemoryStorage) Get(key []byte) ([]byte, error) {
	if entry, ok := db.data.Load(byteutils.Hex(key)); ok {
		return entry.([]byte), nil
	}
	return nil, ErrKeyNotFound
}

// Put put the key-value entry to Storage
func (db *MemoryStorage) Put(key []byte, value []byte) error {
	db.data.Store(byteutils.Hex(key), value)
	return nil
}

// Del delete the key in Storage.
func (db *MemoryStorage) Del(key []byte) error {
	db.data.Delete(byteutils.Hex(key))
	return nil
}

// EnableBatch enable batch write.
func (db *MemoryStorage) EnableBatch() {
}

// Flush write and flush pending batch write.
func (db *MemoryStorage) Flush() error {
	return nil
}

// DisableBatch disable batch write.
func (db *MemoryStorage) DisableBatch() {
}
