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

package core

import (
	"errors"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/common/trie"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/storage"

	"github.com/nebulasio/go-nebulas/util"
)

// Account info in state Trie
type Account struct {
	UserBalance       *util.Uint128
	UserNonce         uint64
	UserGlobalStorage *trie.BatchTrie

	ContractOwner           *Address
	ContractTransactionHash Hash
	ContractLocalStorage    *trie.BatchTrie

	storage storage.Storage
}

// NewAccount create a new account
func NewAccount(storage storage.Storage) *Account {
	globalTrie, _ := trie.NewBatchTrie(nil, storage)
	localTrie, _ := trie.NewBatchTrie(nil, storage)
	return &Account{
		UserBalance:       util.NewUint128(),
		UserNonce:         0,
		UserGlobalStorage: globalTrie,

		ContractOwner:           &Address{address: Hash{}},
		ContractTransactionHash: Hash{},
		ContractLocalStorage:    localTrie,

		storage: storage,
	}
}

// IncreNonce by 1
func (acc *Account) IncreNonce() {
	acc.UserNonce++
}

// AddBalance to an account
func (acc *Account) AddBalance(value *util.Uint128) {
	acc.UserBalance.Add(acc.UserBalance.Int, value.Int)
}

// SubBalance from an account
func (acc *Account) SubBalance(value *util.Uint128) {
	if acc.UserBalance.Cmp(value.Int) < 0 {
		panic("cannot subtract a value which is bigger than current balance")
	}
	acc.UserBalance.Sub(acc.UserBalance.Int, value.Int)
}

// SetContractTransactionHash in account
func (acc *Account) SetContractTransactionHash(code []byte) {
	acc.ContractTransactionHash = code
}

// SetContractOwner with a address
func (acc *Account) SetContractOwner(address *Address) {
	acc.ContractOwner = address
}

// ToProto converts domain Account to proto Account
func (acc *Account) ToProto() (proto.Message, error) {
	value, err := acc.UserBalance.ToFixedSizeByteSlice()
	if err != nil {
		return nil, err
	}
	return &corepb.Account{
		UserBalance:             value,
		UserNonce:               acc.UserNonce,
		UserGlobalStorage:       acc.UserGlobalStorage.RootHash(),
		ContractOwner:           acc.ContractOwner.address,
		ContractTransactionHash: acc.ContractTransactionHash,
		ContractLocalStorage:    acc.ContractLocalStorage.RootHash(),
	}, nil
}

// FromProto converts proto Account to domain Account
func (acc *Account) FromProto(msg proto.Message) error {
	if msg, ok := msg.(*corepb.Account); ok {
		if acc.storage == nil {
			return errors.New("account's storage cannot be nil before FromProto")
		}

		value, err := util.NewUint128FromFixedSizeByteSlice(msg.UserBalance)
		if err != nil {
			return err
		}
		acc.UserBalance = value
		acc.UserNonce = msg.UserNonce
		acc.UserGlobalStorage, err = trie.NewBatchTrie(msg.UserGlobalStorage, acc.storage)
		if err != nil {
			return err
		}
		acc.ContractOwner = &Address{msg.ContractOwner}
		acc.ContractTransactionHash = msg.ContractTransactionHash
		acc.ContractLocalStorage, err = trie.NewBatchTrie(msg.ContractLocalStorage, acc.storage)
		if err != nil {
			return err
		}
		return nil
	}
	return errors.New("Pb Message cannot be converted into Account")
}
