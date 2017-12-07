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

	"github.com/nebulasio/go-nebulas/common/trie"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/crypto/sha3"
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
	DynastySize          = 7
	ReserveSize          = DynastySize / 3
)

// DposContext carry context in dpos consensus
type DposContext struct {
	dynastyTrie     *trie.BatchTrie
	nextDynastyTrie *trie.BatchTrie
	delegateTrie    *trie.BatchTrie
	candidatesTrie  *trie.BatchTrie

	storage storage.Storage
}

// NewDposContext create a new dpos context
func NewDposContext(storage storage.Storage) (*DposContext, error) {
	dynastyTrie, err := trie.NewBatchTrie(nil, storage)
	if err != nil {
		return nil, err
	}
	nextDynastyTrie, err := trie.NewBatchTrie(nil, storage)
	if err != nil {
		return nil, err
	}
	delegateTrie, err := trie.NewBatchTrie(nil, storage)
	if err != nil {
		return nil, err
	}
	candidatesTrie, err := trie.NewBatchTrie(nil, storage)
	if err != nil {
		return nil, err
	}
	return &DposContext{
		dynastyTrie:     dynastyTrie,
		nextDynastyTrie: nextDynastyTrie,
		delegateTrie:    delegateTrie,
		candidatesTrie:  candidatesTrie,
		storage:         storage,
	}, nil
}

// RootHash hash dpos context root hash
func (dc *DposContext) RootHash() byteutils.Hash {
	hasher := sha3.New256()

	hasher.Write(dc.dynastyTrie.RootHash())
	hasher.Write(dc.nextDynastyTrie.RootHash())
	hasher.Write(dc.delegateTrie.RootHash())
	hasher.Write(dc.candidatesTrie.RootHash())

	return hasher.Sum(nil)
}

// BeginBatch starts a batch task
func (dc *DposContext) BeginBatch() {
	log.Info("DposContext Begin.")
	dc.delegateTrie.BeginBatch()
	dc.dynastyTrie.BeginBatch()
	dc.nextDynastyTrie.BeginBatch()
	dc.candidatesTrie.BeginBatch()
}

// Commit a batch task
func (dc *DposContext) Commit() {
	dc.delegateTrie.Commit()
	dc.dynastyTrie.Commit()
	dc.nextDynastyTrie.Commit()
	dc.candidatesTrie.Commit()
	log.Info("DposContext Commit.")
}

// RollBack a batch task
func (dc *DposContext) RollBack() {
	dc.delegateTrie.RollBack()
	dc.dynastyTrie.RollBack()
	dc.nextDynastyTrie.RollBack()
	dc.candidatesTrie.RollBack()
	log.Info("DposContext RollBack.")
}

