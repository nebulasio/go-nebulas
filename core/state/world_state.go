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

	"github.com/nebulasio/go-nebulas/common/mvccdb"

	"github.com/nebulasio/go-nebulas/common/trie"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util/byteutils"
)

type states struct {
	accState       AccountState
	txsState       *trie.Trie
	eventsState    *trie.Trie
	consensusState ConsensusState

	consensus Consensus
}

func newStates(consensus Consensus, storage storage.Storage) (*states, error) {
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
		consensus:      consensus,
	}, nil
}

func (s *states) Merge(done *states) error {
	_, err := s.txsState.Replay(done.txsState)
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

	err = s.accState.Replay(done.accState)
	if err != nil {
		return err
	}

	return nil
}

func (s *states) Clone() (*states, error) {
	accState, err := s.accState.Clone()
	if err != nil {
		return nil, err
	}
	txsState, err := s.txsState.Clone()
	if err != nil {
		return nil, err
	}
	eventsState, err := s.eventsState.Clone()
	if err != nil {
		return nil, err
	}
	consensusState, err := s.consensusState.Clone()
	if err != nil {
		return nil, err
	}
	return &states{
		accState:       accState,
		txsState:       txsState,
		eventsState:    eventsState,
		consensusState: consensusState,
		consensus:      s.consensus,
	}, nil
}

func (s *states) ReplaceDB(storage storage.Storage) (*states, error) {
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
	}, nil
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

func (s *states) ConsensusRoot() (byteutils.Hash, error) {
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
	return s.txsState.Get(txHash)
}

func (s *states) PutTx(txHash byteutils.Hash, txBytes []byte) error {
	_, err := s.txsState.Put(txHash, txBytes)
	return err
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

// WorldState manange all current states in Blockchain
type worldState struct {
	*states

	txStates map[string]*txWorldState
	mvccdb   *mvccdb.MVCCDB
}

// NewWorldState create a new empty WorldState
func NewWorldState(consensus Consensus, storage storage.Storage) (WorldState, error) {
	mvccdb, err := mvccdb.NewMVCCDB(storage)
	if err != nil {
		return nil, err
	}
	states, err := newStates(consensus, mvccdb)
	if err != nil {
		return nil, err
	}
	return &worldState{
		states: states,

		txStates: make(map[string]*txWorldState),
		mvccdb:   mvccdb,
	}, nil
}

func (ws *worldState) LoadAccountsRoot(root byteutils.Hash) error {
	accState, err := NewAccountState(root, ws.mvccdb)
	if err != nil {
		return err
	}
	ws.accState = accState
	return nil
}

func (ws *worldState) LoadTxsRoot(root byteutils.Hash) error {
	txsState, err := trie.NewTrie(root, ws.mvccdb)
	if err != nil {
		return err
	}
	ws.txsState = txsState
	return nil
}

func (ws *worldState) LoadEventsRoot(root byteutils.Hash) error {
	eventsState, err := trie.NewTrie(root, ws.mvccdb)
	if err != nil {
		return err
	}
	ws.eventsState = eventsState
	return nil
}

func (ws *worldState) LoadConsensusRoot(root byteutils.Hash) error {
	consensusState, err := ws.consensus.NewState(root, ws.mvccdb)
	if err != nil {
		return err
	}
	ws.consensusState = consensusState
	return nil
}

func (ws *worldState) NextConsensusState(elapsedSecond int64) (ConsensusState, error) {
	return ws.consensusState.NextConsensusState(elapsedSecond, ws)
}

func (ws *worldState) SetConsensusState(consensusState ConsensusState) {
	ws.consensusState = consensusState
}

// Clone a new WorldState
func (ws *worldState) Clone() (WorldState, error) {
	states, err := ws.states.Clone()
	if err != nil {
		return nil, err
	}
	return &worldState{
		states: states,
		mvccdb: ws.mvccdb,
	}, nil
}

func (ws *worldState) Begin() error {
	return ws.mvccdb.Begin()
}

func (ws *worldState) Commit() error {
	return ws.mvccdb.Commit()
}

func (ws *worldState) RollBack() error {
	return ws.mvccdb.RollBack()
}

type txWorldState struct {
	*states

	txid string
	db   *mvccdb.MVCCDB
}

func (ws *worldState) Prepare(txid string) (TxWorldState, error) {
	if _, ok := ws.txStates[txid]; ok {
		return nil, ErrCannotPrepareTxStateTwice
	}
	db, err := ws.mvccdb.Prepare(txid)
	if err != nil {
		return nil, err
	}
	states, err := ws.states.ReplaceDB(db)
	if err != nil {
		return nil, err
	}
	txState := &txWorldState{
		db:     db,
		txid:   txid,
		states: states,
	}
	ws.txStates[txid] = txState
	return txState, nil
}

func (ws *worldState) CheckAndUpdate(txid string) ([]interface{}, error) {
	txWorldState, ok := ws.txStates[txid]
	if !ok {
		return nil, ErrCannotUpdateTxStateBeforePrepare
	}
	dependencies, err := ws.mvccdb.CheckAndUpdate(txid)
	if err != nil {
		return nil, err
	}
	if err := ws.states.Merge(txWorldState.states); err != nil {
		return nil, err
	}
	return dependencies, nil
}

func (ws *worldState) Reset(txid string) error {
	return ws.mvccdb.Reset(txid)
}

func (ts *txWorldState) TxID() string {
	return ts.txid
}
