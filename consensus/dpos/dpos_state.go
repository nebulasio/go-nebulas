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
	"fmt"
	"time"

	consensuspb "github.com/nebulasio/go-nebulas/consensus/pb"

	"github.com/nebulasio/go-nebulas/core"
	corepb "github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/core/state"

	"github.com/nebulasio/go-nebulas/common/trie"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

// State carry context in dpos consensus
type State struct {
	timestamp int64
	proposer  byteutils.Hash

	dynastyTrie *trie.Trie // key: delegatee, val: delegatee

	chain     *core.BlockChain
	consensus core.Consensus
}

// NewState create a new dpos state
func (dpos *Dpos) NewState(root *consensuspb.ConsensusRoot, stor storage.Storage, needChangeLog bool) (state.ConsensusState, error) {
	var dynastyRoot byteutils.Hash
	if root != nil {
		dynastyRoot = root.DynastyRoot
	}
	dynastyTrie, err := trie.NewTrie(dynastyRoot, stor, needChangeLog)
	if err != nil {
		return nil, err
	}

	return &State{
		timestamp: root.Timestamp,
		proposer:  root.Proposer,

		dynastyTrie: dynastyTrie,

		chain:     dpos.chain,
		consensus: dpos,
	}, nil
}

// CheckTimeout check whether the block is timeout
func (dpos *Dpos) CheckTimeout(block *core.Block) bool {
	nowInMs := time.Now().Unix() * SecondInMs
	blockTimeInMs := block.Timestamp() * SecondInMs
	if nowInMs < blockTimeInMs {
		logging.VLog().WithFields(logrus.Fields{
			"block": block,
			"now":   nowInMs,
			"diff":  blockTimeInMs - nowInMs,
			"err":   "timeout - future block",
		}).Warn("Found a future block.")
		return false
	}
	behindInMs := nowInMs - blockTimeInMs
	if behindInMs > AcceptedNetWorkDelayInMs {
		logging.VLog().WithFields(logrus.Fields{
			"block": block,
			"now":   nowInMs,
			"diff":  behindInMs,
			"limit": AcceptedNetWorkDelayInMs,
			"err":   "timeout - expired block",
		}).Warn("Found a expired block.")
		return true
	}
	return false
}

// GenesisConsensusState create a new genesis dpos state
func (dpos *Dpos) GenesisConsensusState(chain *core.BlockChain, conf *corepb.Genesis) (state.ConsensusState, error) {
	dynastyTrie, err := trie.NewTrie(nil, chain.Storage(), false)
	if err != nil {
		return nil, err
	}
	if len(conf.Consensus.Dpos.Dynasty) < ConsensusSize {
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
		timestamp: core.GenesisTimestamp,
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
		ds.timestamp,
		proposer,
		byteutils.Hex(ds.dynastyTrie.RootHash()),
	)
}

// Replay a dpos
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
		timestamp: ds.timestamp,
		proposer:  ds.proposer,

		dynastyTrie: dynastyTrie,

		chain:     ds.chain,
		consensus: ds.consensus,
	}, nil
}

// RootHash hash dpos state
func (ds *State) RootHash() *consensuspb.ConsensusRoot {
	return &consensuspb.ConsensusRoot{
		DynastyRoot: ds.dynastyTrie.RootHash(),
		Timestamp:   ds.TimeStamp(),
		Proposer:    ds.Proposer(),
	}
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
func FindProposer(now int64, miners []byteutils.Hash) (proposer byteutils.Hash, err error) {
	nowInMs := now * SecondInMs
	offsetInMs := nowInMs % DynastyIntervalInMs
	if (offsetInMs % BlockIntervalInMs) != 0 {
		return nil, ErrNotBlockForgTime
	}
	offset := offsetInMs / BlockIntervalInMs
	offset %= DynastySize

	if offset >= 0 && int(offset) < len(miners) {
		proposer = miners[offset]
	} else {
		logging.VLog().WithFields(logrus.Fields{
			"proposer":  proposer,
			"offset":    offset,
			"delegatee": len(miners),
		}).Warn("Found Nil Proposer.")
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
	return ds.timestamp
}

// NextConsensusState return the new state after some seconds elapsed
func (ds *State) NextConsensusState(elapsedSecond int64, worldState state.WorldState) (state.ConsensusState, error) {
	elapsedSecondInMs := elapsedSecond * SecondInMs
	if elapsedSecondInMs <= 0 || elapsedSecondInMs%BlockIntervalInMs != 0 {
		return nil, ErrNotBlockForgTime
	}

	dpos, ok := ds.consensus.(*Dpos)
	if !ok {
		logging.VLog().WithFields(logrus.Fields{
			"timestamp": ds.timestamp,
		}).Fatal("Type conversion failed, unexpected error.")
	}
	nextTimestamp := ds.timestamp + elapsedSecond
	dynastyTrie, err := dpos.dynasty.getDynasty(nextTimestamp)
	if err != nil {
		return nil, err
	}

	consensusState := &State{
		timestamp: ds.timestamp + elapsedSecond,

		dynastyTrie: dynastyTrie,

		chain:     ds.chain,
		consensus: ds.consensus,
	}

	miners, err := TraverseDynasty(dynastyTrie)
	if err != nil {
		return nil, err
	}
	consensusState.proposer, err = FindProposer(consensusState.timestamp, miners)
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
