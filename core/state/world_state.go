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

	consensuspb "github.com/nebulasio/go-nebulas/consensus/pb"
	corepb "github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/util"

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
	db, err := mvccdb.NewMVCCDB(mem, false)
	if err != nil {
		return nil, err
	}

	db.SetStrictGlobalVersionCheck(true)
	return db, nil
}

func newStateDB(storage storage.Storage) (*mvccdb.MVCCDB, error) {
	return mvccdb.NewMVCCDB(storage, true)
}

type states struct {
	accState       AccountState
	txsState       *trie.Trie
	eventsState    *trie.Trie
	consensusState ConsensusState

	consensus Consensus
	changelog *mvccdb.MVCCDB
	stateDB   *mvccdb.MVCCDB
	innerDB   storage.Storage
	txid      interface{}

	gasConsumed map[string]*util.Uint128
	events      map[string][]*Event
}

func newStates(consensus Consensus, stor storage.Storage) (*states, error) {
	changelog, err := newChangeLog()
	if err != nil {
		return nil, err
	}
	stateDB, err := newStateDB(stor)
	if err != nil {
		return nil, err
	}

	accState, err := NewAccountState(nil, stateDB)
	if err != nil {
		return nil, err
	}
	txsState, err := trie.NewTrie(nil, stateDB, false)
	if err != nil {
		return nil, err
	}
	eventsState, err := trie.NewTrie(nil, stateDB, false)
	if err != nil {
		return nil, err
	}
	consensusState, err := consensus.NewState(&consensuspb.ConsensusRoot{}, stateDB, false)
	if err != nil {
		return nil, err
	}

	return &states{
		accState:       accState,
		txsState:       txsState,
		eventsState:    eventsState,
		consensusState: consensusState,

		consensus: consensus,
		changelog: changelog,
		stateDB:   stateDB,
		innerDB:   stor,
		txid:      nil,

		gasConsumed: make(map[string]*util.Uint128),
		events:      make(map[string][]*Event),
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
	err = s.ReplayEvent(done)
	if err != nil {
		return err
	}
	err = s.consensusState.Replay(done.consensusState)
	if err != nil {
		return err
	}

	// replay gasconsumed
	for from, gas := range done.gasConsumed {
		consumed, ok := s.gasConsumed[from]
		if !ok {
			consumed = util.NewUint128()
		}
		consumed, err := consumed.Add(gas)
		if err != nil {
			return err
		}
		s.gasConsumed[from] = consumed
	}
	return nil
}

func (s *states) ReplayEvent(done *states) error {

	tx := done.txid.(string)
	events, ok := done.events[tx]
	if !ok {
		return nil
	}

	//replay event
	txHash, err := byteutils.FromHex(tx)
	if err != nil {
		return err
	}
	for idx, event := range events {
		cnt := int64(idx + 1)

		key := append(txHash, byteutils.FromInt64(cnt)...)
		bytes, err := json.Marshal(event)
		if err != nil {
			return err
		}

		_, err = s.eventsState.Put(key, bytes)
		if err != nil {
			return err
		}
	}
	done.events = make(map[string][]*Event)
	return nil
}

func (s *states) Clone() (*states, error) {
	changelog, err := newChangeLog()
	if err != nil {
		return nil, err
	}
	stateDB, err := newStateDB(s.innerDB)
	if err != nil {
		return nil, err
	}

	accState, err := NewAccountState(s.accState.RootHash(), stateDB)
	if err != nil {
		return nil, err
	}
	txsState, err := trie.NewTrie(s.txsState.RootHash(), stateDB, false)
	if err != nil {
		return nil, err
	}
	eventsState, err := trie.NewTrie(s.eventsState.RootHash(), stateDB, false)
	if err != nil {
		return nil, err
	}
	consensusState, err := s.consensus.NewState(s.consensusState.RootHash(), stateDB, false)
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
		stateDB:   stateDB,
		innerDB:   s.innerDB,
		txid:      s.txid,

		gasConsumed: make(map[string]*util.Uint128),
		events:      make(map[string][]*Event),
	}, nil
}

func (s *states) Begin() error {
	if err := s.changelog.Begin(); err != nil {
		return err
	}
	if err := s.stateDB.Begin(); err != nil {
		return err
	}
	return nil
}

func (s *states) Commit() error {
	if err := s.Flush(); err != nil {
		return err
	}
	// changelog is used to check conflict temporarily
	// we should rollback it when the transaction is over
	if err := s.changelog.RollBack(); err != nil {
		return err
	}
	if err := s.stateDB.Commit(); err != nil {
		return err
	}

	s.events = make(map[string][]*Event)
	s.gasConsumed = make(map[string]*util.Uint128)
	return nil
}

