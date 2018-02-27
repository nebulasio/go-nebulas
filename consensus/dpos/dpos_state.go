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

package dpos

import (
	"errors"
	"fmt"
	"hash/fnv"
	"sort"
	"strconv"
	"time"

	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/core/state"

	"github.com/nebulasio/go-nebulas/common/trie"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

// Consensus Related Constants
const (
	BlockInterval        = int64(5)
	AcceptedNetWorkDelay = int64(2)
	MaxMintDuration      = int64(2)
	MinMintDuration      = int64(1)
	DynastyInterval      = int64(60) // TODO(roy): 3600
	DynastySize          = 6         // TODO(roy): 21
	SafeSize             = DynastySize/3 + 1
	ConsensusSize        = DynastySize*2/3 + 1
)

// Errors in dpos state
var (
	ErrTooFewCandidates        = errors.New("the size of candidates in consensus is un-safe, should be greater than or equal " + strconv.Itoa(SafeSize))
	ErrInitialDynastyNotEnough = errors.New("the size of initial dynasty in genesis block is un-safe, should be greater than or equal " + strconv.Itoa(SafeSize))
	ErrCloneDynastyTrie        = errors.New("Failed to clone dynasty trie")
	ErrCloneNextDynastyTrie    = errors.New("Failed to clone next dynasty trie")
	ErrCloneDelegateTrie       = errors.New("Failed to clone delegate trie")
	ErrCloneCandidatesTrie     = errors.New("Failed to clone candidates trie")
	ErrCloneVoteTrie           = errors.New("Failed to clone vote trie")
	ErrCloneMintCntTrie        = errors.New("Failed to clone mint count trie")
	ErrNotBlockForgTime        = errors.New("now is not time to forg block")
	ErrFoundNilProposer        = errors.New("found a nil proposer")
)

// State carry context in dpos consensus
type State struct {
	timeStamp int64
	proposer  byteutils.Hash

	dynastyTrie     *trie.BatchTrie // key: delegatee, val: delegatee
	nextDynastyTrie *trie.BatchTrie // key: delegatee, val: delegatee
	delegateTrie    *trie.BatchTrie // key: delegatee + delegator, val: delegator
	voteTrie        *trie.BatchTrie // key: delegator, val: delegatee
	candidateTrie   *trie.BatchTrie // key: delegatee, val: delegatee
	mintCntTrie     *trie.BatchTrie // key: dynastyId + delegatee, val: count
	protectTrie     *trie.BatchTrie // key: delegatee, val: delegatee

	chain     *core.BlockChain
	consensus core.Consensus
}

// NewState create a new dpos state
func (dpos *Dpos) NewState(root byteutils.Hash, stor storage.Storage) (state.ConsensusState, error) {
	stateTrie, err := trie.NewTrie(root, stor)
	if err != nil {
		return nil, err
	}

	var index int16
	bytes, err := stateTrie.Get(byteutils.FromInt16(index))
	if err != nil {
		return nil, err
	}
	timestamp := byteutils.Int64(bytes)

	index++
	proposer, err := stateTrie.Get(byteutils.FromInt16(index))
	if err != nil {
		return nil, err
	}

	index++
	dynastyRoot, err := stateTrie.Get(byteutils.FromInt16(index))
	if err != nil && root != nil {
		return nil, err
	}
	dynastyTrie, err := trie.NewBatchTrie(dynastyRoot, stor)
	if err != nil {
		return nil, err
	}

	index++
	nextDynastyRoot, err := stateTrie.Get(byteutils.FromInt16(index))
	if err != nil && root != nil {
		return nil, err
	}
	nextDynastyTrie, err := trie.NewBatchTrie(nextDynastyRoot, stor)
	if err != nil {
		return nil, err
	}

	index++
	delegateRoot, err := stateTrie.Get(byteutils.FromInt16(index))
	if err != nil && root != nil {
		return nil, err
	}
	delegateTrie, err := trie.NewBatchTrie(delegateRoot, stor)
	if err != nil {
		return nil, err
	}

	index++
	voteRoot, err := stateTrie.Get(byteutils.FromInt16(index))
	if err != nil && root != nil {
		return nil, err
	}
	voteTrie, err := trie.NewBatchTrie(voteRoot, stor)
	if err != nil {
		return nil, err
	}

	index++
	candidateRoot, err := stateTrie.Get(byteutils.FromInt16(index))
	if err != nil && root != nil {
		return nil, err
	}
	candidateTrie, err := trie.NewBatchTrie(candidateRoot, stor)
	if err != nil {
		return nil, err
	}

	index++
	mintCntRoot, err := stateTrie.Get(byteutils.FromInt16(index))
	if err != nil && root != nil {
		return nil, err
	}
	mintCntTrie, err := trie.NewBatchTrie(mintCntRoot, stor)
	if err != nil {
		return nil, err
	}

	index++
	protectRoot, err := stateTrie.Get(byteutils.FromInt16(index))
	if err != nil && root != nil {
		return nil, err
	}
	protectTrie, err := trie.NewBatchTrie(protectRoot, stor)
	if err != nil {
		return nil, err
	}

	return &State{
		timeStamp: timestamp,
		proposer:  proposer,

		dynastyTrie:     dynastyTrie,
		nextDynastyTrie: nextDynastyTrie,
		delegateTrie:    delegateTrie,
		voteTrie:        voteTrie,
		candidateTrie:   candidateTrie,
		mintCntTrie:     mintCntTrie,
		protectTrie:     protectTrie,

		chain:     dpos.chain,
		consensus: dpos,
	}, nil
}

// CheckTimeout check whether the block is timeout
func (dpos *Dpos) CheckTimeout(block *core.Block) bool {
	behind := time.Now().Unix() - block.Timestamp()
	if behind > AcceptedNetWorkDelay {
		logging.VLog().WithFields(logrus.Fields{
			"block": block,
			"diff":  behind,
			"limit": AcceptedNetWorkDelay,
			"err":   "timeout",
		}).Debug("Found a timeout block.")
		return true
	}
	return false
}

// GenesisConsensusState create a new genesis dpos state
func (dpos *Dpos) GenesisConsensusState(chain *core.BlockChain, conf *corepb.Genesis) (state.ConsensusState, error) {
	dynastyTrie, err := trie.NewBatchTrie(nil, chain.Storage())
	if err != nil {
		return nil, err
	}
	delegateTrie, err := trie.NewBatchTrie(nil, chain.Storage())
	if err != nil {
		return nil, err
	}
	candidateTrie, err := trie.NewBatchTrie(nil, chain.Storage())
	if err != nil {
		return nil, err
	}
	voteTrie, err := trie.NewBatchTrie(nil, chain.Storage())
	if err != nil {
		return nil, err
	}
	mintTrie, err := trie.NewBatchTrie(nil, chain.Storage())
	if err != nil {
		return nil, err
	}
	if len(conf.Consensus.Dpos.Dynasty) < SafeSize {
		return nil, ErrInitialDynastyNotEnough
	}
	for i := 0; i < len(conf.Consensus.Dpos.Dynasty); i++ {
		addr := conf.Consensus.Dpos.Dynasty[i]
		member, err := core.AddressParse(addr)
		if err != nil {
			return nil, err
		}
		v := member.Bytes()
		if i < DynastySize {
			if _, err = dynastyTrie.Put(v, v); err != nil {
				return nil, err
			}
		}
		if _, err = voteTrie.Put(v, v); err != nil {
			return nil, err
		}
		key := append(v, v...)
		if _, err = delegateTrie.Put(key, v); err != nil {
			return nil, err
		}
		if _, err = candidateTrie.Put(v, v); err != nil {
			return nil, err
		}
	}
	nextDynastyTrie, err := dynastyTrie.Clone()
	if err != nil {
		return nil, err
	}
	protectTrie, err := candidateTrie.Clone()
	if err != nil {
		return nil, err
	}
	return &State{
		timeStamp: core.GenesisTimestamp,
		proposer:  nil,

		dynastyTrie:     dynastyTrie,
		nextDynastyTrie: nextDynastyTrie,
		delegateTrie:    delegateTrie,
		voteTrie:        voteTrie,
		candidateTrie:   candidateTrie,
		mintCntTrie:     mintTrie,
		protectTrie:     protectTrie,

		chain:     chain,
		consensus: dpos,
	}, nil
}

func (ds *State) String() string {
	proposer := ""
	if ds.proposer != nil {
		proposer = ds.proposer.String()
	}
	return fmt.Sprintf(`{"timestamp": %d, "proposer": "%s", "dynasty": "%s", "next_dynasty": "%s", "delegate": "%s", "vote": "%s", "candidate": "%s", "mintcnt": "%s", "protect": "%s"}`,
		ds.timeStamp,
		proposer,
		byteutils.Hex(ds.dynastyTrie.RootHash()),
		byteutils.Hex(ds.nextDynastyTrie.RootHash()),
		byteutils.Hex(ds.delegateTrie.RootHash()),
		byteutils.Hex(ds.voteTrie.RootHash()),
		byteutils.Hex(ds.candidateTrie.RootHash()),
		byteutils.Hex(ds.mintCntTrie.RootHash()),
		byteutils.Hex(ds.protectTrie.RootHash()),
	)
}

// BeginBatch starts a batch task
func (ds *State) BeginBatch() {
	// logging.VLog().Debug("State Begin.")
	ds.delegateTrie.BeginBatch()
	ds.dynastyTrie.BeginBatch()
	ds.nextDynastyTrie.BeginBatch()
	ds.candidateTrie.BeginBatch()
	ds.voteTrie.BeginBatch()
	ds.mintCntTrie.BeginBatch()
}

// Commit a batch task
func (ds *State) Commit() {
	ds.delegateTrie.Commit()
	ds.dynastyTrie.Commit()
	ds.nextDynastyTrie.Commit()
	ds.candidateTrie.Commit()
	ds.voteTrie.Commit()
	ds.mintCntTrie.Commit()
	// logging.VLog().Debug("State Commit.")
}

// RollBack a batch task
func (ds *State) RollBack() {
	ds.delegateTrie.RollBack()
	ds.dynastyTrie.RollBack()
	ds.nextDynastyTrie.RollBack()
	ds.candidateTrie.RollBack()
	ds.voteTrie.RollBack()
	ds.mintCntTrie.RollBack()
	// logging.VLog().Debug("State RollBack.")
}

// Clone a dpos context
func (ds *State) Clone() (state.ConsensusState, error) {
	var err error
	dynastyTrie, err := ds.dynastyTrie.Clone()
	if err != nil {
		return nil, ErrCloneDynastyTrie
	}
	nextDynastyTrie, err := ds.nextDynastyTrie.Clone()
	if err != nil {
		return nil, ErrCloneNextDynastyTrie
	}
	delegateTrie, err := ds.delegateTrie.Clone()
	if err != nil {
		return nil, ErrCloneDelegateTrie
	}
	candidateTrie, err := ds.candidateTrie.Clone()
	if err != nil {
		return nil, ErrCloneCandidatesTrie
	}
	voteTrie, err := ds.voteTrie.Clone()
	if err != nil {
		return nil, ErrCloneVoteTrie
	}
	mintCntTrie, err := ds.mintCntTrie.Clone()
	if err != nil {
		return nil, ErrCloneMintCntTrie
	}
	return &State{
		timeStamp: ds.timeStamp,
		proposer:  ds.proposer,

		dynastyTrie:     dynastyTrie,
		nextDynastyTrie: nextDynastyTrie,
		delegateTrie:    delegateTrie,
		candidateTrie:   candidateTrie,
		voteTrie:        voteTrie,
		mintCntTrie:     mintCntTrie,
		protectTrie:     ds.protectTrie,

		chain:     ds.chain,
		consensus: ds.consensus,
	}, nil
}

// RootHash hash dpos state
func (ds *State) RootHash() (byteutils.Hash, error) {
	stateTrie, err := trie.NewTrie(nil, ds.chain.Storage())
	if err != nil {
		return nil, err
	}
	var cnt int16
	stateTrie.Put(byteutils.FromInt16(cnt), byteutils.FromInt64(ds.timeStamp))
	cnt++
	stateTrie.Put(byteutils.FromInt16(cnt), ds.proposer)
	cnt++
	stateTrie.Put(byteutils.FromInt16(cnt), ds.dynastyTrie.RootHash())
	cnt++
	stateTrie.Put(byteutils.FromInt16(cnt), ds.nextDynastyTrie.RootHash())
	cnt++
	stateTrie.Put(byteutils.FromInt16(cnt), ds.delegateTrie.RootHash())
	cnt++
	stateTrie.Put(byteutils.FromInt16(cnt), ds.voteTrie.RootHash())
	cnt++
	stateTrie.Put(byteutils.FromInt16(cnt), ds.candidateTrie.RootHash())
	cnt++
	stateTrie.Put(byteutils.FromInt16(cnt), ds.mintCntTrie.RootHash())
	cnt++
	stateTrie.Put(byteutils.FromInt16(cnt), ds.protectTrie.RootHash())

	return stateTrie.RootHash(), nil
}

// GetMintCnt return the count of blocks minted by the miner in given dynasty
func (ds *State) GetMintCnt(timestamp int64, miner byteutils.Hash) (int64, error) {
	dynasty := timestamp / DynastyInterval
	key := append(byteutils.FromInt64(dynasty), miner...)
	bytes, err := ds.mintCntTrie.Get(key)
	if err != nil && err != storage.ErrKeyNotFound {
		return 0, err
	}
	cnt := int64(0)
	if err != storage.ErrKeyNotFound {
		cnt = byteutils.Int64(bytes)
	}
	return cnt, nil
}

// PutMintCnt update the count of blocks minted by the miner in given dynasty
func (ds *State) PutMintCnt(timestamp int64, miner byteutils.Hash, cnt int64) error {
	dynasty := timestamp / DynastyInterval
	key := append(byteutils.FromInt64(dynasty), miner...)
	_, err := ds.mintCntTrie.Put(key, byteutils.FromInt64(cnt))
	if err != nil {
		return err
	}
	return nil
}

// HasCandidate return whether the candidate exists
func (ds *State) HasCandidate(candidate byteutils.Hash) (bool, error) {
	_, err := ds.candidateTrie.Get(candidate)
	if err != nil && err != storage.ErrKeyNotFound {
		return false, err
	}
	if err == storage.ErrKeyNotFound {
		return false, nil
	}
	return true, nil
}

// AddCandidate add a new candidate
func (ds *State) AddCandidate(candidate byteutils.Hash) error {
	_, err := ds.candidateTrie.Put(candidate, candidate)
	return err
}

// DelCandidate kickout the candidate
func (ds *State) DelCandidate(candidate byteutils.Hash) error {
	return ds.kickoutCandidate(candidate)
}

// GetVote return who the voters votes
func (ds *State) GetVote(voter byteutils.Hash) (byteutils.Hash, error) {
	return ds.voteTrie.Get(voter)
}

// AddVote add a new vote
func (ds *State) AddVote(voter byteutils.Hash, votee byteutils.Hash) error {
	_, err := ds.voteTrie.Put(voter, votee)
	return err
}

// DelVote del the voter's vote
func (ds *State) DelVote(voter byteutils.Hash) error {
	_, err := ds.voteTrie.Del(voter)
	return err
}

// IterVote return the iterator of all votes
func (ds *State) IterVote() (state.Iterator, error) {
	return ds.voteTrie.Iterator(nil)
}

// HasDelegate return whether the delegate exists
func (ds *State) HasDelegate(delegator byteutils.Hash, delegatee byteutils.Hash) (bool, error) {
	key := append(delegatee, delegator...)
	_, err := ds.delegateTrie.Get(key)
	if err != nil && err != storage.ErrKeyNotFound {
		return false, err
	}
	if err == storage.ErrKeyNotFound {
		return false, nil
	}
	return true, nil
}

// AddDelegate add a new delegate
func (ds *State) AddDelegate(delegator byteutils.Hash, delegatee byteutils.Hash) error {
	key := append(delegatee, delegator...)
	_, err := ds.delegateTrie.Put(key, delegator)
	return err
}

// DelDelegate del the delegate
func (ds *State) DelDelegate(delegator byteutils.Hash, delegatee byteutils.Hash) error {
	key := append(delegatee, delegator...)
	_, err := ds.delegateTrie.Del(key)
	return err
}

// IterDelegate return the iterator by delegatee
func (ds *State) IterDelegate(delegatee byteutils.Hash) (state.Iterator, error) {
	return ds.delegateTrie.Iterator(delegatee)
}

// Dynasty return the current dynasty
func (ds *State) Dynasty() ([]byteutils.Hash, error) {
	return TraverseDynasty(ds.dynastyTrie)
}

// DynastyRoot return the roothash of current dynasty
func (ds *State) DynastyRoot() byteutils.Hash {
	return ds.dynastyTrie.RootHash()
}

// NextDynasty return the next dynasty
func (ds *State) NextDynasty() ([]byteutils.Hash, error) {
	return TraverseDynasty(ds.nextDynastyTrie)
}

// NextDynastyRoot return the roothash of next dynasty
func (ds *State) NextDynastyRoot() byteutils.Hash {
	return ds.nextDynastyTrie.RootHash()
}

// CandidatesRoot return the roothash of current candidates
func (ds *State) CandidatesRoot() byteutils.Hash {
	return ds.candidateTrie.RootHash()
}

// Candidates return the current candidates
func (ds *State) Candidates() ([]byteutils.Hash, error) {
	return TraverseDynasty(ds.candidateTrie)
}

func (ds *State) tallyVotes(worldState state.WorldState) (map[string]*util.Uint128, error) {
	votes := make(map[string]*util.Uint128)
	delegate := ds.delegateTrie
	candidates := ds.candidateTrie
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
		delegatee, err := core.AddressParseFromBytes(iterCandidates.Value())
		if err != nil {
			return nil, err
		}
		iterDelegate, err := delegate.Iterator(delegatee.Bytes())
		if err != nil && err != storage.ErrKeyNotFound {
			return nil, err
		}
		if err != nil {
			votes[delegatee.String()] = util.NewUint128()
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
			delegator, err := core.AddressParseFromBytes(iterDelegate.Value())
			if err != nil {
				return nil, err
			}
			score, ok := votes[delegatee.String()]
			if !ok {
				score = util.NewUint128()
			}
			acc, err := worldState.GetOrCreateUserAccount(delegator.Bytes())
			if err != nil {
				return nil, err
			}
			weight := acc.Balance()
			score.Add(score.Int, weight.Int)
			votes[delegatee.String()] = score
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
	Address *core.Address
	Votes   *util.Uint128
}

// Candidates is a slice of Candidates that implements sort.Interface to sort by Votes.
type Candidates []*Candidate

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

func fetchActiveBootstapValidators(protect *trie.BatchTrie, candidates *trie.BatchTrie) ([]byteutils.Hash, error) {
	iter, err := protect.Iterator(nil)
	if err != nil && err != storage.ErrKeyNotFound {
		return nil, err
	}
	activeBootstapValidators := []byteutils.Hash{}
	if err != nil {
		return activeBootstapValidators, nil
	}
	exist, err := iter.Next()
	if err != nil {
		return nil, err
	}
	for exist {
		var validator byteutils.Hash = iter.Value()
		_, err = candidates.Get(validator)
		if err != nil && err != storage.ErrKeyNotFound {
			return nil, err
		}
		if err != storage.ErrKeyNotFound {
			activeBootstapValidators = append(activeBootstapValidators, validator)
		}
		exist, err = iter.Next()
		if err != nil {
			return nil, err
		}
	}
	return activeBootstapValidators, nil
}

func checkActiveBootstrapValidator(validator byteutils.Hash, protect *trie.BatchTrie, candidates *trie.BatchTrie) (bool, error) {
	_, err := protect.Get(validator)
	if err != nil && err != storage.ErrKeyNotFound {
		return false, err
	}
	if err != nil {
		return false, nil
	}
	_, err = candidates.Get(validator)
	if err != nil && err != storage.ErrKeyNotFound {
		return false, err
	}
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (ds *State) chooseCandidates(votes map[string]*util.Uint128) (Candidates, error) {
	// startAt := time.Now().Unix()
	// active bootstrap validators
	var bootstrapCandidates Candidates
	activeBootstrapValidators, err := fetchActiveBootstapValidators(ds.protectTrie, ds.candidateTrie)
	if err != nil {
		return nil, err
	}
	// fetchAt := time.Now().Unix()
	for i := 0; i < len(activeBootstrapValidators); i++ {
		address, err := core.AddressParseFromBytes(activeBootstrapValidators[i])
		if err != nil {
			return nil, err
		}
		vote := util.NewUint128()
		if v, ok := votes[address.String()]; ok {
			vote = v
		}
		bootstrapCandidates = append(bootstrapCandidates, &Candidate{address, vote})
		delete(votes, address.String())
	}
	// bootstrapAt := time.Now().Unix()
	sort.Sort(bootstrapCandidates)
	// sortBootStrapAt := time.Now().Unix()
	// sort
	var candidates Candidates
	for k, v := range votes {
		addr, err := core.AddressParse(k)
		if err != nil {
			return nil, err
		}
		candidates = append(candidates, &Candidate{addr, v})
	}
	// candidateAt := time.Now().Unix()
	sort.Sort(candidates)
	// sortCandidateAt := time.Now().Unix()
	// merge
	candidates = append(bootstrapCandidates, candidates...)

	/* 	logging.VLog().WithFields(logrus.Fields{
		"bootstrap.size":      bootstrapCandidates.Len(),
		"candidate.size":      len(candidates) - bootstrapCandidates.Len(),
		"time.fetch":          fetchAt - startAt,
		"time.bootstrap":      bootstrapAt - fetchAt,
		"time.sort.bootstrap": sortBootStrapAt - bootstrapAt,
		"time.candidate":      candidateAt - sortBootStrapAt,
		"time.sort.candidate": sortCandidateAt - candidateAt,
		"time.choose":         time.Now().Unix() - startAt,
	}).Debug("Choose candidates.") */

	return candidates, nil
}

func kickout(stor storage.Storage, candidatesTrie *trie.BatchTrie, delegateTrie *trie.BatchTrie, voteTrie *trie.BatchTrie, candidate byteutils.Hash) error {
	logging.VLog().Debugf("Kickout %s", candidate.String())

	_, err := candidatesTrie.Del(candidate)
	if err != nil && err != storage.ErrKeyNotFound {
		return err
	}
	if err != nil {
		return nil
	}
	iter, err := delegateTrie.Iterator(candidate)
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
		if _, err := delegateTrie.Del(key); err != nil && err != storage.ErrKeyNotFound {
			return err
		}
		bytes, err := voteTrie.Get(delegator)
		if err != nil && err != storage.ErrKeyNotFound {
			return err
		}
		if err == storage.ErrKeyNotFound {
			logging.VLog().WithFields(logrus.Fields{
				"voter":     byteutils.Hex(delegator),
				"candidate": candidate.Hex(),
			}).Debug("Errors: unexpected voter who votes nobody appears in delegate trie")
		}
		if err == nil && byteutils.Equal(bytes, candidate) {
			if _, err := voteTrie.Del(delegator); err != nil && err != storage.ErrKeyNotFound {
				return err
			}
		}
		exist, err = iter.Next()
		if err != nil {
			return err
		}
	}
	// logging.VLog().Info("Kickouted candidate: ", candidate.Hex())
	return nil
}

func (ds *State) kickoutCandidate(candidate byteutils.Hash) error {
	return kickout(ds.chain.Storage(), ds.candidateTrie, ds.delegateTrie, ds.voteTrie, candidate)
}

func (ds *State) kickoutDynasty(dynastyID int64) error {
	// startAt := time.Now().Unix()

	dynastyTrie := ds.dynastyTrie
	iter, err := dynastyTrie.Iterator(nil)
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

	// prepareAt := time.Now().Unix()

	for exist {
		// vStartAt := time.Now().Unix()

		validator := iter.Value()
		key := append(byteutils.FromInt64(dynastyID), validator...)
		bytes, err := ds.mintCntTrie.Get(key)
		if err != nil && err != storage.ErrKeyNotFound {
			return err
		}
		if err != storage.ErrKeyNotFound {
			cnt := byteutils.Int64(bytes)
			if cnt >= DynastyInterval/BlockInterval/DynastySize/2 {
				exist, err = iter.Next()
				if err != nil {
					return err
				}
				continue
			}
		}
		// vCheckAt := time.Now().Unix()

		isActiveBootstrapValidator, err := checkActiveBootstrapValidator(validator, ds.protectTrie, ds.candidateTrie)
		if err != nil {
			return err
		}
		// vCheckProtectAt := time.Now().Unix()

		if !isActiveBootstrapValidator {
			if err := ds.kickoutCandidate(validator); err != nil {
				return err
			}
		}
		// vKickoutAt := time.Now().Unix()

		exist, err = iter.Next()
		if err != nil {
			return err
		}
		// vNextAt := time.Now().Unix()

		/* 		logging.VLog().WithFields(logrus.Fields{
			"time.check.mint":        vCheckAt - vStartAt,
			"time.check.protect":     vCheckProtectAt - vCheckAt,
			"time.validator.kickout": vKickoutAt - vCheckProtectAt,
			"time.validator.next":    vNextAt - vKickoutAt,
			"time.kickout":           vNextAt - vStartAt,
		}).Debug("Kickouted Validator: ", byteutils.Hex(validator)) */
	}

	// endAt := time.Now().Unix()

	/* 	logging.VLog().WithFields(logrus.Fields{
		"time.prepare":         prepareAt - startAt,
		"time.member.kickout":  endAt - prepareAt,
		"time.dynasty.kickout": endAt - startAt,
	}).Debug("Kickouted dynasty: ", dynastyID) */
	return nil
}

func (ds *State) electNextDynastyOnBaseDynasty(worldState state.WorldState, baseDynastyID int64, nextDynastyID int64, baseGenesis bool) error {
	/* 	logging.VLog().WithFields(logrus.Fields{
		"base":            baseDynastyID,
		"next":            nextDynastyID,
		"base is genesis": baseGenesis,
	}).Debug("Try to elect new dynasty") */

	// startAt := time.Now().Unix()

	if baseGenesis {
		baseDynastyID = nextDynastyID - 1
	}
	for i := baseDynastyID; i < nextDynastyID; i++ {
		// electAt := time.Now().Unix()
		// collect candidates
		if !baseGenesis {
			err := ds.kickoutDynasty(i)
			if err != nil {
				return err
			}
		}
		// kickAt := time.Now().Unix()

		votes, err := ds.tallyVotes(worldState)
		if err != nil {
			return err
		}
		// tallyAt := time.Now().Unix()

		candidates, err := ds.chooseCandidates(votes)
		if err != nil {
			return err
		}
		if len(candidates) < SafeSize {
			return ErrTooFewCandidates
		}
		// chooseAt := time.Now().Unix()

		// Top 20 are selected directly
		newDynasty := []string{}
		nextDynastyTrie, err := trie.NewBatchTrie(nil, ds.chain.Storage())
		directSelected := DynastySize - 1
		for i := 0; i < directSelected && i < len(candidates); i++ {
			delegatee := candidates[i].Address.Bytes()
			_, err := nextDynastyTrie.Put(delegatee, delegatee)
			if err != nil {
				return err
			}
			newDynasty = append(newDynasty, candidates[i].Address.String())
		}
		// topAt := time.Now().Unix()

		// The last one is selected randomly
		if len(candidates) > directSelected {
			accStateRoot, err := worldState.AccountsRoot()
			if err != nil {
				return err
			}
			hasher := fnv.New32a()
			hasher.Write(byteutils.FromInt64(nextDynastyID))
			hasher.Write(accStateRoot)
			result := int(hasher.Sum32()) % (len(candidates) - directSelected)
			offset := result + DynastySize - 1
			delegatee := candidates[offset].Address.Bytes()
			_, err = nextDynastyTrie.Put(delegatee, delegatee)
			if err != nil {
				return err
			}
			newDynasty = append(newDynasty, candidates[offset].Address.String())
		}
		// lastAt := time.Now().Unix()

		ds.dynastyTrie = ds.nextDynastyTrie
		ds.nextDynastyTrie = nextDynastyTrie

		/* 		logging.VLog().WithFields(logrus.Fields{
			"dynasty.members":    newDynasty,
			"dynasty.id":         strconv.Itoa(int(i + 1)),
			"time.kickout":       kickAt - electAt,
			"time.tally":         tallyAt - kickAt,
			"time.choose":        chooseAt - tallyAt,
			"time.elect.top":     topAt - chooseAt,
			"time.elect.last":    lastAt - topAt,
			"time.elect.dynasty": time.Now().Unix() - electAt,
		}).Debug("Elected new dynasty") */
	}

	/* 	logging.VLog().WithFields(logrus.Fields{
		"time.elect.over": time.Now().Unix() - startAt,
	}).Debug("Elected Over") */

	return nil
}

// FindProposer for now in given dynasty
func FindProposer(now int64, dynasty *trie.BatchTrie) (proposer byteutils.Hash, err error) {
	offset := now % DynastyInterval
	if offset%BlockInterval != 0 {
		return nil, ErrNotBlockForgTime
	}
	offset /= BlockInterval
	offset %= DynastySize
	delegatees, err := TraverseDynasty(dynasty)
	if err != nil {
		return nil, err
	}

	if int(offset) < len(delegatees) {
		proposer = delegatees[offset]
	} else {
		logging.VLog().WithFields(logrus.Fields{
			"proposer":  proposer,
			"offset":    offset,
			"delegatee": len(delegatees),
		}).Debug("Found Nil Proposer.")
		return nil, ErrFoundNilProposer
	}
	return proposer, nil
}

// Proposer return the current proposer
func (ds *State) Proposer() byteutils.Hash {
	return ds.proposer
}

// TimeStamp return the current timestamp
func (ds *State) TimeStamp() int64 {
	return ds.timeStamp
}

// NextConsensusState return the new state after some seconds elapsed
func (ds *State) NextConsensusState(elapsedSecond int64, worldState state.WorldState) (state.ConsensusState, error) {
	if elapsedSecond%BlockInterval != 0 {
		return nil, ErrNotBlockForgTime
	}

	dynastyTrie, err := ds.dynastyTrie.Clone()
	if err != nil {
		return nil, err
	}
	nextDynastyTrie, err := ds.nextDynastyTrie.Clone()
	if err != nil {
		return nil, err
	}
	delegateTrie, err := ds.delegateTrie.Clone()
	if err != nil {
		return nil, err
	}
	candidateTrie, err := ds.candidateTrie.Clone()
	if err != nil {
		return nil, err
	}
	voteTrie, err := ds.voteTrie.Clone()
	if err != nil {
		return nil, err
	}
	mintCntTrie, err := ds.mintCntTrie.Clone()
	if err != nil {
		return nil, err
	}

	consensusState := &State{
		timeStamp: ds.timeStamp + elapsedSecond,

		dynastyTrie:     dynastyTrie,
		nextDynastyTrie: nextDynastyTrie,
		delegateTrie:    delegateTrie,
		candidateTrie:   candidateTrie,
		voteTrie:        voteTrie,
		mintCntTrie:     mintCntTrie,
		protectTrie:     ds.protectTrie,

		chain:     ds.chain,
		consensus: ds.consensus,
	}

	baseDynastyID := ds.timeStamp / DynastyInterval
	newDynastyID := consensusState.timeStamp / DynastyInterval
	if baseDynastyID < newDynastyID {
		if baseDynastyID+1 < newDynastyID {
			// do not kickout genesis dynasty
			err = consensusState.electNextDynastyOnBaseDynasty(worldState, baseDynastyID, newDynastyID-1, baseDynastyID == 0)
			if err != nil {
				return nil, err
			}
		}
		// do not kickout genesis's next dynasty
		err = consensusState.electNextDynastyOnBaseDynasty(worldState, newDynastyID-1, newDynastyID, baseDynastyID == 0)
		if err != nil {
			return nil, err
		}
	}

	consensusState.proposer, err = FindProposer(consensusState.timeStamp, consensusState.dynastyTrie)
	if err != nil {
		return nil, err
	}
	return consensusState, nil
}

// TraverseDynasty return all members in the dynasty
func TraverseDynasty(dynasty *trie.BatchTrie) ([]byteutils.Hash, error) {
	members := []byteutils.Hash{}
	iter, err := dynasty.Iterator(nil)
	if err != nil && err != storage.ErrKeyNotFound {
		return nil, err
	}
	if err != nil {
		return members, nil
	}
	exist, err := iter.Next()
	for exist {
		members = append(members, iter.Value())
		exist, err = iter.Next()
	}
	return members, nil
}
