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

package trie

import (
	"errors"
	"github.com/nebulasio/go-nebulas/utils/byteutils"
)

// Storage the nodes in trie.
type Storage struct {
	data map[string][]byte
}

// NewStorage init a storage
func NewStorage() (*Storage, error) {
	return &Storage{
		data: make(map[string][]byte),
	}, nil
}

// Get return value to the key in Storage
func (db *Storage) Get(key []byte) ([]byte, error) {
	if entry, ok := db.data[byteutils.Hex(key)]; ok {
		return entry, nil
	}
	return nil, errors.New("not found")
}

// Put put the key-value entry to Storage
func (db *Storage) Put(key []byte, value []byte) error {
	db.data[byteutils.Hex(key)] = value
	return nil
}

// Del delete the key in Storage.
func (db *Storage) Del(key []byte) error {
	delete(db.data, byteutils.Hex(key))
	return nil
}
