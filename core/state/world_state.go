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

	"github.com/nebulasio/go-nebulas/util/logging"

	"github.com/nebulasio/go-nebulas/consensus/pb"
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

func newStorage(storage storage.Storage) (*mvccdb.MVCCDB, error) { // TODO rename NewStateDB
	return mvccdb.NewMVCCDB(storage, true)
}

type states struct {
	accState       AccountState
	txsState       *trie.Trie
	eventsState    *trie.Trie
	consensusState ConsensusState

	consensus Consensus
	changelog *mvccdb.MVCCDB
	storage   *mvccdb.MVCCDB  // TODO renmae stateDB
	db        storage.Storage // TODO rename interStorage
	txid      interface{}

	gasConsumed map[string]*util.Uint128
	events      map[string][]*Event
}

func newStates(consensus Consensus, stor storage.Storage) (*states, error) {
	// logging.CLog().Info("WS New MVCCDB: ", nil)
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
	txsState, err := trie.NewTrie(nil, storage, false)
	if err != nil {
		return nil, err
	}
	eventsState, err := trie.NewTrie(nil, storage, false)
	if err != nil {
		return nil, err
	}
	consensusState, err := consensus.NewState(&consensuspb.ConsensusRoot{}, storage, false)
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
		storage:   storage,
		db:        stor,
		txid:      nil,

		gasConsumed: make(map[string]*util.Uint128),
		events:      make(map[string][]*Event),
	}, nil
}

func (s *states) Replay(done *states) error {
	// logging.CLog().Info("WS Replay MVCCDB: ", s.txid)
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
		var err error
		s.gasConsumed[from], err = consumed.Add(gas) // TODO use tmp var, assign if success
		if err != nil {
			return err
		}
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
	// logging.CLog().Info("WS Clone MVCCDB: ", s.txid)
	changelog, err := newChangeLog()
	if err != nil {
		logging.CLog().Info("CE 1")
		return nil, err
	}
	storage, err := newStorage(s.db)
	if err != nil {
		logging.CLog().Info("CE 2")
		return nil, err
	}

	accState, err := NewAccountState(s.accState.RootHash(), storage)
	if err != nil {
		logging.CLog().Info("CE 3")
		return nil, err
	}
	txsState, err := trie.NewTrie(s.txsState.RootHash(), storage, false)
	if err != nil {
		logging.CLog().Info("CE 4")
		return nil, err
	}
	eventsState, err := trie.NewTrie(s.eventsState.RootHash(), storage, false)
	if err != nil {
		logging.CLog().Info("CE 5")
		return nil, err
	}
	consensusState, err := s.consensus.NewState(s.consensusState.RootHash(), storage, false)
	if err != nil {
		logging.CLog().Info("CE 6")
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
		db:        s.db,
		txid:      s.txid,

		gasConsumed: make(map[string]*util.Uint128),
		events:      make(map[string][]*Event),
	}, nil
}

func (s *states) Begin() error {
	// logging.CLog().Info("WS Begin MVCCDB: ", s.txid)
	if err := s.changelog.Begin(); err != nil {
		logging.CLog().Info("BE 11")
		return err
	}
	if err := s.storage.Begin(); err != nil {
		logging.CLog().Info("BE 12")
		return err
	}
	return nil
}

func (s *states) Commit() error {
	// logging.CLog().Info("WS Commit MVCCDB: ", s.txid)
	if err := s.Flush(); err != nil {
		return err
	}
	if err := s.changelog.RollBack(); err != nil { // TODO add comment
		return err
	}
	if err := s.storage.Commit(); err != nil {
		return err
	}

	s.events = make(map[string][]*Event)
	s.gasConsumed = make(map[string]*util.Uint128)
	return nil
}

func (s *states) RollBack() error {
	// logging.CLog().Info("WS Rollback MVCCDB: ", s.txid)
	if err := s.Abort(); err != nil {
		return err
	}
	if err := s.changelog.RollBack(); err != nil {
		return err
	}
	if err := s.storage.RollBack(); err != nil {
		return err
	}

	s.events = make(map[string][]*Event)
	s.gasConsumed = make(map[string]*util.Uint128)
	return nil
}

