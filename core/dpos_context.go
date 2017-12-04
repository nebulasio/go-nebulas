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
	"hash/fnv"
	"sort"

	"github.com/nebulasio/go-nebulas/crypto/sha3"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/common/trie"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	log "github.com/sirupsen/logrus"
)

// Consensus Related Constants
const (
	BlockInterval        = int64(5)
	AcceptedNetWorkDelay = int64(2)
	DynastyInterval      = int64(3600)
	DynastySize          = 21
)

// DposContext carry context in dpos consensus
type DposContext struct {
	dynastyTrie     *trie.BatchTrie
	nextDynastyTrie *trie.BatchTrie
	delegateTrie    *trie.BatchTrie

	storage storage.Storage
}

// NewDposContext create a new dpos context
func NewDposContext(storage storage.Storage) *DposContext {
	dynastyTrie, _ := trie.NewBatchTrie(nil, storage)
	nextDynastyTrie, _ := trie.NewBatchTrie(nil, storage)
	delegateTrie, _ := trie.NewBatchTrie(nil, storage)
	return &DposContext{
		dynastyTrie:     dynastyTrie,
		nextDynastyTrie: nextDynastyTrie,
		delegateTrie:    delegateTrie,
		storage:         storage,
	}
}

// RootHash hash dpos context root hash
func (dc *DposContext) RootHash() byteutils.Hash {
	hasher := sha3.New256()

	hasher.Write(dc.dynastyTrie.RootHash())
	hasher.Write(dc.nextDynastyTrie.RootHash())
	hasher.Write(dc.delegateTrie.RootHash())

	return hasher.Sum(nil)
}

// BeginBatch starts a batch task
func (dc *DposContext) BeginBatch() {
	dc.delegateTrie.BeginBatch()
	dc.dynastyTrie.BeginBatch()
	dc.nextDynastyTrie.BeginBatch()
}

// Commit a batch task
func (dc *DposContext) Commit() {
	dc.delegateTrie.Commit()
	dc.dynastyTrie.Commit()
	dc.nextDynastyTrie.Commit()
}

// RollBack a batch task
func (dc *DposContext) RollBack() {
	dc.delegateTrie.RollBack()
	dc.dynastyTrie.RollBack()
	dc.nextDynastyTrie.RollBack()
}

// Clone a dpos context
func (dc *DposContext) Clone() (*DposContext, error) {
	var err error
	context := NewDposContext(dc.storage)
	if context.dynastyTrie, err = dc.dynastyTrie.Clone(); err != nil {
		return nil, err
	}
	if context.nextDynastyTrie, err = dc.nextDynastyTrie.Clone(); err != nil {
		return nil, err
	}
	if context.delegateTrie, err = dc.delegateTrie.Clone(); err != nil {
		return nil, err
	}
	return context, nil
}

// ToProto converts domain DposContext to proto DposContext
func (dc *DposContext) ToProto() (*corepb.DposContext, error) {
	return &corepb.DposContext{
		DynastyRoot:     dc.dynastyTrie.RootHash(),
		NextDynastyRoot: dc.nextDynastyTrie.RootHash(),
		DelegateRoot:    dc.delegateTrie.RootHash(),
	}, nil
}

// FromProto converts proto DposContext to domain DposContext
func (dc *DposContext) FromProto(msg *corepb.DposContext) error {
	var err error
	if dc.dynastyTrie, err = trie.NewBatchTrie(msg.DynastyRoot, dc.storage); err != nil {
		return err
	}
	if dc.nextDynastyTrie, err = trie.NewBatchTrie(msg.NextDynastyRoot, dc.storage); err != nil {
		return err
	}
	if dc.delegateTrie, err = trie.NewBatchTrie(msg.DelegateRoot, dc.storage); err != nil {
		return err
	}
	return nil
}

// DynastyContext contains the dynasty context at given timestamp
type DynastyContext struct {
	TimeStamp       int64
	Offset          int64
	Proposer        byteutils.Hash
	DynastyTrie     *trie.BatchTrie
	NextDynastyTrie *trie.BatchTrie
	Storage         storage.Storage
}

func (block *Block) tallyVotes() (map[string]*util.Uint128, error) {
	votes := make(map[string]*util.Uint128)
	delegate := block.dposContext.delegateTrie
	if delegate.Empty() {
		return votes, nil
	}
	iter, err := delegate.Iterator(nil)
	if err == storage.ErrKeyNotFound {
		return nil, err
	}
	exist, err := iter.Next()
	if err != nil {
		return nil, err
	}
	for exist {
		bytes := iter.Value()
		delegate := &corepb.Delegate{}
		if proto.Unmarshal(bytes, delegate) != nil {
			return nil, err
		}
		delegatee, err := AddressParseFromBytes(delegate.Delegatee)
		if err != nil {
			return nil, err
		}
		score, ok := votes[delegatee.ToHex()]
		if !ok {
			score = util.NewUint128()
		}
		weight := block.accState.GetOrCreateUserAccount(delegate.Delegator).Balance()
		score.Add(score.Int, weight.Int)
		votes[delegatee.ToHex()] = score
		exist, err = iter.Next()
	}
	return votes, nil
}