// Clone a dpos context
func (dc *DposContext) Clone() (*DposContext, error) {
	var err error
	context, err := NewDposContext(dc.storage)
	if err != nil {
		return nil, err
	}
	if context.dynastyTrie, err = dc.dynastyTrie.Clone(); err != nil {
		log.Error("DynastyTrie Clone Error")
		return nil, err
	}
	if context.nextDynastyTrie, err = dc.nextDynastyTrie.Clone(); err != nil {
		log.Error("NextDynastyTrie Clone Error")
		return nil, err
	}
	if context.delegateTrie, err = dc.delegateTrie.Clone(); err != nil {
		log.Error("DelegateTrie Clone Error")
		return nil, err
	}
	if context.candidatesTrie, err = dc.candidatesTrie.Clone(); err != nil {
		log.Error("CandidatesTrie Clone Error")
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
		CandidatesRoot:  dc.candidatesTrie.RootHash(),
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
	if dc.candidatesTrie, err = trie.NewBatchTrie(msg.CandidatesRoot, dc.storage); err != nil {
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
	DelegateTrie    *trie.BatchTrie
	CandidatesTrie  *trie.BatchTrie
	Storage         storage.Storage
}

func (block *Block) tallyVotes() (map[string]*util.Uint128, error) {
	votes := make(map[string]*util.Uint128)
	delegate := block.dposContext.delegateTrie
	candidates := block.dposContext.candidatesTrie
	if candidates.Empty() || delegate.Empty() {
		return votes, nil
	}
	iterCandidates, err := candidates.Iterator(nil)
	existCandidates, err := iterCandidates.Next()
	if err != nil {
		return nil, err
	}
	for existCandidates {
		delegatee, err := AddressParseFromBytes(iterCandidates.Value())
		if err != nil {
			return nil, err
		}
		iterDelegate, err := delegate.Iterator(delegatee.Bytes())
		existDelegate, err := iterDelegate.Next()
		if err != nil {
			return nil, err
		}
		for existDelegate {
			delegator, err := AddressParseFromBytes(iterDelegate.Value())
			if err != nil {
				return nil, err
			}
			score, ok := votes[delegatee.ToHex()]
			if !ok {
				score = util.NewUint128()
			}
			weight := block.accState.GetOrCreateUserAccount(delegator.Bytes()).Balance()
			score.Add(score.Int, weight.Int)
			votes[delegatee.ToHex()] = score
			existDelegate, err = iterDelegate.Next()
			if err != nil {
				return nil, err
			}
		}
		existCandidates, err = iterCandidates.Next()
		if err != nil {
			return nil, err
		}
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

func (p Candidates) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p Candidates) Len() int      { return len(p) }
func (p Candidates) Less(i, j int) bool {
	if p[i].Votes.Cmp(p[j].Votes.Int) < 0 {
		return true
	} else if p[i].Votes.Cmp(p[j].Votes.Int) > 0 {
		return false
	} else {
		return p[i].Address.String() < p[j].Address.String()
	}
}

func candidates(votes map[string]*util.Uint128) (Candidates, error) {
	// remove reserved
	reserved := make(Candidates, ReserveSize+1)
	for i := 0; i <= ReserveSize; i++ {
		delete(votes, GenesisDynasty[i])
		address, err := AddressParse(GenesisDynasty[i])
		if err != nil {
			return nil, err
		}
		reserved[i] = Candidate{
			Address: address,
			Votes:   util.NewUint128(),
		}
	}
	// sort
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
	sort.Sort(candidates)
	// add reserved
	candidates = append(reserved, candidates...)
	return candidates, nil
}

func (block *Block) electNewDynasty(seed int64) (*trie.BatchTrie, error) {
	// collect candidates
	votes, err := block.tallyVotes()
	if len(votes) < DynastySize {
		log.Error(len(votes))
		return nil, ErrTooFewCandidates
	}
	candidates, err := candidates(votes)
	if err != nil {
		return nil, err
	}
	// Top 20 are selected directly
	dynasty, err := trie.NewBatchTrie(nil, block.storage)
	directSelected := DynastySize - 1
	for i := 0; i < directSelected; i++ {
		delegatee := candidates[i].Address.Bytes()
		log.Info(candidates[i].Address.ToHex())
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
	log.Info(candidates[offset].Address.ToHex())
	_, err = dynasty.Put(delegatee, delegatee)
	if err != nil {
		return nil, err
	}
	log.Info("new dynasty")
	return dynasty, nil
}

// LoadDynastyContext from a given context
func (block *Block) LoadDynastyContext(context *DynastyContext) {
	block.header.timestamp = context.TimeStamp
	block.dposContext = &DposContext{
		dynastyTrie:     context.DynastyTrie,
		nextDynastyTrie: context.NextDynastyTrie,
		delegateTrie:    context.DelegateTrie,
		candidatesTrie:  context.CandidatesTrie,
		storage:         block.storage,
	}
}

// GenesisDynastyContext return dynasty context in genesis
func GenesisDynastyContext(storage storage.Storage) (*DynastyContext, error) {
	dynasty, err := trie.NewBatchTrie(nil, storage)
	if err != nil {
		return nil, err
	}
	delegate, err := trie.NewBatchTrie(nil, storage)
	if err != nil {
		return nil, err
	}
	candidates, err := trie.NewBatchTrie(nil, storage)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(GenesisDynasty); i++ {
		member, err := AddressParse(GenesisDynasty[i])
		if err != nil {
			return nil, err
		}
		v := member.Bytes()
		if _, err = dynasty.Put(v, v); err != nil {
			return nil, err
		}
		if _, err = candidates.Put(v, v); err != nil {
			return nil, err
		}
		key := append(v, v...)
		if _, err = delegate.Put(key, v); err != nil {
			return nil, err
		}
	}
	nextDynasty, err := dynasty.Clone()
	if err != nil {
		return nil, err
	}
	return &DynastyContext{
		TimeStamp:       GenesisTimestamp,
		DynastyTrie:     dynasty,
		NextDynastyTrie: nextDynasty,
		DelegateTrie:    delegate,
		CandidatesTrie:  candidates,
	}, nil
}

// NextDynastyContext when some seconds elapsed
func (block *Block) NextDynastyContext(elapsedSecond int64) (*DynastyContext, error) {
	var err error
	nextTimeStamp := block.header.timestamp + elapsedSecond
	dynastyTrie := block.dposContext.dynastyTrie
	nextDynastyTrie := block.dposContext.nextDynastyTrie
	delegateTrie := block.dposContext.delegateTrie
	candidatesTrie := block.dposContext.candidatesTrie
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
		DelegateTrie:    delegateTrie,
		CandidatesTrie:  candidatesTrie,
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
