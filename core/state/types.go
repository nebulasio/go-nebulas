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

	consensuspb "github.com/nebulasio/go-nebulas/consensus/pb"
	corepb "github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/nebulasio/go-nebulas/util/byteutils"
)

// Errors
var (
	ErrCannotPrepareTxStateTwice           = errors.New("cannot prepare tx state twice")
	ErrCannotCloneOngoingWorldState        = errors.New("cannot clone an ongoing world state")
	ErrCannotBeginWorldStateTwice          = errors.New("cannot begin a world state twice")
	ErrCannotCommitWorldStateBeforeBegin   = errors.New("cannot commit a world state before begin")
	ErrCannotRollBackWorldStateBeforeBegin = errors.New("cannot rollback a world state before begin")
	ErrCannotPrepareTxStateBeforeBegin     = errors.New("cannot prepare a tx state before begin")
	ErrCannotCheckTxStateBeforePrepare     = errors.New("cannot check a tx state before prepare")
	ErrCannotUpdateTxStateBeforePrepare    = errors.New("cannot update a tx state before prepare")
	ErrCannotResetTxStateBeforePrepare     = errors.New("cannot reset a tx state before prepare")
	ErrContractCheckFailed                 = errors.New("contract check failed")
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

	Clone() (Account, error)

	ToBytes() ([]byte, error)
	FromBytes(bytes []byte, storage storage.Storage) error

	IncrNonce()
	AddBalance(value *util.Uint128) error
	SubBalance(value *util.Uint128) error
	Put(key []byte, value []byte) error
	Get(key []byte) ([]byte, error)
	Del(key []byte) error
	Iterator(prefix []byte) (Iterator, error)
	ContractMeta() *corepb.ContractMeta
}

// AccountState Interface
type AccountState interface {
	RootHash() byteutils.Hash

	Flush() error
	Abort() error

	DirtyAccounts() ([]Account, error)
	Accounts() ([]Account, error)

	Clone() (AccountState, error)
	Replay(AccountState) error

	GetOrCreateUserAccount(byteutils.Hash) (Account, error)
	GetContractAccount(byteutils.Hash) (Account, error)
	CreateContractAccount(byteutils.Hash, byteutils.Hash, *corepb.ContractMeta) (Account, error)
}

// Event event structure.
type Event struct {
	Topic string
	Data  string
}

// Consensus interface
type Consensus interface {
	NewState(*consensuspb.ConsensusRoot, storage.Storage, bool) (ConsensusState, error)
}

// ConsensusState interface of consensus state
type ConsensusState interface {
	RootHash() *consensuspb.ConsensusRoot
	String() string
	Clone() (ConsensusState, error)
	Replay(ConsensusState) error

	Proposer() byteutils.Hash
	TimeStamp() int64
	NextConsensusState(int64, WorldState) (ConsensusState, error)

	Dynasty() ([]byteutils.Hash, error)
	DynastyRoot() byteutils.Hash
}

// WorldState interface of world state
type WorldState interface {
	Begin() error
	Commit() error
	RollBack() error

	Prepare(interface{}) (TxWorldState, error)
	Reset(addr byteutils.Hash, isResetChangeLog bool) error
	Flush() error
	Abort() error

	LoadAccountsRoot(byteutils.Hash) error
	LoadTxsRoot(byteutils.Hash) error
	LoadEventsRoot(byteutils.Hash) error
	LoadConsensusRoot(*consensuspb.ConsensusRoot) error

	NextConsensusState(int64) (ConsensusState, error)
	SetConsensusState(ConsensusState)

	Clone() (WorldState, error)

	AccountsRoot() byteutils.Hash
	TxsRoot() byteutils.Hash
	EventsRoot() byteutils.Hash
	ConsensusRoot() *consensuspb.ConsensusRoot

	Accounts() ([]Account, error)
	GetOrCreateUserAccount(addr byteutils.Hash) (Account, error)
	GetContractAccount(addr byteutils.Hash) (Account, error)
	CreateContractAccount(owner byteutils.Hash, birthPlace byteutils.Hash, contractMeta *corepb.ContractMeta) (Account, error)

	GetTx(txHash byteutils.Hash) ([]byte, error)
	PutTx(txHash byteutils.Hash, txBytes []byte) error

	RecordEvent(txHash byteutils.Hash, event *Event)
	FetchEvents(byteutils.Hash) ([]*Event, error)

	Dynasty() ([]byteutils.Hash, error)
	DynastyRoot() byteutils.Hash

	RecordGas(from string, gas *util.Uint128) error
	GetGas() map[string]*util.Uint128
	GetBlockHashByHeight(height uint64) ([]byte, error)
	GetBlock(txHash byteutils.Hash) ([]byte, error)
}

// TxWorldState is the world state of a single transaction
type TxWorldState interface {
	AccountsRoot() byteutils.Hash
	TxsRoot() byteutils.Hash
	EventsRoot() byteutils.Hash
	ConsensusRoot() *consensuspb.ConsensusRoot

	CheckAndUpdate() ([]interface{}, error)
	Reset(addr byteutils.Hash, isResetChangeLog bool) error
	Close() error

	Accounts() ([]Account, error)
	GetOrCreateUserAccount(addr byteutils.Hash) (Account, error)
	GetContractAccount(addr byteutils.Hash) (Account, error)
	CreateContractAccount(owner byteutils.Hash, birthPlace byteutils.Hash, contractMeta *corepb.ContractMeta) (Account, error)

	GetTx(txHash byteutils.Hash) ([]byte, error)
	PutTx(txHash byteutils.Hash, txBytes []byte) error

	RecordEvent(txHash byteutils.Hash, event *Event)
	FetchEvents(byteutils.Hash) ([]*Event, error)

	Dynasty() ([]byteutils.Hash, error)
	DynastyRoot() byteutils.Hash

	RecordGas(from string, gas *util.Uint128) error
	GetBlockHashByHeight(height uint64) ([]byte, error)
	GetBlock(txHash byteutils.Hash) ([]byte, error)
}
