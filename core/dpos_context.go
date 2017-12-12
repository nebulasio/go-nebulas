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
	DynastyInterval      = int64(60)
	DynastySize          = 7
	ReserveSize          = DynastySize / 3
)

// DposContext carry context in dpos consensus
type DposContext struct {
	dynastyTrie     *trie.BatchTrie // key: delegatee, val: delegatee
	nextDynastyTrie *trie.BatchTrie // key: delegatee, val: delegatee
	delegateTrie    *trie.BatchTrie // key: delegatee + delegator, val: delegator
	voteTrie        *trie.BatchTrie // key: delegator, val: delegatee
	candidateTrie   *trie.BatchTrie // key: delegatee, val: delegatee
	mintCntTrie     *trie.BatchTrie // key: dynastyId + delegatee, val: count

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
	voteTrie, err := trie.NewBatchTrie(nil, storage)
	if err != nil {
		return nil, err
	}
	candidateTrie, err := trie.NewBatchTrie(nil, storage)
	if err != nil {
		return nil, err
	}
	mintCntTrie, err := trie.NewBatchTrie(nil, storage)
	if err != nil {
		return nil, err
	}
	return &DposContext{
		dynastyTrie:     dynastyTrie,
		nextDynastyTrie: nextDynastyTrie,
		delegateTrie:    delegateTrie,
		voteTrie:        voteTrie,
		candidateTrie:   candidateTrie,
		mintCntTrie:     mintCntTrie,
		storage:         storage,
	}, nil
}

// RootHash hash dpos context root hash
func (dc *DposContext) RootHash() byteutils.Hash {
	hasher := sha3.New256()

	log.Info("RootHash")
	log.Info(dc.dynastyTrie.RootHash())
	log.Info(dc.nextDynastyTrie.RootHash())
	log.Info(dc.delegateTrie.RootHash())
	log.Info(dc.voteTrie.RootHash())
	log.Info(dc.candidateTrie.RootHash())
	log.Info(dc.mintCntTrie.RootHash())
	hasher.Write(dc.dynastyTrie.RootHash())
	hasher.Write(dc.nextDynastyTrie.RootHash())
	hasher.Write(dc.delegateTrie.RootHash())
	hasher.Write(dc.voteTrie.RootHash())
	hasher.Write(dc.candidateTrie.RootHash())
	hasher.Write(dc.mintCntTrie.RootHash())

	return hasher.Sum(nil)
}

// BeginBatch starts a batch task
func (dc *DposContext) BeginBatch() {
	log.Info("DposContext Begin.")
	dc.delegateTrie.BeginBatch()
	dc.dynastyTrie.BeginBatch()
	dc.nextDynastyTrie.BeginBatch()
	dc.candidateTrie.BeginBatch()
	dc.voteTrie.BeginBatch()
	dc.mintCntTrie.BeginBatch()
}

// Commit a batch task
func (dc *DposContext) Commit() {
	dc.delegateTrie.Commit()
	dc.dynastyTrie.Commit()
	dc.nextDynastyTrie.Commit()
	dc.candidateTrie.Commit()
	dc.voteTrie.Commit()
	dc.mintCntTrie.Commit()
	log.Info("DposContext Commit.")
}

