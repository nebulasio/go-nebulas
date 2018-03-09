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
	"encoding/json"

	"github.com/nebulasio/go-nebulas/consensus/pb"

	"github.com/nebulasio/go-nebulas/util/logging"

	"github.com/nebulasio/go-nebulas/common/mvccdb"

	"github.com/nebulasio/go-nebulas/common/trie"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util/byteutils"
)

func newChangeLog() (*mvccdb.MVCCDB, error) {
	mem, err := storage.NewMemoryStorage()
	if err != nil {
		return nil, err
	}
	return mvccdb.NewMVCCDB(mem, true)
}

func newStorage(storage storage.Storage) (*mvccdb.MVCCDB, error) {
	return mvccdb.NewMVCCDB(storage, false)
}

type states struct {
	accState       AccountState
	txsState       *trie.Trie
	eventsState    *trie.Trie
	consensusState ConsensusState

	consensus Consensus
	changelog *mvccdb.MVCCDB
	storage   *mvccdb.MVCCDB
	txid      interface{}
}

func newStates(consensus Consensus, stor storage.Storage) (*states, error) {
	changelog, err := newChangeLog()
	if err != nil {
		return nil, err
	}
	storage, err := newStorage(stor)
	if err != nil {
		return nil, err
	}

	accState, err := NewAccountState(nil, storage)
	if err != nil {
		return nil, err
	}
	txsState, err := trie.NewTrie(nil, storage)
	if err != nil {
		return nil, err
	}
	eventsState, err := trie.NewTrie(nil, storage)
	if err != nil {
		return nil, err
	}

	return &states{
		accState:       accState,
		txsState:       txsState,
		eventsState:    eventsState,
		consensusState: nil,

		consensus: consensus,
		changelog: changelog,
		storage:   storage,
		txid:      nil,
	}, nil
}

func (s *states) Replay(done *states) error {
	err := s.accState.Replay(done.accState)
	if err != nil {
		return err
	}
	_, err = s.txsState.Replay(done.txsState)
	if err != nil {
		return err
	}
	_, err = s.eventsState.Replay(done.eventsState)
	if err != nil {
		return err
	}
	err = s.consensusState.Replay(done.consensusState)
	if err != nil {
		return err
	}

	return nil
}

func (s *states) Clone() (WorldState, error) {
	changelog, err := newChangeLog()
	if err != nil {
		return nil, err
	}
	storage, err := newStorage(s.storage)
	if err != nil {
		return nil, err
	}

	accRoot, err := s.accState.RootHash()
	if err != nil {
		return nil, err
	}
	accState, err := NewAccountState(accRoot, storage)
	if err != nil {
		return nil, err
	}
	txsState, err := trie.NewTrie(s.txsState.RootHash(), storage)
	if err != nil {
		return nil, err
	}
	eventsState, err := trie.NewTrie(s.eventsState.RootHash(), storage)
	if err != nil {
		return nil, err
	}
	consensusRoot, err := s.consensusState.RootHash()
	if err != nil {
		return nil, err
	}
	consensusState, err := s.consensus.NewState(consensusRoot, storage)
	if err != nil {
		return nil, err
	}

	return &states{
		accState:       accState,
		txsState:       txsState,
		eventsState:    eventsState,
		consensusState: consensusState,

		consensus: s.consensus,
		changelog: changelog,
		storage:   storage,
		txid:      s.txid,
	}, nil
}

func (s *states) Begin() error {
	if err := s.changelog.Begin(); err != nil {
		return err
	}
	return s.storage.Begin()
}

func (s *states) Commit() error {
	if err := s.changelog.Commit(); err != nil {
		return err
	}
	return s.storage.Commit()
}

func (s *states) RollBack() error {
	if err := s.changelog.RollBack(); err != nil {
		return err
	}
	return s.storage.RollBack()
}

func (s *states) Prepare(txid interface{}) (TxWorldState, error) {
	changelog, err := s.changelog.Prepare(txid)
	if err != nil {
		return nil, err
	}
	storage, err := s.storage.Prepare(txid)
	if err != nil {
		return nil, err
	}

	accRoot, err := s.AccountsRoot()
	if err != nil {
		return nil, err
	}
	accState, err := NewAccountState(accRoot, storage)
	if err != nil {
		return nil, err
	}
	txsRoot, err := s.TxsRoot()
	if err != nil {
		return nil, err
	}
	txsState, err := trie.NewTrie(txsRoot, storage)
	if err != nil {
		return nil, err
	}
	eventsRoot, err := s.EventsRoot()
	if err != nil {
		return nil, err
	}
	eventsState, err := trie.NewTrie(eventsRoot, storage)
	if err != nil {
		return nil, err
	}
	consensusRoot, err := s.ConsensusRoot()
	if err != nil {
		return nil, err
	}
	consensusState, err := s.consensus.NewState(consensusRoot, storage)
	if err != nil {
		return nil, err
	}

	return &states{
		accState:       accState,
		txsState:       txsState,
		eventsState:    eventsState,
		consensusState: consensusState,

		consensus: s.consensus,
		changelog: changelog,
		storage:   storage,
		txid:      txid,
	}, nil
}