func (s *states) RollBack() error {
	if err := s.Abort(); err != nil {
		return err
	}
	if err := s.changelog.RollBack(); err != nil {
		return err
	}
	if err := s.stateDB.RollBack(); err != nil {
		return err
	}

	s.events = make(map[string][]*Event)
	s.gasConsumed = make(map[string]*util.Uint128)
	return nil
}

func (s *states) Prepare(txid interface{}) (*states, error) {
	changelog, err := s.changelog.Prepare(txid)
	if err != nil {
		return nil, err
	}
	stateDB, err := s.stateDB.Prepare(txid)
	if err != nil {
		return nil, err
	}

	// Flush all changes in world state into merkle trie
	// make a snapshot of world state
	if err := s.Flush(); err != nil {
		return nil, err
	}

	accState, err := NewAccountState(s.AccountsRoot(), stateDB)
	if err != nil {
		return nil, err
	}
	txsState, err := trie.NewTrie(s.TxsRoot(), stateDB, true)
	if err != nil {
		return nil, err
	}
	eventsState, err := trie.NewTrie(s.EventsRoot(), stateDB, true)
	if err != nil {
		return nil, err
	}
	consensusState, err := s.consensus.NewState(s.ConsensusRoot(), stateDB, true)
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
		stateDB:   stateDB,
		innerDB:   s.innerDB,
		txid:      txid,

		gasConsumed: make(map[string]*util.Uint128),
		events:      make(map[string][]*Event),
	}, nil
}

func (s *states) CheckAndUpdateTo(parent *states) ([]interface{}, error) {
	dependency, err := s.changelog.CheckAndUpdate()
	if err != nil {
		return nil, err
	}
	_, err = s.stateDB.CheckAndUpdate()
	if err != nil {
		return nil, err
	}
	if err := parent.Replay(s); err != nil {
		return nil, err
	}
	return dependency, nil
}

