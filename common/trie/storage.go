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
	"sync"
)

// Storage the nodes in trie.
type Storage struct {
	data map[string][]byte
	lock sync.RWMutex
}

// Get return value to the key in Storage
func (db *Storage) Get(key []byte) ([]byte, error) {
	return nil, nil
}

// Put put the key-value entry to Storage
func (db *Storage) Put(key []byte, value []byte) error {
	return nil
}

// Del delete the key in Storage.
func (db *Storage) Del(key []byte) error {
	return nil
}
