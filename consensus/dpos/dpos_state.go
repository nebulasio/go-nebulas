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
	"strconv"
	"time"

	"github.com/nebulasio/go-nebulas/consensus/pb"

	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/core/state"

	"github.com/nebulasio/go-nebulas/common/trie"
	"github.com/nebulasio/go-nebulas/storage"
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
	ErrInvalidDynasty          = errors.New("the size of initial dynasty in genesis block is invalid, should be equal " + strconv.Itoa(DynastySize))
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

	dynastyTrie *trie.Trie // key: delegatee, val: delegatee

	chain     *core.BlockChain
	consensus core.Consensus
}

// NewState create a new dpos state
func (dpos *Dpos) NewState(root *consensuspb.ConsensusRoot, stor storage.Storage) (state.ConsensusState, error) {
	dynastyTrie, err := trie.NewTrie(root.DynastyRoot, stor)
	if err != nil {
		return nil, err
	}

	return &State{
		timeStamp: root.Timestamp,
		proposer:  root.Proposer,

		dynastyTrie: dynastyTrie,

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
	logging.CLog().Info("1111")
	dynastyTrie, err := trie.NewTrie(nil, chain.Storage())
	if err != nil {
		return nil, err
	}
	if len(conf.Consensus.Dpos.Dynasty) < SafeSize {
		return nil, ErrInitialDynastyNotEnough
	}
	if len(conf.Consensus.Dpos.Dynasty) != DynastySize {
		return nil, ErrInvalidDynasty
	}
	for i := 0; i < len(conf.Consensus.Dpos.Dynasty); i++ {
		addr := conf.Consensus.Dpos.Dynasty[i]
		member, err := core.AddressParse(addr)
		if err != nil {
			return nil, err
		}
		v := member.Bytes()
		if _, err = dynastyTrie.Put(v, v); err != nil {
			return nil, err
		}
	}
	return &State{
		timeStamp: core.GenesisTimestamp,
		proposer:  nil,

		dynastyTrie: dynastyTrie,

		chain:     chain,
		consensus: dpos,
	}, nil
}

func (ds *State) String() string {
	proposer := ""
	if ds.proposer != nil {
		proposer = ds.proposer.String()
	}
	return fmt.Sprintf(`{"timestamp": %d, "proposer": "%s", "dynasty": "%s"}`,
		ds.timeStamp,
		proposer,
		byteutils.Hex(ds.dynastyTrie.RootHash()),
	)
}

//replay a dpos
func (ds *State) Replay(done state.ConsensusState) error {
	state := done.(*State)
	if _, err := ds.dynastyTrie.Replay(state.dynastyTrie); err != nil {
		return err
	}
	return nil
}

// Clone a dpos context
func (ds *State) Clone() (state.ConsensusState, error) {
	var err error
	dynastyTrie, err := ds.dynastyTrie.Clone()
	if err != nil {
		return nil, ErrCloneDynastyTrie
	}
	return &State{
		timeStamp: ds.timeStamp,
		proposer:  ds.proposer,

		dynastyTrie: dynastyTrie,

		chain:     ds.chain,
		consensus: ds.consensus,
	}, nil
}

// RootHash hash dpos state
func (ds *State) RootHash() (*consensuspb.ConsensusRoot, error) {
	return &consensuspb.ConsensusRoot{
		DynastyRoot: ds.dynastyTrie.RootHash(),
	}, nil
}

// Dynasty return the current dynasty
func (ds *State) Dynasty() ([]byteutils.Hash, error) {
	return TraverseDynasty(ds.dynastyTrie)
}

// DynastyRoot return the roothash of current dynasty
func (ds *State) DynastyRoot() byteutils.Hash {
	return ds.dynastyTrie.RootHash()
}

// FindProposer for now in given dynasty
func FindProposer(now int64, dynasty *trie.Trie) (proposer byteutils.Hash, err error) {
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

	logging.CLog().Info(delegatees)
	logging.CLog().Info(offset)
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

	logging.CLog().Info("elapsed ", elapsedSecond)
	consensusState := &State{
		timeStamp: ds.timeStamp + elapsedSecond,

		dynastyTrie: dynastyTrie,

		chain:     ds.chain,
		consensus: ds.consensus,
	}

	consensusState.proposer, err = FindProposer(consensusState.timeStamp, consensusState.dynastyTrie)
	if err != nil {
		return nil, err
	}
	return consensusState, nil
}

// TraverseDynasty return all members in the dynasty
func TraverseDynasty(dynasty *trie.Trie) ([]byteutils.Hash, error) {
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
