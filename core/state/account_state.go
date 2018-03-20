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
	"errors"
	"fmt"

	"github.com/nebulasio/go-nebulas/util/logging"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/common/trie"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/nebulasio/go-nebulas/util/byteutils"
)

// Errors
var (
	ErrBalanceInsufficient = errors.New("cannot subtract a value which is bigger than current balance")
	ErrAccountNotFound     = errors.New("cannot found account in storage")
)

// account info in state Trie
type account struct {
	address byteutils.Hash
	balance *util.Uint128
	nonce   uint64
	// UserType: Global Storage
	// ContractType: Local Storage
	variables *trie.Trie
	// ContractType: Transaction Hash
	birthPlace byteutils.Hash
}

// ToBytes converts domain Account to bytes
func (acc *account) ToBytes() ([]byte, error) {
	value, err := acc.balance.ToFixedSizeByteSlice()
	if err != nil {
		return nil, err
	}
	pbAcc := &corepb.Account{
		Address:    acc.address,
		Balance:    value,
		Nonce:      acc.nonce,
		VarsHash:   acc.variables.RootHash(),
		BirthPlace: acc.birthPlace,
	}
	bytes, err := proto.Marshal(pbAcc)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// FromBytes converts bytes to Account
func (acc *account) FromBytes(bytes []byte, storage storage.Storage) error {
	pbAcc := &corepb.Account{}
	if err := proto.Unmarshal(bytes, pbAcc); err != nil {
		return err
	}
	value, err := util.NewUint128FromFixedSizeByteSlice(pbAcc.Balance)
	if err != nil {
		return err
	}
	acc.address = pbAcc.Address
	acc.balance = value
	acc.nonce = pbAcc.Nonce
	acc.birthPlace = pbAcc.BirthPlace
	acc.variables, err = trie.NewTrie(pbAcc.VarsHash, storage, false)
	if err != nil {
		return err
	}
	return nil
}

// Balance return account's balance
func (acc *account) Balance() *util.Uint128 {
	return acc.balance
}

// Address return account's address
func (acc *account) Address() byteutils.Hash {
	return acc.address
}

// Nonce return account's nonce
func (acc *account) Nonce() uint64 {
	return acc.nonce
}

// VarsHash return account's variables hash
func (acc *account) VarsHash() byteutils.Hash {
	return acc.variables.RootHash()
}

// BirthPlace return account's birth place
func (acc *account) BirthPlace() byteutils.Hash {
	return acc.birthPlace
}

// Clone account
func (acc *account) CopyTo(storage storage.Storage, needChangeLog bool) (Account, error) {
	variables, err := acc.variables.CopyTo(storage, needChangeLog)
	if err != nil {
		return nil, err
	}

	return &account{
		address:    acc.address,
		balance:    acc.balance,
		nonce:      acc.nonce,
		variables:  variables,
		birthPlace: acc.birthPlace,
	}, nil
}

// IncrNonce by 1
func (acc *account) IncrNonce() {
	acc.nonce++
}

// AddBalance to an account
func (acc *account) AddBalance(value *util.Uint128) error {
	var err error
	acc.balance, err = acc.balance.Add(value)
	return err
}

// SubBalance to an account
func (acc *account) SubBalance(value *util.Uint128) error {
	var err error
	if acc.balance.Cmp(value) < 0 {
		err = ErrBalanceInsufficient
	} else {
		acc.balance, err = acc.balance.Sub(value)
	}
	return err
}

// Put into account's storage
func (acc *account) Put(key []byte, value []byte) error {
	_, err := acc.variables.Put(key, value)
	return err
}

// Get from account's storage
func (acc *account) Get(key []byte) ([]byte, error) {
	return acc.variables.Get(key)
}

// Del from account's storage
func (acc *account) Del(key []byte) error {
	if _, err := acc.variables.Del(key); err != nil {
		return err
	}
	return nil
}

// Iterator map var from account's storage
func (acc *account) Iterator(prefix []byte) (Iterator, error) {
	return acc.variables.Iterator(prefix)
}

func (acc *account) String() string {
	return fmt.Sprintf("Account %p {Address: %v, Balance:%v; Nonce:%v; VarsHash:%v; BirthPlace:%v}",
		acc,
		byteutils.Hex(acc.address),
		acc.balance.Int,
		acc.nonce,
		byteutils.Hex(acc.variables.RootHash()),
		acc.birthPlace.Hex(),
	)
}

type cachedAccount struct {
	acc   Account
	dirty bool
}

// AccountState manage account state in Block
type accountState struct {
	stateTrie      *trie.Trie
	cachedAccounts map[byteutils.HexHash]*cachedAccount
	storage        storage.Storage
}

// NewAccountState create a new account state
func NewAccountState(root byteutils.Hash, storage storage.Storage, needChangeLog bool) (AccountState, error) {
	stateTrie, err := trie.NewTrie(root, storage, needChangeLog)
	if err != nil {
		return nil, err
	}

	return &accountState{
		stateTrie:      stateTrie,
		cachedAccounts: make(map[byteutils.HexHash]*cachedAccount),
		storage:        storage,
	}, nil
}

func (as *accountState) recordDirtyAccount(addr byteutils.Hash, acc Account) {
	logging.CLog().Info("Record Dirty ", addr.String(), " ", acc.Balance().String(), " ", acc.Nonce())
	if _, ok := as.cachedAccounts[addr.Hex()]; ok {
		as.cachedAccounts[addr.Hex()].dirty = true
	} else {
		as.cachedAccounts[addr.Hex()] = &cachedAccount{acc, true}
	}
}

func (as *accountState) newAccount(addr byteutils.Hash, birthPlace byteutils.Hash) (Account, error) {
	varTrie, err := trie.NewTrie(nil, as.storage, false)
	if err != nil {
		return nil, err
	}
	acc := &account{
		address:    addr,
		balance:    util.NewUint128(),
		nonce:      0,
		variables:  varTrie,
		birthPlace: birthPlace,
	}
	return acc, nil
}

func (as *accountState) getAccount(addr byteutils.Hash) (Account, error) {
	// search in dirty account
	if ca, ok := as.cachedAccounts[addr.Hex()]; ok {
		return ca.acc, nil
	}
	// search in storage
	bytes, err := as.stateTrie.Get(addr)
	if err == nil {
		acc := new(account)
		err = acc.FromBytes(bytes, as.storage)
		if err != nil {
			return nil, err
		}
		return acc, nil
	}
	return nil, ErrAccountNotFound
}

// RootHash return root hash of account state
func (as *accountState) RootHash() (byteutils.Hash, error) {
	if err := as.flush(); err != nil {
		return nil, err
	}
	return as.stateTrie.RootHash(), nil
}

// GetOrCreateUserAccount according to the addr
func (as *accountState) GetOrCreateUserAccount(addr byteutils.Hash) (Account, error) {
	acc, err := as.getAccount(addr)
	if err != nil {
		acc, err := as.newAccount(addr, nil)
		if err != nil {
			return nil, err
		}
		as.recordDirtyAccount(addr, acc)
		return acc, nil
	}
	as.recordDirtyAccount(addr, acc)
	return acc, nil
}

// GetContractAccount from current AccountState
func (as *accountState) GetContractAccount(addr byteutils.Hash) (Account, error) {
	acc, err := as.getAccount(addr)
	if err != nil {
		return nil, err
	}
	as.recordDirtyAccount(addr, acc)
	return acc, nil
}

// CreateContractAccount according to the addr, and set birthPlace as creation tx hash
func (as *accountState) CreateContractAccount(addr byteutils.Hash, birthPlace byteutils.Hash) (Account, error) {
	acc, err := as.newAccount(addr, birthPlace)
	if err != nil {
		return nil, err
	}
	as.recordDirtyAccount(addr, acc)
	return acc, nil
}

func (as *accountState) Accounts() ([]Account, error) {
	accounts := []Account{}
	iter, err := as.stateTrie.Iterator(nil)
	if err != nil && err != storage.ErrKeyNotFound {
		return nil, err
	}
	if err != nil {
		return accounts, nil
	}
	exist, err := iter.Next()
	if err != nil {
		return nil, err
	}
	for exist {
		acc := new(account)
		err = acc.FromBytes(iter.Value(), as.storage)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, acc)
		exist, err = iter.Next()
		if err != nil {
			return nil, err
		}
	}
	return accounts, nil
}

