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

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/common/trie"
	corepb "github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/nebulasio/go-nebulas/util/byteutils"
)

// Errors
var (
	ErrBalanceInsufficient     = errors.New("cannot subtract a value which is bigger than current balance")
	ErrAccountNotFound         = errors.New("cannot found account in storage")
	ErrContractAccountNotFound = errors.New("cannot found contract account in storage please check contract address is valid or deploy is success")
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

	contractMeta *corepb.ContractMeta
}

// ToBytes converts domain Account to bytes
func (acc *account) ToBytes() ([]byte, error) {
	value, err := acc.balance.ToFixedSizeByteSlice()
	if err != nil {
		return nil, err
	}
	pbAcc := &corepb.Account{
		Address:      acc.address,
		Balance:      value,
		Nonce:        acc.nonce,
		VarsHash:     acc.variables.RootHash(),
		BirthPlace:   acc.birthPlace,
		ContractMeta: acc.contractMeta,
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
	acc.contractMeta = pbAcc.ContractMeta
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

// ContractMeta ..
func (acc *account) ContractMeta() *corepb.ContractMeta {
	return acc.contractMeta
}

// Clone account
func (acc *account) Clone() (Account, error) {
	variables, err := acc.variables.Clone()
	if err != nil {
		return nil, err
	}

	return &account{
		address:      acc.address,
		balance:      acc.balance,
		nonce:        acc.nonce,
		variables:    variables,
		birthPlace:   acc.birthPlace,
		contractMeta: acc.contractMeta, // TODO: Clone() ?
	}, nil
}

// IncrNonce by 1
func (acc *account) IncrNonce() {
	acc.nonce++
}

// AddBalance to an account
func (acc *account) AddBalance(value *util.Uint128) error {
	balance, err := acc.balance.Add(value)
	if err != nil {
		return err
	}
	acc.balance = balance
	return nil
}

// SubBalance to an account
func (acc *account) SubBalance(value *util.Uint128) error {
	if acc.balance.Cmp(value) < 0 {
		return ErrBalanceInsufficient
	}
	balance, err := acc.balance.Sub(value)
	if err != nil {
		return err
	}
	acc.balance = balance
	return nil
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
	return fmt.Sprintf("Account %p {Address: %v, Balance:%v; Nonce:%v; VarsHash:%v; BirthPlace:%v; ContractMeta:%v}",
		acc,
		byteutils.Hex(acc.address),
		acc.balance,
		acc.nonce,
		byteutils.Hex(acc.variables.RootHash()),
		acc.birthPlace.Hex(),
		acc.contractMeta.String(), // TODO: check nil?
	)
}

// AccountState manage account state in Block
type accountState struct {
	stateTrie    *trie.Trie
	dirtyAccount map[byteutils.HexHash]Account
	storage      storage.Storage
}

// NewAccountState create a new account state
func NewAccountState(root byteutils.Hash, storage storage.Storage) (AccountState, error) {
	stateTrie, err := trie.NewTrie(root, storage, false)
	if err != nil {
		return nil, err
	}

	return &accountState{
		stateTrie:    stateTrie,
		dirtyAccount: make(map[byteutils.HexHash]Account),
		storage:      storage,
	}, nil
}

func (as *accountState) recordDirtyAccount(addr byteutils.Hash, acc Account) {
	as.dirtyAccount[addr.Hex()] = acc
}

func (as *accountState) newAccount(addr byteutils.Hash, birthPlace byteutils.Hash, contractMeta *corepb.ContractMeta) (Account, error) {
	varTrie, err := trie.NewTrie(nil, as.storage, false)
	if err != nil {
		return nil, err
	}
	acc := &account{
		address:      addr,
		balance:      util.NewUint128(),
		nonce:        0,
		variables:    varTrie,
		birthPlace:   birthPlace,
		contractMeta: contractMeta,
	}
	as.recordDirtyAccount(addr, acc)
	return acc, nil
}

func (as *accountState) getAccount(addr byteutils.Hash) (Account, error) {
	// search in dirty account
	if acc, ok := as.dirtyAccount[addr.Hex()]; ok {
		return acc, nil
	}
	// search in storage
	bytes, err := as.stateTrie.Get(addr)
	if err != nil && err != storage.ErrKeyNotFound {
		return nil, err
	}
	if err == nil {
		acc := new(account)
		err = acc.FromBytes(bytes, as.storage)
		if err != nil {
			return nil, err
		}
		as.recordDirtyAccount(addr, acc)
		return acc, nil
	}
	return nil, ErrAccountNotFound
}

func (as *accountState) Flush() error {
	for addr, acc := range as.dirtyAccount {
		bytes, err := acc.ToBytes()
		if err != nil {
			return err
		}
		key, err := addr.Hash()
		if err != nil {
			return err
		}
		as.stateTrie.Put(key, bytes)
	}
	as.dirtyAccount = make(map[byteutils.HexHash]Account)
	return nil
}

func (as *accountState) Abort() error {
	as.dirtyAccount = make(map[byteutils.HexHash]Account)
	return nil
}

// RootHash return root hash of account state
func (as *accountState) RootHash() byteutils.Hash {
	return as.stateTrie.RootHash()
}

// GetOrCreateUserAccount according to the addr
func (as *accountState) GetOrCreateUserAccount(addr byteutils.Hash) (Account, error) {
	acc, err := as.getAccount(addr)
	if err != nil && err != ErrAccountNotFound {
		return nil, err
	}
	if err == ErrAccountNotFound {
		acc, err = as.newAccount(addr, nil, nil)
		if err != nil {
			return nil, err
		}
		return acc, nil
	}
	return acc, nil
}

// GetContractAccount from current AccountState
func (as *accountState) GetContractAccount(addr byteutils.Hash) (Account, error) {
	acc, err := as.getAccount(addr)

	if err == ErrAccountNotFound {
		err = ErrContractAccountNotFound
	}
	if err != nil {
		return nil, err
	}

	return acc, nil
}

// CreateContractAccount according to the addr, and set birthPlace as creation tx hash
func (as *accountState) CreateContractAccount(addr byteutils.Hash, birthPlace byteutils.Hash, contractMeta *corepb.ContractMeta) (Account, error) {
	return as.newAccount(addr, birthPlace, contractMeta)
}

func (as *accountState) Accounts() ([]Account, error) { // TODO delete
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
	for _, account := range as.dirtyAccount {
		accounts = append(accounts, account)
	}
	return accounts, nil
}

// Relay merge the done account state
func (as *accountState) Replay(done AccountState) error {
	state := done.(*accountState)
	for addr, acc := range state.dirtyAccount {
		as.dirtyAccount[addr] = acc
	}
	return nil
}

// Clone an accountState
func (as *accountState) Clone() (AccountState, error) {
	stateTrie, err := as.stateTrie.Clone()
	if err != nil {
		return nil, err
	}

	dirtyAccount := make(map[byteutils.HexHash]Account)
	for addr, acc := range as.dirtyAccount {
		dirtyAccount[addr], err = acc.Clone()
		if err != nil {
			return nil, err
		}
	}

	return &accountState{
		stateTrie:    stateTrie,
		dirtyAccount: dirtyAccount,
		storage:      as.storage,
	}, nil
}

func (as *accountState) String() string {
	return fmt.Sprintf("AccountState %p {RootHash:%s; dirtyAccount:%v; Storage:%p}",
		as,
		byteutils.Hex(as.stateTrie.RootHash()),
		as.dirtyAccount,
		as.storage,
	)
}

// MockAccount nf/nvm/engine.CheckV8Run()  & cmd/v8/main.go
func MockAccount(version string) Account {
	return &account{
		contractMeta: &corepb.ContractMeta{
			Version: version,
		},
	}
}