func (s *states) Prepare(txid interface{}) (*states, error) {
	// logging.CLog().Info("WS Prepare MVCCDB: ", txid)
	changelog, err := s.changelog.Prepare(txid)
	if err != nil {
		logging.VLog().Info("PPE 11")
		return nil, err
	}
	storage, err := s.storage.Prepare(txid)
	if err != nil {
		logging.VLog().Info("PPE 12")
		return nil, err
	}

	if err := s.Flush(); err != nil { // TODO: add comment
		return nil, err
	}

	accState, err := NewAccountState(s.AccountsRoot(), storage)
	if err != nil {
		return nil, err
	}
	txsState, err := trie.NewTrie(s.TxsRoot(), storage, true)
	if err != nil {
		return nil, err
	}
	eventsState, err := trie.NewTrie(s.EventsRoot(), storage, true)
	if err != nil {
		return nil, err
	}
	consensusState, err := s.consensus.NewState(s.ConsensusRoot(), storage, true)
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
		db:        s.db,
		txid:      txid,

		gasConsumed: make(map[string]*util.Uint128),
		events:      make(map[string][]*Event),
	}, nil
}

func (s *states) CheckAndUpdateTo(parent *states) ([]interface{}, error) {
	// logging.CLog().Info("WS CheckAndUpdateTo MVCCDB: ", s.txid)
	dependency, err := s.changelog.CheckAndUpdate()
	if err != nil {
		logging.VLog().Info("CUE 11")
		return nil, err
	}
	_, err = s.storage.CheckAndUpdate() // TODO delete
	if err != nil {
		logging.VLog().Info("CUE 12")
		return nil, err
	}
	if err := parent.Replay(s); err != nil {
		return nil, err
	}
	return dependency, nil
}

func (s *states) Reset() error {
	// logging.CLog().Info("WS Reset MVCCDB: ", s.txid)
	if err := s.changelog.Reset(); err != nil {
		logging.VLog().Info("RSE 11")
		return err
	}
	if err := s.storage.Reset(); err != nil {
		logging.VLog().Info("RSE 12")
		return err
	}
	if err := s.Abort(); err != nil {
		return err
	}
	return nil
}