func (s *states) recordAccounts() error {
	accounts, err := s.accState.DirtyAccounts()
	if err != nil {
		return err
	}
	// record change log
	for _, account := range accounts {
		bytes, err := account.ToBytes()
		if err != nil {
			return err
		}
		if err := s.changelog.Put(account.Address(), bytes); err != nil {
			return err
		}
		logging.CLog().Info(s.txid, " [Put] Account:", account.Address().String())
	}
	return nil
}

func (s *states) CheckAndUpdate(txid interface{}) ([]interface{}, error) {
	if err := s.recordAccounts(); err != nil {
		return nil, err
	}
	dependency, err := s.changelog.CheckAndUpdate()
	if err != nil {
		return nil, err
	}
	_, err = s.storage.CheckAndUpdate()
	if err != nil {
		return nil, err
	}
	return dependency, nil
}

func (s *states) Reset(txid interface{}) error {
	if err := s.changelog.Reset(); err != nil {
		return err
	}
	if err := s.storage.Reset(); err != nil {
		return err
	}
	return nil
}

func (s *states) AccountsRoot() (byteutils.Hash, error) {
	return s.accState.RootHash()
}

func (s *states) TxsRoot() (byteutils.Hash, error) {
	return s.txsState.RootHash(), nil
}

func (s *states) EventsRoot() (byteutils.Hash, error) {
	return s.eventsState.RootHash(), nil
}

func (s *states) ConsensusRoot() (*consensuspb.ConsensusRoot, error) {
	return s.consensusState.RootHash()
}

func (s *states) GetOrCreateUserAccount(addr byteutils.Hash) (Account, error) {
	return s.accState.GetOrCreateUserAccount(addr)
}

func (s *states) GetContractAccount(addr byteutils.Hash) (Account, error) {
	return s.accState.GetContractAccount(addr)
}

func (s *states) CreateContractAccount(owner byteutils.Hash, birthPlace byteutils.Hash) (Account, error) {
	return s.accState.CreateContractAccount(owner, birthPlace)
}

func (s *states) GetTx(txHash byteutils.Hash) ([]byte, error) {
	bytes, err := s.txsState.Get(txHash)
	if err != nil {
		return nil, err
	}
	// record change log
	if _, err := s.changelog.Get(txHash); err != nil && err != storage.ErrKeyNotFound {
		return nil, err
	}
	logging.CLog().Info(s.txid, " [Get] Tx:", txHash.String())
	return bytes, nil
}

func (s *states) PutTx(txHash byteutils.Hash, txBytes []byte) error {
	_, err := s.txsState.Put(txHash, txBytes)
	if err != nil {
		return err
	}
	// record change log
	if err := s.changelog.Put(txHash, txBytes); err != nil {
		return err
	}
	logging.CLog().Info(s.txid, " [Put] Tx:", txHash.String())
	return nil
}

func (s *states) RecordEvent(txHash byteutils.Hash, event *Event) error {
	iter, err := s.eventsState.Iterator(txHash)
	if err != nil && err != storage.ErrKeyNotFound {
		return err
	}
	cnt := int64(0)
	if err != storage.ErrKeyNotFound {
		exist, err := iter.Next()
		if err != nil {
			return err
		}
		for exist {
			cnt++
			exist, err = iter.Next()
			if err != nil {
				return err
			}
		}
	}
	cnt++
	key := append(txHash, byteutils.FromInt64(cnt)...)
	bytes, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, err = s.eventsState.Put(key, bytes)
	if err != nil {
		return err
	}
	// record change log
	if err := s.changelog.Put(key, bytes); err != nil {
		return err
	}
	logging.CLog().Info(s.txid, " [Put] Event:", byteutils.Hex(key))
	return nil
}

