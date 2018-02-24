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
	"encoding/json"
	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/common/trie"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/core/state"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util/byteutils"
)

// WorldStateHeader is the header of WorldState
type WorldStateHeader struct {
	accStateRoot       byteutils.Hash
	txsStateRoot       byteutils.Hash
	eventsStateRoot    byteutils.Hash
	consensusStateRoot *corepb.DposContext
}

// WorldState manange all current states in Blockchain
type WorldState struct {
	accState       state.AccountState
	txsState       *trie.BatchTrie
	eventsState    *trie.BatchTrie
	consensusState *DposContext
}

// NewWorldState create a new empty WorldState
func NewWorldState(storage storage.Storage) (*WorldState, error) {
	accState, err := state.NewAccountState(nil, storage)
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
	consensusState, err := NewDposContext(storage)
	if err != nil {
		return nil, err
	}
	return &WorldState{
		accState:       accState,
		txsState:       txsState,
		eventsState:    eventsState,
		consensusState: consensusState,
	}, nil
}

func NewWorldStateFromHeader(storage storage.Storage, wsh *WorldStateHeader) (*WorldState, error) {
	accState, err := state.NewAccountState(wsh.accStateRoot, storage)
	if err != nil {
		return nil, err
	}
	txsState, err := trie.NewBatchTrie(wsh.txsStateRoot, storage)
	if err != nil {
		return nil, err
	}
	eventsState, err := trie.NewBatchTrie(wsh.eventsStateRoot, storage)
	if err != nil {
		return nil, err
	}
	consensusState, err := NewDposContext(storage)
	if err != nil {
		return nil, err
	}
	if err := consensusState.FromProto(wsh.consensusStateRoot); err != nil {
		return nil, err
	}
	return &WorldState{
		accState:       accState,
		txsState:       txsState,
		eventsState:    eventsState,
		consensusState: consensusState,
	}, nil
}

// Clone a new WorldState
func (ws *WorldState) Clone() (*WorldState, error) {
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
	return &WorldState{
		accState:       accState,
		txsState:       txsState,
		eventsState:    eventsState,
		consensusState: consensusState,
	}, nil
}

func (ws *WorldState) Begin() {
	ws.accState.BeginBatch()
	ws.txsState.BeginBatch()
	ws.eventsState.BeginBatch()
	ws.consensusState.BeginBatch()
}

func (ws *WorldState) Commit() {
	ws.accState.Commit()
	ws.txsState.Commit()
	ws.eventsState.Commit()
	ws.consensusState.Commit()
}

func (ws *WorldState) RollBack() {
	ws.accState.RollBack()
	ws.txsState.RollBack()
	ws.eventsState.RollBack()
	ws.consensusState.RollBack()
}

func (ws *WorldState) ToHeaders() (*WorldStateHeader, error) {
	accStateRoot, err := ws.accState.RootHash()
	if err != nil {
		return nil, err
	}
	consensusStateRoot, err := ws.consensusState.ToProto()
	if err != nil {
		return nil, err
	}
	return &WorldStateHeader{
		accStateRoot:       accStateRoot,
		txsStateRoot:       ws.txsState.RootHash(),
		eventsStateRoot:    ws.eventsState.RootHash(),
		consensusStateRoot: consensusStateRoot,
	}, nil
}

func (ws *WorldState) GetOrCreateUserAccount(addr byteutils.Hash) (state.Account, error) {
	return ws.accState.GetOrCreateUserAccount(addr)
}

func (ws *WorldState) GetContractAccount(addr byteutils.Hash) (state.Account, error) {
	return ws.accState.GetContractAccount(addr)
}

func (ws *WorldState) CreateContractAccount(owner byteutils.Hash, birthPlace byteutils.Hash) (state.Account, error) {
	return ws.accState.CreateContractAccount(owner, birthPlace)
}

func (ws *WorldState) GetTx(txHash byteutils.Hash) (*Transaction, error) {
	txBytes, err := ws.txsState.Get(txHash)
	if err != nil {
		return nil, err
	}
	pbTx := new(corepb.Transaction)
	if err := proto.Unmarshal(txBytes, pbTx); err != nil {
		return nil, err
	}

	tx := new(Transaction)
	if err = tx.FromProto(pbTx); err != nil {
		return nil, err
	}
	return tx, nil
}

func (ws *WorldState) PutTx(tx *Transaction) error {
	pbTx, err := tx.ToProto()
	if err != nil {
		return err
	}
	txBytes, err := proto.Marshal(pbTx)
	if err != nil {
		return err
	}
	if _, err := ws.txsState.Put(tx.hash, txBytes); err != nil {
		return err
	}
	return nil
}