func (s *states) Reset(addr byteutils.Hash, isResetChangeLog bool) error {

	if err := s.stateDB.Reset(); err != nil {
		return err
	}
	if err := s.Abort(); err != nil {
		return err
	}

	if isResetChangeLog {
		if err := s.changelog.Reset(); err != nil {
			return err
		}
		if addr != nil {
			// record dependency
			if err := s.changelog.Put(addr, addr); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *states) Close() error {
	if err := s.changelog.Close(); err != nil {
		return err
	}
	if err := s.stateDB.Close(); err != nil {
		return err
	}
	if err := s.Abort(); err != nil {
		return err
	}
	return nil
}

func (s *states) AccountsRoot() byteutils.Hash {
	return s.accState.RootHash()
}

func (s *states) TxsRoot() byteutils.Hash {
	return s.txsState.RootHash()
}

func (s *states) EventsRoot() byteutils.Hash {
	return s.eventsState.RootHash()
}

func (s *states) ConsensusRoot() *consensuspb.ConsensusRoot {
	return s.consensusState.RootHash()
}

func (s *states) Flush() error {
	return s.accState.Flush()
}

func (s *states) Abort() error {
	// TODO: Abort txsState, eventsState, consensusState
	// we don't need to abort the three states now
	// because we only use abort in reset, close and rollback
	// in close & rollback, we won't use states any more
	// in reset, we won't change the three states before we reset them
	return s.accState.Abort()
}

func (s *states) recordAccount(acc Account) (Account, error) {
	if err := s.changelog.Put(acc.Address(), acc.Address()); err != nil {
		return nil, err
	}
	return acc, nil
}

func (s *states) GetOrCreateUserAccount(addr byteutils.Hash) (Account, error) {
	acc, err := s.accState.GetOrCreateUserAccount(addr)
	if err != nil {
		return nil, err
	}
	return s.recordAccount(acc)
}

func (s *states) GetContractAccount(addr byteutils.Hash) (Account, error) {
	acc, err := s.accState.GetContractAccount(addr)
	if err != nil {
		return nil, err
	}
	if len(acc.BirthPlace()) == 0 {
		return nil, ErrContractCheckFailed
	}
	return s.recordAccount(acc)
}

func (s *states) CreateContractAccount(owner byteutils.Hash, birthPlace byteutils.Hash, contractMeta *corepb.ContractMeta) (Account, error) {
	acc, err := s.accState.CreateContractAccount(owner, birthPlace, contractMeta)
	if err != nil {
		return nil, err
	}
	return s.recordAccount(acc)
}

func (s *states) GetTx(txHash byteutils.Hash) ([]byte, error) {
	bytes, err := s.txsState.Get(txHash)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (s *states) PutTx(txHash byteutils.Hash, txBytes []byte) error {
	_, err := s.txsState.Put(txHash, txBytes)
	if err != nil {
		return err
	}
	return nil
}

func (s *states) RecordEvent(txHash byteutils.Hash, event *Event) {
	events, ok := s.events[txHash.String()]
	if !ok {
		events = make([]*Event, 0)
	}
	s.events[txHash.String()] = append(events, event)
}

func (s *states) FetchEvents(txHash byteutils.Hash) ([]*Event, error) {
	events := []*Event{}
	iter, err := s.eventsState.Iterator(txHash)
	if err != nil && err != storage.ErrKeyNotFound {
		return nil, err
	}
	if err == nil {
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

func (s *states) Accounts() ([]Account, error) { // TODO delete
	return s.accState.Accounts()
}

func (s *states) LoadAccountsRoot(root byteutils.Hash) error {
	accState, err := NewAccountState(root, s.stateDB)
	if err != nil {
		return err
	}
	s.accState = accState
	return nil
}

func (s *states) LoadTxsRoot(root byteutils.Hash) error {
	txsState, err := trie.NewTrie(root, s.stateDB, false)
	if err != nil {
		return err
	}
	s.txsState = txsState
	return nil
}

func (s *states) LoadEventsRoot(root byteutils.Hash) error {
	eventsState, err := trie.NewTrie(root, s.stateDB, false)
	if err != nil {
		return err
	}
	s.eventsState = eventsState
	return nil
}

func (s *states) LoadConsensusRoot(root *consensuspb.ConsensusRoot) error {
	consensusState, err := s.consensus.NewState(root, s.stateDB, false)
	if err != nil {
		return err
	}
	s.consensusState = consensusState
	return nil
}

func (s *states) RecordGas(from string, gas *util.Uint128) error {
	consumed, ok := s.gasConsumed[from]
	if !ok {
		consumed = util.NewUint128()
	}
	consumed, err := consumed.Add(gas)
	if err != nil {
		return err
	}
	s.gasConsumed[from] = consumed
	return nil
}

func (s *states) GetGas() map[string]*util.Uint128 {
	gasConsumed := make(map[string]*util.Uint128)
	for from, gas := range s.gasConsumed {
		gasConsumed[from] = gas
	}
	s.gasConsumed = make(map[string]*util.Uint128)
	return gasConsumed
}

func (s *states) GetBlockHashByHeight(height uint64) ([]byte, error) {
	bytes, err := s.innerDB.Get(byteutils.FromUint64(height))
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (s *states) GetBlock(hash byteutils.Hash) ([]byte, error) {
	bytes, err := s.innerDB.Get(hash)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// WorldState manange all current states in Blockchain
type worldState struct {
	*states
	snapshot *states
}

// NewWorldState create a new empty WorldState
func NewWorldState(consensus Consensus, storage storage.Storage) (WorldState, error) {
	states, err := newStates(consensus, storage)
	if err != nil {
		return nil, err
	}
	return &worldState{
		states:   states,
		snapshot: nil,
	}, nil
}

// Clone a new WorldState
func (ws *worldState) Clone() (WorldState, error) {
	s, err := ws.states.Clone()
	if err != nil {
		return nil, err
	}
	return &worldState{
		states:   s,
		snapshot: nil,
	}, nil
}

func (ws *worldState) Begin() error {
	snapshot, err := ws.states.Clone()
	if err != nil {
		return err
	}
	if err := ws.states.Begin(); err != nil {
		return err
	}
	ws.snapshot = snapshot
	return nil
}

func (ws *worldState) Commit() error {
	if err := ws.states.Commit(); err != nil {
		return err
	}
	ws.snapshot = nil
	return nil
}

func (ws *worldState) RollBack() error {
	if err := ws.states.RollBack(); err != nil {
		return err
	}
	ws.states = ws.snapshot
	ws.snapshot = nil
	return nil
}

func (ws *worldState) Prepare(txid interface{}) (TxWorldState, error) {
	s, err := ws.states.Prepare(txid)
	if err != nil {
		return nil, err
	}
	txState := &txWorldState{
		states: s,
		txid:   txid,
		parent: ws,
	}
	return txState, nil
}

func (ws *worldState) NextConsensusState(elapsedSecond int64) (ConsensusState, error) {
	return ws.states.consensusState.NextConsensusState(elapsedSecond, ws)
}

func (ws *worldState) SetConsensusState(consensusState ConsensusState) {
	ws.states.consensusState = consensusState
}

type txWorldState struct {
	*states
	txid   interface{}
	parent *worldState
}

func (tws *txWorldState) CheckAndUpdate() ([]interface{}, error) {
	dependencies, err := tws.states.CheckAndUpdateTo(tws.parent.states)
	if err != nil {
		return nil, err
	}
	tws.parent = nil
	return dependencies, nil
}

func (tws *txWorldState) Reset(addr byteutils.Hash, isResetChangeLog bool) error {
	if err := tws.states.Reset(addr, isResetChangeLog); err != nil {
		return err
	}
	return nil
}

func (tws *txWorldState) Close() error {
	if err := tws.states.Close(); err != nil {
		return err
	}
	tws.parent = nil
	return nil
}

func (tws *txWorldState) TxID() interface{} {
	return tws.txid
}