// RollBack a batch task
func (dc *DposContext) RollBack() {
	dc.delegateTrie.RollBack()
	dc.dynastyTrie.RollBack()
	dc.nextDynastyTrie.RollBack()
	dc.candidateTrie.RollBack()
	dc.voteTrie.RollBack()
	dc.mintCntTrie.RollBack()
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
	if context.candidateTrie, err = dc.candidateTrie.Clone(); err != nil {
		log.Error("CandidatesTrie Clone Error")
		return nil, err
	}
	if context.voteTrie, err = dc.voteTrie.Clone(); err != nil {
		log.Error("VoteTrie Clone Error")
		return nil, err
	}
	if context.mintCntTrie, err = dc.mintCntTrie.Clone(); err != nil {
		log.Error("MintCntTrie Clone Error")
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
		CandidateRoot:   dc.candidateTrie.RootHash(),
		VoteRoot:        dc.voteTrie.RootHash(),
		MintCntRoot:     dc.mintCntTrie.RootHash(),
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
	if dc.candidateTrie, err = trie.NewBatchTrie(msg.CandidateRoot, dc.storage); err != nil {
		return err
	}
	if dc.voteTrie, err = trie.NewBatchTrie(msg.VoteRoot, dc.storage); err != nil {
		return err
	}
	if dc.mintCntTrie, err = trie.NewBatchTrie(msg.MintCntRoot, dc.storage); err != nil {
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
	CandidateTrie   *trie.BatchTrie
	VoteTrie        *trie.BatchTrie
	MintCntTrie     *trie.BatchTrie
	Storage         storage.Storage
}

func (block *Block) tallyVotes() (map[string]*util.Uint128, error) {
	votes := make(map[string]*util.Uint128)
	delegate := block.dposContext.delegateTrie
	candidates := block.dposContext.candidateTrie
	if candidates.Empty() {
		return votes, nil
	}
	iterCandidates, err := candidates.Iterator(nil)
	if err != nil && err != storage.ErrKeyNotFound {
		return nil, err
	}
	if err != nil {
		return votes, nil
	}
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
		if err != nil && err != storage.ErrKeyNotFound {
			return nil, err
		}
		if err != nil {
			votes[delegatee.ToHex()] = util.NewUint128()
			existCandidates, err = iterCandidates.Next()
			if err != nil {
				return nil, err
			}
			continue
		}
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
		return false
	} else if p[i].Votes.Cmp(p[j].Votes.Int) > 0 {
		return true
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

func (block *Block) kickoutCandidate(candidate byteutils.Hash) error {
	log.Info("Kickout Candidate: ", candidate.Hex())
	if _, err := block.dposContext.candidateTrie.Del(candidate); err != nil {
		return err
	}
	iter, err := block.dposContext.delegateTrie.Iterator(candidate)
	if err != nil && err != storage.ErrKeyNotFound {
		return err
	}
	if err != nil {
		return nil
	}
	exist, err := iter.Next()
	if err != nil {
		return err
	}
	for exist {
		delegator := iter.Value()
		key := append(candidate, delegator...)
		if _, err := block.dposContext.delegateTrie.Del(key); err != nil {
			return err
		}
		bytes, err := block.dposContext.voteTrie.Get(delegator)
		if err != nil {
			return err
		}
		if byteutils.Equal(bytes, candidate) {
			if _, err := block.dposContext.voteTrie.Del(delegator); err != nil {
				return err
			}
		}
		exist, err = iter.Next()
		if err != nil {
			return err
		}
	}
	return nil
}

func (block *Block) kickoutDumbDynasty(id int64, dynasty *trie.BatchTrie) error {
	log.Info("Kickout Dynasty: ", id)
	iter, err := dynasty.Iterator(nil)
	if err != nil {
		return err
	}
	exist, err := iter.Next()
	if err != nil {
		return err
	}
	for exist {
		validator := iter.Value()
		key := append(byteutils.FromInt64(id), validator...)
		bytes, err := block.dposContext.mintCntTrie.Get(key)
		if err != nil && err != storage.ErrKeyNotFound {
			return err
		}
		if err != storage.ErrKeyNotFound {
			cnt := byteutils.Int64(bytes)
			if cnt >= DynastyInterval/BlockInterval/2 {
				exist, err = iter.Next()
				if err != nil {
					return err
				}
				continue
			}
		}
		if err := block.kickoutCandidate(validator); err != nil {
			return err
		}
		exist, err = iter.Next()
		if err != nil {
			return err
		}
	}
	return nil
}

func (block *Block) kickoutDumbValidators(curDynastyID int64, nextDynastyID int64) error {
	// do not kickout genesis dynasty
	if curDynastyID <= 0 {
		return nil
	}
	for i := curDynastyID; i < nextDynastyID; i++ {
		context, err := block.NextDynastyContext(curDynastyID*DynastyInterval - block.Timestamp())
		if err != nil {
			return err
		}
		err = block.kickoutDumbDynasty(i, context.DynastyTrie)
		if err != nil {
			return err
		}
	}
	return nil
}

func (block *Block) electNewDynasty(curDynastyID int64, newDynastyID int64) (*trie.BatchTrie, error) {
	// collect candidates
	err := block.kickoutDumbValidators(curDynastyID, newDynastyID)
	if err != nil {
		return nil, err
	}
	votes, err := block.tallyVotes()
	if err != nil {
		return nil, err
	}
	candidates, err := candidates(votes)
	if err != nil {
		return nil, err
	}
	if len(candidates) < DynastySize {
		log.Error(len(candidates))
		return nil, ErrTooFewCandidates
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
	hasher.Write(byteutils.FromInt64(newDynastyID))
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
func (block *Block) LoadDynastyContext(context *DynastyContext) error {
	block.header.timestamp = context.TimeStamp
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
	block.dposContext = &DposContext{
		dynastyTrie:     dynastyTrie,
		nextDynastyTrie: nextDynastyTrie,
		delegateTrie:    delegateTrie,
		candidateTrie:   candidateTrie,
		voteTrie:        voteTrie,
		mintCntTrie:     mintCntTrie,
		storage:         block.storage,
	}
	return nil
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
	candidate, err := trie.NewBatchTrie(nil, storage)
	if err != nil {
		return nil, err
	}
	vote, err := trie.NewBatchTrie(nil, storage)
	if err != nil {
		return nil, err
	}
	mint, err := trie.NewBatchTrie(nil, storage)
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
		if _, err = candidate.Put(v, v); err != nil {
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
		CandidateTrie:   candidate,
		MintCntTrie:     mint,
		VoteTrie:        vote,
	}, nil
}

// NextDynastyContext when some seconds elapsed
func (block *Block) NextDynastyContext(elapsedSecond int64) (*DynastyContext, error) {
	var err error
	nextTimeStamp := block.header.timestamp + elapsedSecond
	dynastyTrie := block.dposContext.dynastyTrie
	nextDynastyTrie := block.dposContext.nextDynastyTrie
	delegateTrie := block.dposContext.delegateTrie
	candidateTrie := block.dposContext.candidateTrie
	voteTrie := block.dposContext.voteTrie
	mintCntTrie := block.dposContext.mintCntTrie
	currentHour := block.header.timestamp / DynastyInterval
	nextHour := nextTimeStamp / DynastyInterval
	offset := nextTimeStamp % DynastyInterval
	if offset%BlockInterval != 0 {
		return nil, ErrNotBlockForgTime
	}
	offset /= BlockInterval
	offset %= DynastySize
	if currentHour < nextHour {
		if nextHour == currentHour+1 {
			dynastyTrie = nextDynastyTrie
		} else {
			dynastyTrie, err = block.electNewDynasty(currentHour, nextHour-1)
			if err != nil {
				return nil, err
			}
		}
		nextDynastyTrie, err = block.electNewDynasty(currentHour, nextHour)
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
		CandidateTrie:   candidateTrie,
		VoteTrie:        voteTrie,
		MintCntTrie:     mintCntTrie,
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
