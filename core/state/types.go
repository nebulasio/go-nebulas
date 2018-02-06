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

package state

import (
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/nebulasio/go-nebulas/util/byteutils"
)

// Iterator Variables in Account Storage
type Iterator interface {
	Next() (bool, error)
	Value() []byte
}

// Account Interface
type Account interface {
	Address() byteutils.Hash
	Balance() *util.Uint128
	Nonce() uint64
	BirthPlace() byteutils.Hash
	VarsHash() byteutils.Hash

	BeginBatch()
	Commit()
	RollBack()
	Clone() (Account, error)

	ToBytes() ([]byte, error)
	FromBytes(bytes []byte, storage storage.Storage) error

	IncrNonce()
	AddBalance(value *util.Uint128)
	SubBalance(value *util.Uint128) error
	Put(key []byte, value []byte) error
	Get(key []byte) ([]byte, error)
	Del(key []byte) error
	Iterator(prefix []byte) (Iterator, error)
}

// AccountState Interface
type AccountState interface {
	RootHash() (byteutils.Hash, error)
	Accounts() ([]Account, error)

	DirtyAccountSize() int

	BeginBatch()
	Commit() error
	RollBack()

	Clone() (AccountState, error)

	GetOrCreateUserAccount(addr []byte) (Account, error)
	GetContractAccount(addr []byte) (Account, error)
	CreateContractAccount(addr []byte, birthPlace []byte) (Account, error)
}