func (ws *WorldState) RecordEvent(txHash byteutils.Hash, event *Event) error {
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

func (ws *WorldState) FetchEvents(txHash byteutils.Hash) ([]*Event, error) {
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

func (ws *WorldState) GetMintCnt(dynasty int64, miner *Address) (int64, error) {
	key := append(byteutils.FromInt64(dynasty), miner.Bytes()...)
	bytes, err := ws.consensusState.mintCntTrie.Get(key)
	if err != nil && err != storage.ErrKeyNotFound {
		return 0, err
	}
	cnt := int64(0)
	if err != storage.ErrKeyNotFound {
		cnt = byteutils.Int64(bytes)
	}
	return cnt, nil
}

func (ws *WorldState) PutMintCnt(dynasty int64, miner *Address, cnt int64) error {
	key := append(byteutils.FromInt64(dynasty), miner.Bytes()...)
	_, err := ws.consensusState.mintCntTrie.Put(key, byteutils.FromInt64(cnt))
	if err != nil {
		return err
	}
	return nil
}

func (ws *WorldState) GetCandidate(candidate byteutils.Hash) ([]byte, error) {
	return ws.consensusState.candidateTrie.Get(candidate)
}

func (ws *WorldState) AddCandidate(candidate byteutils.Hash) error {
	_, err := ws.consensusState.candidateTrie.Put(candidate, candidate)
	return err
}

func (ws *WorldState) DelCandidate(candidate byteutils.Hash) error {
	return ws.consensusState.kickoutCandidate(candidate)
}

func (ws *WorldState) GetVote(voter byteutils.Hash) ([]byte, error) {
	return ws.consensusState.voteTrie.Get(voter)
}

func (ws *WorldState) AddVote(voter byteutils.Hash, votee byteutils.Hash) error {
	_, err := ws.consensusState.voteTrie.Put(voter, votee)
	return err
}

func (ws *WorldState) DelVote(voter byteutils.Hash) error {
	_, err := ws.consensusState.voteTrie.Del(voter)
	return err
}

func (ws *WorldState) IterVote(voter byteutils.Hash) (*trie.Iterator, error) {
	return ws.consensusState.voteTrie.Iterator(voter)
}

func (ws *WorldState) AddDelegate(delegator byteutils.Hash, delegatee byteutils.Hash) error {
	key := append(delegatee, delegator...)
	_, err := ws.consensusState.delegateTrie.Put(key, delegator)
	return err
}

func (ws *WorldState) GetDelegate(delegator byteutils.Hash, delegatee byteutils.Hash) ([]byte, error) {
	key := append(delegatee, delegator...)
	return ws.consensusState.delegateTrie.Get(key)
}

func (ws *WorldState) DelDelegate(delegator byteutils.Hash, delegatee byteutils.Hash) error {
	key := append(delegatee, delegator...)
	_, err := ws.consensusState.delegateTrie.Del(key)
	return err
}

func (ws *WorldState) IterDelegate(delgatee byteutils.Hash) (*trie.Iterator, error) {
	return ws.consensusState.delegateTrie.Iterator(delgatee)
}

func (ws *WorldState) accounts() ([]state.Account, error) {
	return ws.accState.Accounts()
}

func (ws *WorldState) candidates() ([]byteutils.Hash, error) {
	return TraverseDynasty(ws.consensusState.candidateTrie)
}

func (ws *WorldState) dynasty() ([]byteutils.Hash, error) {
	return TraverseDynasty(ws.consensusState.dynastyTrie)
}

func (ws *WorldState) nxtDynasty() ([]byteutils.Hash, error) {
	return TraverseDynasty(ws.consensusState.nextDynastyTrie)
}

func (ws *WorldState) loadConsensusContext(context *DynastyContext) error {
	dynastyTrie, err := context.DynastyTrie.Clone()
	if err != nil {
		return err
	}
	nextDynastyTrie, err := context.NextDynastyTrie.Clone()
	if err != nil {
		return err
	}
	delegateTrie, err := context.DelegateTrie.Clone()
	if err != nil {
		return err
	}
	candidateTrie, err := context.CandidateTrie.Clone()
	if err != nil {
		return err
	}
	voteTrie, err := context.VoteTrie.Clone()
	if err != nil {
		return err
	}
	mintCntTrie, err := context.MintCntTrie.Clone()
	if err != nil {
		return err
	}
	ws.consensusState = &DposContext{
		dynastyTrie:     dynastyTrie,
		nextDynastyTrie: nextDynastyTrie,
		delegateTrie:    delegateTrie,
		candidateTrie:   candidateTrie,
		voteTrie:        voteTrie,
		mintCntTrie:     mintCntTrie,
		storage:         context.Storage,
	}
	return nil
}