func (s *states) Close() error {
	// logging.CLog().Info("WS Close MVCCDB: ", s.txid)
	if err := s.changelog.Close(); err != nil {
		logging.VLog().Info("CSE 11")
		return err
	}
	if err := s.storage.Close(); err != nil {
		logging.VLog().Info("CSE 12")
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
	return s.accState.Abort() // TODO: Abort txsState, eventsState, consensusState // TODO: add comment, why success
}

func (s *states) recordAccount(acc Account) (Account, error) {
	bytes, err := acc.ToBytes()
	if err != nil {
		logging.VLog().Info("RAE 2")
		return nil, err
	}
	if err := s.changelog.Put(acc.Address(), bytes); err != nil {
		logging.VLog().Info("RAE 3")
		return nil, err
	}
	return acc, nil
}

func (s *states) GetOrCreateUserAccount(addr byteutils.Hash) (Account, error) {
	acc, err := s.accState.GetOrCreateUserAccount(addr)
	if err != nil {
		logging.VLog().Info("GCUE 1")
		return nil, err
	}
	return s.recordAccount(acc)
}

func (s *states) GetContractAccount(addr byteutils.Hash) (Account, error) {
	acc, err := s.accState.GetContractAccount(addr)
	if err != nil {
		logging.VLog().Info("GCE 1")
		return nil, err
	}
	return s.recordAccount(acc)
}

func (s *states) CreateContractAccount(owner byteutils.Hash, birthPlace byteutils.Hash) (Account, error) {
	acc, err := s.accState.CreateContractAccount(owner, birthPlace)
	if err != nil {
		logging.VLog().Info("CCE 1")
		return nil, err
	}
	return s.recordAccount(acc)
}

func (s *states) GetTx(txHash byteutils.Hash) ([]byte, error) {
	bytes, err := s.txsState.Get(txHash)
	if err != nil {
		logging.VLog().Info("GTE 11")
		return nil, err
	}
	// record change log
	if _, err := s.changelog.Get(txHash); err != nil && err != storage.ErrKeyNotFound {
		logging.VLog().Info("GTE 12")
		return nil, err
	}
	return bytes, nil
}

func (s *states) PutTx(txHash byteutils.Hash, txBytes []byte) error {
	_, err := s.txsState.Put(txHash, txBytes)
	if err != nil {
		logging.VLog().Info("PTE 11")
		return err
	}
	// record change log
	if err := s.changelog.Put(txHash, txBytes); err != nil {
		logging.VLog().Info("PTE 12")
		return err
	}
	return nil
}

func (s *states) RecordEvent(txHash byteutils.Hash, event *Event) error {
	events, ok := s.events[txHash.String()]
	if !ok {
		events = make([]*Event, 0)
	}

	cnt := int64(len(events) + 1)

	key := append(txHash, byteutils.FromInt64(cnt)...)
	bytes, err := json.Marshal(event)
	if err != nil {
		logging.VLog().Info("REE 11")
		return err
	}
	s.events[txHash.String()] = append(events, event)

	// record change log
	if err := s.changelog.Put(key, bytes); err != nil { // TODO value can be any value
		logging.VLog().Info("REE 12")
		return err
	}
	return nil
}

func (s *states) FetchCacheEventsOfCurBlock(txHash byteutils.Hash) ([]*Event, error) { // TODO delete, & interface
	txevents, ok := s.events[txHash.String()]
	if !ok {
		return nil, nil
	}

	events := []*Event{}
	for _, event := range txevents {
		events = append(events, event)
	}

	return events, nil
}

func (s *states) FetchEvents(txHash byteutils.Hash) ([]*Event, error) {
	events := []*Event{}
	iter, err := s.eventsState.Iterator(txHash)
	if err != nil && err != storage.ErrKeyNotFound {
		return nil, err
	}
	if err != storage.ErrKeyNotFound { // TODO -> err == nil
		exist, err := iter.Next()
		if err != nil {
			logging.VLog().Info("FEE 11")
			return nil, err
		}
		for exist {
			event := new(Event)
			err = json.Unmarshal(iter.Value(), event)
			if err != nil {
				logging.VLog().Info("FEE 12")
				return nil, err
			}
			events = append(events, event)
			// record change log
			if _, err := s.changelog.Get(iter.Key()); err != nil && err != storage.ErrKeyNotFound { // TODO remove events & txs changelog
				logging.VLog().Info("FEE 13")
				return nil, err
			}
			exist, err = iter.Next()
			if err != nil {
				logging.VLog().Info("FEE 14")
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
	accState, err := NewAccountState(root, s.storage)
	if err != nil {
		return err
	}
	s.accState = accState
	return nil
}

func (s *states) LoadTxsRoot(root byteutils.Hash) error {
	txsState, err := trie.NewTrie(root, s.storage, false)
	if err != nil {
		return err
	}
	s.txsState = txsState
	return nil
}

func (s *states) LoadEventsRoot(root byteutils.Hash) error {
	eventsState, err := trie.NewTrie(root, s.storage, false)
	if err != nil {
		return err
	}
	s.eventsState = eventsState
	return nil
}

func (s *states) LoadConsensusRoot(root *consensuspb.ConsensusRoot) error {
	consensusState, err := s.consensus.NewState(root, s.storage, false)
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
	var err error
	s.gasConsumed[from], err = consumed.Add(gas) // TODO use tmp var, assign if success

	return err
}

func (s *states) GetGas() map[string]*util.Uint128 {
	gasConsumed := make(map[string]*util.Uint128)
	for from, gas := range s.gasConsumed {
		gasConsumed[from] = gas
	}
	s.gasConsumed = make(map[string]*util.Uint128)
	return gasConsumed
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
		logging.CLog().Info("BE 1")
		return err
	}
	if err := ws.states.Begin(); err != nil {
		logging.CLog().Info("BE 2")
		return err
	}
	ws.snapshot = snapshot
	return nil
}

func (ws *worldState) Commit() error {
	if err := ws.states.Commit(); err != nil {
		logging.CLog().Info("CE 1")
		return err
	}
	ws.snapshot = nil
	return nil
}

func (ws *worldState) RollBack() error {
	if err := ws.states.RollBack(); err != nil {
		logging.CLog().Info("RE 1")
		return err
	}
	ws.states = ws.snapshot
	ws.snapshot = nil
	return nil
}

func (ws *worldState) Prepare(txid interface{}) (TxWorldState, error) {
	s, err := ws.states.Prepare(txid)
	if err != nil {
		logging.VLog().Info("PPE 1")
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
		logging.VLog().Info("CUE 1")
		return nil, err
	}
	tws.parent = nil
	return dependencies, nil
}

func (tws *txWorldState) Reset() error {
	if err := tws.states.Reset(); err != nil {
		logging.VLog().Info("RSE 1")
		return err
	}
	return nil
}

func (tws *txWorldState) Close() error {
	if err := tws.states.Close(); err != nil {
		logging.VLog().Info("CSE 1")
		return err
	}
	tws.parent = nil
	return nil
}

func (tws *txWorldState) TxID() interface{} {
	return tws.txid
}
