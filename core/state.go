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
	"github.com/nebulasio/go-nebulas/common/trie"
	"github.com/nebulasio/go-nebulas/storage"

	"github.com/nebulasio/go-nebulas/util"
)

// Account info in state Trie
type Account struct {
	balance *util.Uint128
	nonce   uint64
	// UserType: Global Storage
	// ContractType: Local Storage
	storage *trie.BatchTrie
	// Contract Location: Transaction Hash
	location Hash
}

// Hash an account
func (acc *Account) Hash() Hash {
	return nil
}

// AccountState manage account state in Block
type AccountState struct {
	stateTrie    *trie.BatchTrie
	dirtyAccount map[string]*Account
	inBatch      bool
	storage      storage.Storage
}

// NewAccountState create a new account state
func NewAccountState(storage storage.Storage) *Account {
	stateTrie, _ := trie.NewBatchTrie(nil, storage)
	return &AccountState{
		stateTrie:    stateTrie,
		dirtyAccount: make(map[string]*Account),
		inBatch:      false,
		storage:      storage,
	}
}

// SetLocation to a contract account
func (state *AccountState) SetLocation(accAddr Hash, location Hash) {
	acc := GetAccount(accAddr)
	acc.location = location
}

// GetAccount according to the addr
func (state *AccountState) GetAccount(accAddr Hash) *Account {
	return nil
}

// IncreNonce by 1
func (state *AccountState) IncreNonce(accAddr Hash) error {
	return nil
}

// AddBalance to an account
func (state *AccountState) AddBalance(accAddr Hash, value *util.Uint128) error {
	return nil
}

// SubBalance to an account
func (state *AccountState) SubBalance(accAddr Hash, value *util.Uint128) error {
	return nil
}

// PutVar into account's storage
func (state *AccountState) PutVar(accAddr Hash, key []byte, value []byte) error {
	return nil
}

// GetVar from account's storage
func (state *AccountState) GetVar(accAddr Hash, key []byte) ([]byte, error) {
	return nil
}

// PutMapVar into account's storage
func (state *AccountState) PutMapVar(accAddr Hash, value []byte, domains ...string) error {
	return nil
}

// Iterator map var from account's storage
func (state *AccountState) Iterator(accAddr Hash, domains ...string) (*trie.Iterator, error) {
	return nil, nil
}

// Begin a batch task
func (state *AccountState) Begin() {
}

// Commit a batch task
func (state *AccountState) Commit() {
}

// RollBack a batch task
func (state *AccountState) RollBack() {
}
