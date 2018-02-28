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
	txsState       *trie.BatchTrie
	eventsState    *trie.BatchTrie
	consensusState ConsensusState

	consensus Consensus
}

func newStates(consensus Consensus, storage storage.Storage) (*states, error) {
	accState, err := NewAccountState(nil, storage)
	if err != nil {
		return nil, err
	}
	txsState, err := trie.NewBatchTrie(nil, storage)
	if err != nil {
		return nil, err
	}
	eventsState, err := trie.NewBatchTrie(nil, storage)
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
	txsState, err := trie.NewBatchTrie(txsRoot, storage)
	if err != nil {
		return nil, err
	}
	eventsRoot, err := s.EventsRoot()
	if err != nil {
		return nil, err
	}
	eventsState, err := trie.NewBatchTrie(eventsRoot, storage)
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

func (s *states) GetMintCnt(timestamp int64, miner byteutils.Hash) (int64, error) {
	return s.consensusState.GetMintCnt(timestamp, miner)
}

func (s *states) PutMintCnt(timestamp int64, miner byteutils.Hash, cnt int64) error {
	return s.consensusState.PutMintCnt(timestamp, miner, cnt)
}

func (s *states) HasCandidate(candidate byteutils.Hash) (bool, error) {
	return s.consensusState.HasCandidate(candidate)
}

func (s *states) AddCandidate(candidate byteutils.Hash) error {
	return s.consensusState.AddCandidate(candidate)
}

func (s *states) DelCandidate(candidate byteutils.Hash) error {
	return s.consensusState.DelCandidate(candidate)
}

func (s *states) GetVote(voter byteutils.Hash) (byteutils.Hash, error) {
	return s.consensusState.GetVote(voter)
}

func (s *states) AddVote(voter byteutils.Hash, votee byteutils.Hash) error {
	return s.consensusState.AddVote(voter, votee)
}

func (s *states) DelVote(voter byteutils.Hash) error {
	return s.consensusState.DelVote(voter)
}

func (s *states) IterVote() (Iterator, error) {
	return s.consensusState.IterVote()
}

func (s *states) AddDelegate(delegator byteutils.Hash, delegatee byteutils.Hash) error {
	return s.consensusState.AddDelegate(delegator, delegatee)
}

func (s *states) HasDelegate(delegator byteutils.Hash, delegatee byteutils.Hash) (bool, error) {
	return s.consensusState.HasDelegate(delegator, delegatee)
}

func (s *states) DelDelegate(delegator byteutils.Hash, delegatee byteutils.Hash) error {
	return s.consensusState.DelDelegate(delegator, delegatee)
}

func (s *states) IterDelegate(delgatee byteutils.Hash) (Iterator, error) {
	return s.consensusState.IterDelegate(delgatee)
}

func (s *states) Candidates() ([]byteutils.Hash, error) {
	return s.consensusState.Candidates()
}

func (s *states) CandidatesRoot() byteutils.Hash {
	return s.consensusState.CandidatesRoot()
}

func (s *states) Dynasty() ([]byteutils.Hash, error) {
	return s.consensusState.Dynasty()
}

func (s *states) NextDynasty() ([]byteutils.Hash, error) {
	return s.consensusState.NextDynasty()
}

func (s *states) DynastyRoot() byteutils.Hash {
	return s.consensusState.DynastyRoot()
}

func (s *states) NextDynastyRoot() byteutils.Hash {
	return s.consensusState.NextDynastyRoot()
}

func (s *states) Accounts() ([]Account, error) {
	return s.accState.Accounts()
}

// WorldState manange all current states in Blockchain
type worldState struct {
	*states

	txStates map[string]*txState
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
		mvccdb: mvccdb,
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
	txsState, err := trie.NewBatchTrie(root, ws.mvccdb)
	if err != nil {
		return err
	}
	ws.txsState = txsState
	return nil
}

func (ws *worldState) LoadEventsRoot(root byteutils.Hash) error {
	eventsState, err := trie.NewBatchTrie(root, ws.mvccdb)
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

type txState struct {
	*states

	txid string
	db   *mvccdb.DB
}

func (ws *worldState) Prepare(txid string) (TxState, error) {
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
	txState := &txState{
		db:     db,
		txid:   txid,
		states: states,
	}
	ws.txStates[txid] = txState
	return txState, nil
}

func (ws *worldState) Update(txid string) error {
	if _, ok := ws.txStates[txid]; ok {
		return ErrCannotUpdateTxStateBeforePrepare
	}
	if err := ws.mvccdb.Update(txid); err != nil {
		return err
	}
	states, err := ws.txStates[txid].states.ReplaceDB(ws.mvccdb)
	if err != nil {
		return err
	}
	ws.states = states
	return nil
}

func (ws *worldState) Check(txid string) (bool, error) {
	return false, nil
}

func (ts *txState) TxID() string {
	return ts.txid
}