func (s *states) FetchEvents(txHash byteutils.Hash) ([]*Event, error) {
	events := []*Event{}
	iter, err := s.eventsState.Iterator(txHash)
	if err != nil && err != storage.ErrKeyNotFound {
		return nil, err
	}
	if err != storage.ErrKeyNotFound {
		exist, err := iter.Next()
		if err != nil {
			return nil, err
		}
		for exist {
			event := new(Event)
			err = json.Unmarshal(iter.Value(), event)
			if err != nil {
				return nil, err
			}
			events = append(events, event)
			// record change log
			if _, err := s.changelog.Get(iter.Key()); err != nil && err != storage.ErrKeyNotFound {
				return nil, err
			}
			logging.CLog().Info(s.txid, " [Get] Event:", byteutils.Hex(iter.Key()))
			exist, err = iter.Next()
			if err != nil {
				return nil, err
			}
		}
	}
	return events, nil
}

func (s *states) Dynasty() ([]byteutils.Hash, error) {
	return s.consensusState.Dynasty()
}

func (s *states) DynastyRoot() byteutils.Hash {
	return s.consensusState.DynastyRoot()
}

func (s *states) Accounts() ([]Account, error) {
	return s.accState.Accounts()
}

func (s *states) LoadAccountsRoot(root byteutils.Hash) error {
	accState, err := NewAccountState(root, s.storage)
	if err != nil {
		return err
	}
	s.accState = accState
	return nil
}

func (s *states) LoadTxsRoot(root byteutils.Hash) error {
	txsState, err := trie.NewTrie(root, s.storage)
	if err != nil {
		return err
	}
	s.txsState = txsState
	return nil
}

func (s *states) LoadEventsRoot(root byteutils.Hash) error {
	eventsState, err := trie.NewTrie(root, s.storage)
	if err != nil {
		return err
	}
	s.eventsState = eventsState
	return nil
}

func (s *states) LoadConsensusRoot(root *consensuspb.ConsensusRoot) error {
	consensusState, err := s.consensus.NewState(root, s.storage)
	if err != nil {
		return err
	}
	s.consensusState = consensusState
	return nil
}

func (s *states) NextConsensusState(elapsedSecond int64) (ConsensusState, error) {
	return s.consensusState.NextConsensusState(elapsedSecond, s)
}

func (s *states) SetConsensusState(consensusState ConsensusState) {
	s.consensusState = consensusState
}

// WorldState manange all current states in Blockchain
type worldState struct {
	*states
	txStates map[interface{}]*txWorldState
}

// NewWorldState create a new empty WorldState
func NewWorldState(consensus Consensus, storage storage.Storage) (WorldState, error) {
	states, err := newStates(consensus, storage)
	if err != nil {
		return nil, err
	}
	return &worldState{
		states:   states,
		txStates: make(map[interface{}]*txWorldState),
	}, nil
}

// Clone a new WorldState
func (ws *worldState) Clone() (WorldState, error) {
	s, err := ws.states.Clone()
	if err != nil {
		return nil, err
	}
	return &worldState{
		states:   s.(*states),
		txStates: make(map[interface{}]*txWorldState),
	}, nil
}

func (ws *worldState) Begin() error {
	if err := ws.states.Begin(); err != nil {
		return err
	}
	return nil
}

func (ws *worldState) Commit() error {
	if err := ws.states.Commit(); err != nil {
		return err
	}
	return nil
}

func (ws *worldState) RollBack() error {
	if err := ws.states.RollBack(); err != nil {
		return err
	}
	return nil
}

type txWorldState struct {
	*states
	txid interface{}
}

func (ws *worldState) Prepare(txid interface{}) (TxWorldState, error) {
	if _, ok := ws.txStates[txid]; ok {
		return nil, ErrCannotPrepareTxStateTwice
	}
	s, err := ws.states.Prepare(txid)
	if err != nil {
		return nil, err
	}
	txState := &txWorldState{
		states: s.(*states),
		txid:   txid,
	}
	ws.txStates[txid] = txState
	return txState, nil
}

func (ws *worldState) CheckAndUpdate(txid interface{}) ([]interface{}, error) {
	txWorldState, ok := ws.txStates[txid]
	if !ok {
		return nil, ErrCannotUpdateTxStateBeforePrepare
	}
	dependencies, err := txWorldState.CheckAndUpdate(txid)
	if err != nil {
		return nil, err
	}
	if err := ws.states.Replay(txWorldState.states); err != nil {
		return nil, err
	}
	return dependencies, nil
}

func (ws *worldState) Reset(txid interface{}) error {
	txWorldState, ok := ws.txStates[txid]
	if !ok {
		return ErrCannotUpdateTxStateBeforePrepare
	}
	if err := txWorldState.Reset(txid); err != nil {
		return err
	}
	delete(ws.txStates, txid)
	return nil
}

func (ts *txWorldState) TxID() interface{} {
	return ts.txid
}
