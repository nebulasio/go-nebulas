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
	// "strconv"

	"github.com/nebulasio/go-nebulas/common/trie"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/core/state"
	"github.com/nebulasio/go-nebulas/crypto/sha3"
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
	// logging.VLog().Debug("DposContext Begin.")
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
	// logging.VLog().Debug("DposContext Commit.")
}

// RollBack a batch task
func (dc *DposContext) RollBack() {
	dc.delegateTrie.RollBack()
	dc.dynastyTrie.RollBack()
	dc.nextDynastyTrie.RollBack()
	dc.candidateTrie.RollBack()
	dc.voteTrie.RollBack()
	dc.mintCntTrie.RollBack()
	// logging.VLog().Debug("DposContext RollBack.")
}

// Clone a dpos context
func (dc *DposContext) Clone() (*DposContext, error) {
	var err error
	context, err := NewDposContext(dc.storage)
	if err != nil {
		return nil, err
	}
	if context.dynastyTrie, err = dc.dynastyTrie.Clone(); err != nil {
		return nil, ErrCloneDynastyTrie
	}
	if context.nextDynastyTrie, err = dc.nextDynastyTrie.Clone(); err != nil {
		return nil, ErrCloneNextDynastyTrie
	}
	if context.delegateTrie, err = dc.delegateTrie.Clone(); err != nil {
		return nil, ErrCloneDelegateTrie
	}
	if context.candidateTrie, err = dc.candidateTrie.Clone(); err != nil {
		return nil, ErrCloneCandidatesTrie
	}
	if context.voteTrie, err = dc.voteTrie.Clone(); err != nil {
		return nil, ErrCloneVoteTrie
	}
	if context.mintCntTrie, err = dc.mintCntTrie.Clone(); err != nil {
		return nil, ErrCloneMintCntTrie
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
	Proposer        byteutils.Hash
	DynastyTrie     *trie.BatchTrie
	NextDynastyTrie *trie.BatchTrie
	DelegateTrie    *trie.BatchTrie
	CandidateTrie   *trie.BatchTrie
	ProtectTrie     *trie.BatchTrie
	VoteTrie        *trie.BatchTrie
	MintCntTrie     *trie.BatchTrie
	Accounts        state.AccountState
	Storage         storage.Storage
}

func (dc *DynastyContext) tallyVotes() (map[string]*util.Uint128, error) {
	votes := make(map[string]*util.Uint128)
	delegate := dc.DelegateTrie
	candidates := dc.CandidateTrie
	accounts := dc.Accounts
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
			delegator, err := AddressParseFromBytes(iterDelegate.Value())
			if err != nil {
				return nil, err
			}
			score, ok := votes[delegatee.String()]
			if !ok {
				score = util.NewUint128()
			}
			acc, err := accounts.GetOrCreateUserAccount(delegator.Bytes())
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
	Address *Address
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

func (dc *DynastyContext) chooseCandidates(votes map[string]*util.Uint128) (Candidates, error) {
	// startAt := time.Now().Unix()
	// active bootstrap validators
	var bootstrapCandidates Candidates
	activeBootstrapValidators, err := fetchActiveBootstapValidators(dc.ProtectTrie, dc.CandidateTrie)
	if err != nil {
		return nil, err
	}
	// fetchAt := time.Now().Unix()
	for i := 0; i < len(activeBootstrapValidators); i++ {
		address, err := AddressParseFromBytes(activeBootstrapValidators[i])
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
		addr, err := AddressParse(k)
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

func (dc *DposContext) kickoutCandidate(candidate byteutils.Hash) error {
	return kickout(dc.storage, dc.candidateTrie, dc.delegateTrie, dc.voteTrie, candidate)
}

func (dc *DynastyContext) kickoutCandidate(candidate byteutils.Hash) error {
	return kickout(dc.Storage, dc.CandidateTrie, dc.DelegateTrie, dc.VoteTrie, candidate)
}

func (dc *DynastyContext) kickoutDynasty(dynastyID int64) error {
	// startAt := time.Now().Unix()

	dynastyTrie := dc.DynastyTrie
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
		bytes, err := dc.MintCntTrie.Get(key)
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

		isActiveBootstrapValidator, err := checkActiveBootstrapValidator(validator, dc.ProtectTrie, dc.CandidateTrie)
		if err != nil {
			return err
		}
		// vCheckProtectAt := time.Now().Unix()

		if !isActiveBootstrapValidator {
			if err := dc.kickoutCandidate(validator); err != nil {
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

func (dc *DynastyContext) electNextDynastyOnBaseDynasty(baseDynastyID int64, nextDynastyID int64, baseGenesis bool) error {
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
			err := dc.kickoutDynasty(i)
			if err != nil {
				return err
			}
		}
		// kickAt := time.Now().Unix()

		votes, err := dc.tallyVotes()
		if err != nil {
			return err
		}
		// tallyAt := time.Now().Unix()

		candidates, err := dc.chooseCandidates(votes)
		if err != nil {
			return err
		}
		if len(candidates) < SafeSize {
			return ErrTooFewCandidates
		}
		// chooseAt := time.Now().Unix()

		// Top 20 are selected directly
		newDynasty := []string{}
		nextDynastyTrie, err := trie.NewBatchTrie(nil, dc.Storage)
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
			accState, err := dc.Accounts.RootHash()
			if err != nil {
				return err
			}
			hasher := fnv.New32a()
			hasher.Write(byteutils.FromInt64(nextDynastyID))
			hasher.Write(accState)
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

		dc.DynastyTrie = dc.NextDynastyTrie
		dc.NextDynastyTrie = nextDynastyTrie

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
func GenesisDynastyContext(chain *BlockChain, conf *corepb.Genesis) (*DynastyContext, error) {
	dynastyTrie, err := trie.NewBatchTrie(nil, chain.storage)
	if err != nil {
		return nil, err
	}
	delegateTrie, err := trie.NewBatchTrie(nil, chain.storage)
	if err != nil {
		return nil, err
	}
	candidateTrie, err := trie.NewBatchTrie(nil, chain.storage)
	if err != nil {
		return nil, err
	}
	voteTrie, err := trie.NewBatchTrie(nil, chain.storage)
	if err != nil {
		return nil, err
	}
	mintTrie, err := trie.NewBatchTrie(nil, chain.storage)
	if err != nil {
		return nil, err
	}
	if len(conf.Consensus.Dpos.Dynasty) < SafeSize {
		return nil, ErrInitialDynastyNotEnough
	}
	for i := 0; i < len(conf.Consensus.Dpos.Dynasty); i++ {
		addr := conf.Consensus.Dpos.Dynasty[i]
		member, err := AddressParse(addr)
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
	return &DynastyContext{
		TimeStamp:       GenesisTimestamp,
		DynastyTrie:     dynastyTrie,
		NextDynastyTrie: nextDynastyTrie,
		DelegateTrie:    delegateTrie,
		CandidateTrie:   candidateTrie,
		ProtectTrie:     protectTrie,
		MintCntTrie:     mintTrie,
		VoteTrie:        voteTrie,
	}, nil
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
		}).Debug("Find Nil Proposer.")
		return nil, ErrFoundNilProposer
	}
	return proposer, nil
}

// NextDynastyContext when some seconds elapsed
func (block *Block) NextDynastyContext(chain *BlockChain, elapsedSecond int64) (*DynastyContext, error) {
	if elapsedSecond%BlockInterval != 0 {
		return nil, ErrNotBlockForgTime
	}

	dynastyTrie, err := block.dposContext.dynastyTrie.Clone()
	if err != nil {
		return nil, err
	}
	nextDynastyTrie, err := block.dposContext.nextDynastyTrie.Clone()
	if err != nil {
		return nil, err
	}
	delegateTrie, err := block.dposContext.delegateTrie.Clone()
	if err != nil {
		return nil, err
	}
	candidateTrie, err := block.dposContext.candidateTrie.Clone()
	if err != nil {
		return nil, err
	}
	protectTrie, err := chain.genesisBlock.dposContext.candidateTrie.Clone()
	if err != nil {
		return nil, err
	}
	voteTrie, err := block.dposContext.voteTrie.Clone()
	if err != nil {
		return nil, err
	}
	mintCntTrie, err := block.dposContext.mintCntTrie.Clone()
	if err != nil {
		return nil, err
	}

	context := &DynastyContext{
		TimeStamp:       block.header.timestamp + elapsedSecond,
		DynastyTrie:     dynastyTrie,
		NextDynastyTrie: nextDynastyTrie,
		DelegateTrie:    delegateTrie,
		CandidateTrie:   candidateTrie,
		ProtectTrie:     protectTrie,
		VoteTrie:        voteTrie,
		MintCntTrie:     mintCntTrie,
		Accounts:        block.accState,
		Storage:         block.storage,
	}

	baseDynastyID := block.header.timestamp / DynastyInterval
	newDynastyID := context.TimeStamp / DynastyInterval
	if baseDynastyID < newDynastyID {
		if baseDynastyID+1 < newDynastyID {
			// do not kickout genesis dynasty
			err = context.electNextDynastyOnBaseDynasty(baseDynastyID, newDynastyID-1, baseDynastyID == 0)
			if err != nil {
				return nil, err
			}
		}
		// do not kickout genesis's next dynasty
		err = context.electNextDynastyOnBaseDynasty(newDynastyID-1, newDynastyID, baseDynastyID == 0)
		if err != nil {
			return nil, err
		}
	}

	context.Proposer, err = FindProposer(context.TimeStamp, context.DynastyTrie)
	if err != nil {
		return nil, err
	}
	return context, nil
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