// Candidate is a data structure to hold candidate's info.
type Candidate struct {
	Address *Address
	Votes   *util.Uint128
}

// Candidates is a slice of Candidates that implements sort.Interface to sort by Votes.
type Candidates []Candidate

func (p Candidates) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p Candidates) Len() int           { return len(p) }
func (p Candidates) Less(i, j int) bool { return p[i].Votes.Cmp(p[j].Votes.Int) < 0 }

func (block *Block) electNewDynasty(seed int64) (*trie.BatchTrie, error) {
	// collect candidates
	votes, err := block.tallyVotes()
	if len(votes) < DynastySize {
		log.Error(len(votes))
		return nil, ErrTooFewCandidates
	}
	candidates := make(Candidates, len(votes))
	idx := 0
	for k, v := range votes {
		addr, err := AddressParse(k)
		if err != nil {
			return nil, err
		}
		candidates[idx] = Candidate{addr, v}
		idx++
	}
	sort.Stable(candidates)
	// Top 20 are selected directly
	dynasty, err := trie.NewBatchTrie(nil, block.storage)
	directSelected := DynastySize - 1
	for i := 0; i < directSelected; i++ {
		delegatee := candidates[i].Address.Bytes()
		_, err := dynasty.Put(delegatee, delegatee)
		if err != nil {
			return nil, err
		}
	}
	// The last one is selected randomly
	hasher := fnv.New32a()
	hasher.Write(byteutils.FromInt64(seed))
	hasher.Write(block.Hash())
	result := int(hasher.Sum32()) % (len(votes) - directSelected)
	offset := result + DynastySize - 1
	delegatee := candidates[offset].Address.Bytes()
	_, err = dynasty.Put(delegatee, delegatee)
	if err != nil {
		return nil, err
	}
	return dynasty, nil
}

// LoadDynastyContext from a given context
func (block *Block) LoadDynastyContext(context *DynastyContext) {
	block.header.timestamp = context.TimeStamp
	block.dposContext.dynastyTrie = context.DynastyTrie
	block.dposContext.nextDynastyTrie = context.NextDynastyTrie
}

// NextDynastyContext when some seconds elapsed
func (block *Block) NextDynastyContext(elapsedSecond int64) (*DynastyContext, error) {
	var err error
	nextTimeStamp := block.header.timestamp + elapsedSecond
	dynastyTrie := block.dposContext.dynastyTrie
	nextDynastyTrie := block.dposContext.nextDynastyTrie
	currentHour := block.header.timestamp / DynastyInterval
	nextHour := nextTimeStamp / DynastyInterval
	offset := nextTimeStamp % DynastyInterval
	if offset%BlockInterval != 0 {
		return nil, ErrNotBlockForgTime
	}
	offset /= BlockInterval
	offset %= DynastySize
	if currentHour < nextHour {
		seed := nextHour - currentHour
		if seed == 1 {
			dynastyTrie = nextDynastyTrie
		} else {
			dynastyTrie, err = block.electNewDynasty(seed - 1)
			if err != nil {
				return nil, err
			}
		}
		nextDynastyTrie, err = block.electNewDynasty(seed)
		if err != nil {
			return nil, err
		}
	}
	delegatees, err := TraverseDynasty(dynastyTrie)
	if err != nil {
		return nil, err
	}
	proposer := delegatees[offset]
	return &DynastyContext{
		TimeStamp:       nextTimeStamp,
		Offset:          offset,
		Proposer:        proposer,
		DynastyTrie:     dynastyTrie,
		NextDynastyTrie: nextDynastyTrie,
		Storage:         block.storage,
	}, nil
}

// TraverseDynasty return all members in the dynasty
func TraverseDynasty(dynasty *trie.BatchTrie) ([]byteutils.Hash, error) {
	members := []byteutils.Hash{}
	if dynasty.Empty() {
		return members, nil
	}
	iter, err := dynasty.Iterator(nil)
	if err == storage.ErrKeyNotFound {
		return members, nil
	}
	if err != nil {
		return nil, err
	}
	exist, err := iter.Next()
	for exist {
		members = append(members, iter.Value())
		exist, err = iter.Next()
	}
	return members, nil
}
