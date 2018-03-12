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

import "errors"

// const
var (
	ErrKeyNotFound = errors.New("not found")
)

// Storage interface of Storage.
type Storage interface {
	// Get return the value to the key in Storage.
	Get(key []byte) ([]byte, error)

	// Put put the key-value entry to Storage.
	Put(key []byte, value []byte) error

	// Del delete the key entry in Storage.
	Del(key []byte) error

	// EnableBatch enable batch write.
	EnableBatch()

	// DisableBatch disable batch write.
	DisableBatch()

	// Flush write and flush pending batch write.
	Flush() error
}