// DirtyAccounts return all changed accounts
func (as *accountState) DirtyAccounts() ([]Account, error) {
	accounts := []Account{}
	for _, ca := range as.cachedAccounts {
		if ca.dirty {
			accounts = append(accounts, ca.acc)
		}
	}
	return accounts, nil
}

func (as *accountState) RollBackAccounts() {
	as.cachedAccounts = make(map[byteutils.HexHash]*cachedAccount)
}

func (as *accountState) flush() error {
	for addr, ca := range as.cachedAccounts {
		bytes, err := ca.acc.ToBytes()
		if err != nil {
			return err
		}
		key, err := addr.Hash()
		if err != nil {
			return err
		}
		if _, err := as.stateTrie.Put(key, bytes); err != nil {
			return err
		}
	}
	return nil
}

func (as *accountState) CommitAccounts() error {
	if err := as.flush(); err != nil {
		return err
	}
	as.cachedAccounts = make(map[byteutils.HexHash]*cachedAccount)
	return nil
}

// Relay merge the done account state
func (as *accountState) Replay(done AccountState) error {
	state := done.(*accountState)
	for addr, ca := range state.cachedAccounts {
		if ca.dirty {
			bytes, err := ca.acc.ToBytes()
			if err != nil {
				return err
			}
			key, err := addr.Hash()
			if err != nil {
				return err
			}
			if _, err := as.stateTrie.Put(key, bytes); err != nil {
				return err
			}
			delete(as.cachedAccounts, addr)
		}
	}
	return nil
}

// Clone an accountState
func (as *accountState) CopyTo(storage storage.Storage, needChangeLog bool) (AccountState, error) {
	stateTrie, err := as.stateTrie.CopyTo(storage, needChangeLog)
	if err != nil {
		return nil, err
	}

	cachedAccounts := make(map[byteutils.HexHash]*cachedAccount)
	for addr, ca := range as.cachedAccounts {
		copyAcc, err := ca.acc.CopyTo(storage, needChangeLog)
		if err != nil {
			return nil, err
		}
		cachedAccounts[addr] = &cachedAccount{copyAcc, false}
	}

	return &accountState{
		stateTrie:      stateTrie,
		cachedAccounts: cachedAccounts,
		storage:        storage,
	}, nil
}
