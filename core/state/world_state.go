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

	"github.com/nebulasio/go-nebulas/common/trie"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util/byteutils"
)

// WorldState manange all current states in Blockchain
type worldState struct {
	accState       AccountState
	txsState       *trie.BatchTrie
	eventsState    *trie.BatchTrie
	consensusState ConsensusState

	storage   storage.Storage
	consensus Consensus
}

// NewWorldState create a new empty WorldState
func NewWorldState(consensus Consensus, storage storage.Storage) (WorldState, error) {
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
	consensusState, err := consensus.NewState(nil, storage)
	if err != nil {
		return nil, err
	}
	return &worldState{
		accState:       accState,
		txsState:       txsState,
		eventsState:    eventsState,
		consensusState: consensusState,

		storage:   storage,
		consensus: consensus,
	}, nil
}

func (ws *worldState) LoadAccountsRoot(root byteutils.Hash) error {
	accState, err := NewAccountState(root, ws.storage)
	if err != nil {
		return err
	}
	ws.accState = accState
	return nil
}

func (ws *worldState) LoadTxsRoot(root byteutils.Hash) error {
	txsState, err := trie.NewBatchTrie(root, ws.storage)
	if err != nil {
		return err
	}
	ws.txsState = txsState
	return nil
}

func (ws *worldState) LoadEventsRoot(root byteutils.Hash) error {
	eventsState, err := trie.NewBatchTrie(root, ws.storage)
	if err != nil {
		return err
	}
	ws.eventsState = eventsState
	return nil
}

func (ws *worldState) LoadConsensusRoot(root byteutils.Hash) error {
	consensusState, err := ws.consensus.NewState(root, ws.storage)
	if err != nil {
		return err
	}
	ws.consensusState = consensusState
	return nil
}

// Clone a new WorldState
func (ws *worldState) Clone() (WorldState, error) {
	accState, err := ws.accState.Clone()
	if err != nil {
		return nil, err
	}
	txsState, err := ws.txsState.Clone()
	if err != nil {
		return nil, err
	}
	eventsState, err := ws.eventsState.Clone()
	if err != nil {
		return nil, err
	}
	consensusState, err := ws.consensusState.Clone()
	if err != nil {
		return nil, err
	}
	return &worldState{
		accState:       accState,
		txsState:       txsState,
		eventsState:    eventsState,
		consensusState: consensusState,
	}, nil
}

func (ws *worldState) Begin() {
	ws.accState.BeginBatch()
	ws.txsState.BeginBatch()
	ws.eventsState.BeginBatch()
	ws.consensusState.BeginBatch()
}

func (ws *worldState) Commit() {
	ws.accState.Commit()
	ws.txsState.Commit()
	ws.eventsState.Commit()
	ws.consensusState.Commit()
}

func (ws *worldState) RollBack() {
	ws.accState.RollBack()
	ws.txsState.RollBack()
	ws.eventsState.RollBack()
	ws.consensusState.RollBack()
}

func (ws *worldState) AccountsRoot() (byteutils.Hash, error) {
	return ws.accState.RootHash()
}

func (ws *worldState) TxsRoot() (byteutils.Hash, error) {
	return ws.txsState.RootHash(), nil
}

func (ws *worldState) EventsRoot() (byteutils.Hash, error) {
	return ws.eventsState.RootHash(), nil
}

func (ws *worldState) ConsensusRoot() (byteutils.Hash, error) {
	return ws.consensusState.RootHash()
}

func (ws *worldState) GetOrCreateUserAccount(addr byteutils.Hash) (Account, error) {
	return ws.accState.GetOrCreateUserAccount(addr)
}

func (ws *worldState) GetContractAccount(addr byteutils.Hash) (Account, error) {
	return ws.accState.GetContractAccount(addr)
}

func (ws *worldState) CreateContractAccount(owner byteutils.Hash, birthPlace byteutils.Hash) (Account, error) {
	return ws.accState.CreateContractAccount(owner, birthPlace)
}

func (ws *worldState) GetTx(txHash byteutils.Hash) ([]byte, error) {
	return ws.txsState.Get(txHash)
}

func (ws *worldState) PutTx(txHash byteutils.Hash, txBytes []byte) error {
	_, err := ws.txsState.Put(txHash, txBytes)
	return err
}

func (ws *worldState) RecordEvent(txHash byteutils.Hash, event *Event) error {
	iter, err := ws.eventsState.Iterator(txHash)
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
	_, err = ws.eventsState.Put(key, bytes)
	if err != nil {
		return err
	}
	return nil
}

func (ws *worldState) FetchEvents(txHash byteutils.Hash) ([]*Event, error) {
	events := []*Event{}
	iter, err := ws.eventsState.Iterator(txHash)
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

func (ws *worldState) NextConsensusState(elapsedSecond int64) (ConsensusState, error) {
	return ws.consensusState.NextConsensusState(elapsedSecond, ws)
}

func (ws *worldState) GetMintCnt(timestamp int64, miner byteutils.Hash) (int64, error) {
	return ws.consensusState.GetMintCnt(timestamp, miner)
}

func (ws *worldState) PutMintCnt(timestamp int64, miner byteutils.Hash, cnt int64) error {
	return ws.consensusState.PutMintCnt(timestamp, miner, cnt)
}

func (ws *worldState) GetCandidate(candidate byteutils.Hash) (byteutils.Hash, error) {
	return ws.consensusState.GetCandidate(candidate)
}

func (ws *worldState) AddCandidate(candidate byteutils.Hash) error {
	return ws.consensusState.AddCandidate(candidate)
}

func (ws *worldState) DelCandidate(candidate byteutils.Hash) error {
	return ws.consensusState.DelCandidate(candidate)
}

func (ws *worldState) GetVote(voter byteutils.Hash) (byteutils.Hash, error) {
	return ws.consensusState.GetVote(voter)
}

func (ws *worldState) AddVote(voter byteutils.Hash, votee byteutils.Hash) error {
	return ws.consensusState.AddVote(voter, votee)
}

func (ws *worldState) DelVote(voter byteutils.Hash) error {
	return ws.consensusState.DelVote(voter)
}

func (ws *worldState) IterVote() (Iterator, error) {
	return ws.consensusState.IterVote()
}

func (ws *worldState) AddDelegate(delegator byteutils.Hash, delegatee byteutils.Hash) error {
	return ws.consensusState.AddDelegate(delegator, delegatee)
}

func (ws *worldState) HasDelegate(delegator byteutils.Hash, delegatee byteutils.Hash) (bool, error) {
	return ws.consensusState.HasDelegate(delegator, delegatee)
}

func (ws *worldState) DelDelegate(delegator byteutils.Hash, delegatee byteutils.Hash) error {
	return ws.consensusState.DelDelegate(delegator, delegatee)
}

func (ws *worldState) IterDelegate(delgatee byteutils.Hash) (Iterator, error) {
	return ws.consensusState.IterDelegate(delgatee)
}

func (ws *worldState) Candidates() ([]byteutils.Hash, error) {
	return ws.consensusState.Candidates()
}

func (ws *worldState) Dynasty() ([]byteutils.Hash, error) {
	return ws.consensusState.Dynasty()
}

func (ws *worldState) NextDynasty() ([]byteutils.Hash, error) {
	return ws.consensusState.NextDynasty()
}

func (ws *worldState) DynastyRoot() byteutils.Hash {
	return ws.consensusState.DynastyRoot()
}

func (ws *worldState) NextDynastyRoot() byteutils.Hash {
	return ws.consensusState.NextDynastyRoot()
}

func (ws *worldState) Accounts() ([]Account, error) {
	return ws.accState.Accounts()
}
