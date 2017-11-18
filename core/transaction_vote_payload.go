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
	"errors"

	"github.com/nebulasio/go-nebulas/common/trie"

	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	log "github.com/sirupsen/logrus"
)

// Action types
const (
	PrepareAction  = "prepare"
	CommitAction   = "commit"
	ChangeAction   = "change"
	AbdicateAction = "abdicate"
)

// Errors constants
var (
	ErrInvalidVoteAction             = errors.New("invalid vote action")
	ErrDupVoteAction                 = errors.New("different vote, but same action")
	ErrCommitBeforePrepare           = errors.New("cannot commit before prepare a block")
	ErrInvalidVotedBlock             = errors.New("cannot vote for a block which isn't created by validators")
	ErrInvalidBlockFromNonValidators = errors.New("invalid block created by non-validators")
	ErrInvalidVoterNotAcitve         = errors.New("invalid voter, not in current active validators set")
)

var (
	vote = []byte{1}
)

// VotePayload carry vote information
// 1. action: prepare, block: current block hash, view: based block hash
// 2. action: commit, block: current block hash
// 3. action: change, block: parent block, times: change times
// 4. action: abdicate, block: parent block
type VotePayload struct {
	Action    string
	BlockHash byteutils.Hash
	ViewHash  byteutils.Hash
	Times     uint32
}

// LoadVotePayload from bytes
func LoadVotePayload(bytes []byte) (*VotePayload, error) {
	payload := &VotePayload{}
	if err := json.Unmarshal(bytes, payload); err != nil {
		return nil, err
	}
	return payload, nil
}

// NewPrepareVotePayload create a new prepare vote payload
func NewPrepareVotePayload(action string, block byteutils.Hash, view byteutils.Hash) *VotePayload {
	return &VotePayload{
		Action:    action,
		BlockHash: block,
		ViewHash:  view,
	}
}

// NewCommitVotePayload create a new commit vote payload
func NewCommitVotePayload(action string, block byteutils.Hash) *VotePayload {
	return &VotePayload{
		Action:    action,
		BlockHash: block,
	}
}

// NewChangeVotePayload create a new change vote payload
func NewChangeVotePayload(action string, block byteutils.Hash, times uint32) *VotePayload {
	return &VotePayload{
		Action:    action,
		BlockHash: block,
		Times:     times,
	}
}

// NewAbdicateVotePayload create a new abdicate vote payload
func NewAbdicateVotePayload(action string, block byteutils.Hash) *VotePayload {
	return &VotePayload{
		Action:    action,
		BlockHash: block,
	}
}

// ToBytes serialize payload
func (payload *VotePayload) ToBytes() ([]byte, error) {
	return json.Marshal(payload)
}

type voteContext struct {
	Voter             byteutils.Hash
	ProposerIndex     int
	VotedBlock        *Block
	VotedBlockParent  *Block
	PackedBlock       *Block
	RewardDynastyRoot byteutils.Hash
}

// block is valid, created by validators
// voter is valid, in validators
// vote is valid, not dup
// return ctx, nil. valid. voter is the proposer at position N.
// return nil, err. invalid.
func (payload *VotePayload) checkVoteValid(from []byte, packedBlock *Block) (*voteContext, error) {
	index := -1
	votedBlock, err := LoadBlockFromStorage(payload.BlockHash, packedBlock.storage, packedBlock.txPool)
	if err != nil {
		return nil, err
	}
	votedBlockParent, err := LoadBlockFromStorage(votedBlock.ParentHash(), votedBlock.storage, votedBlock.txPool)
	if err != nil {
		return nil, err
	}
	// check block & voter valid
	validators, err := votedBlockParent.NextBlockSortedValidators()
	if err != nil {
		return nil, err
	}
	validBlock := false
	validVoter := false
	for k, v := range validators {
		if v.Equals(votedBlock.Coinbase().Bytes()) {
			validBlock = true
			index = k
		}
		if v.Equals(from) {
			validVoter = true
		}
	}
	if !validBlock {
		return nil, ErrInvalidBlockFromNonValidators
	}
	if !validVoter {
		return nil, ErrInvalidVoterNotAcitve
	}
	// check dup action
	var voteTrie *trie.BatchTrie
	var key byteutils.Hash
	switch payload.Action {
	case PrepareAction:
		key = append(payload.BlockHash, from...)
		voteTrie = packedBlock.prepareVotesTrie
	case CommitAction:
		key = append(payload.BlockHash, from...)
		voteTrie = packedBlock.commitVotesTrie
	case ChangeAction:
		key = append(payload.BlockHash, byteutils.FromUint32(payload.Times)...)
		key = append(key, from...)
		voteTrie = packedBlock.changeVotesTrie
	case AbdicateAction:
		key = append(votedBlock.CurDynastyRoot(), from...)
		voteTrie = packedBlock.abdicateVotesTrie
	default:
		return nil, ErrInvalidVoteAction
	}
	_, err = voteTrie.Get(key)
	if err == nil {
		return nil, ErrDupVoteAction
	}
	if err != storage.ErrKeyNotFound {
		return nil, err
	}
	return &voteContext{
		Voter:            from,
		ProposerIndex:    index,
		VotedBlock:       votedBlock,
		VotedBlockParent: votedBlockParent,
		PackedBlock:      packedBlock,
	}, nil
}

// prove there is a chain from the ancestor block to the given block
// return nil, err. invalid tx.
// return nil, nil. cannot find the chain.
// return result, nil. find the chain.
// chain. given block -> ... -> ancestor block -> ancestor block hash
func proveAncestorOfBlock(ancestorHash byteutils.Hash, blockHash byteutils.Hash, miner *Block) ([]*Block, error) {
	result := []*Block{}
	ans, err := LoadBlockFromStorage(ancestorHash, miner.storage, miner.txPool)
	if err != nil {
		return nil, err
	}
	cur, err := LoadBlockFromStorage(blockHash, miner.storage, miner.txPool)
	if err != nil {
		return nil, err
	}
	if ans.Height() > cur.Height() {
		return nil, nil
	}
	result = append(result, cur)
	for ans.Height() < cur.Height() {
		cur, err = LoadBlockFromStorage(cur.ParentHash(), miner.storage, miner.txPool)
		if err != nil {
			return nil, err
		}
		result = append(result, cur)
	}
	if ans.Hash().Equals(cur.Hash()) {
		return result, nil
	}
	return nil, nil
}

func (payload *VotePayload) slash(voteCtx *voteContext) error {
	depositBytes, err := voteCtx.PackedBlock.depositTrie.Get(voteCtx.Voter)
	if err != nil {
		return err
	}
	deposit, err := util.NewUint128FromFixedSizeByteSlice(depositBytes)
	if err != nil {
		return err
	}
	if _, err = voteCtx.PackedBlock.depositTrie.Del(voteCtx.Voter); err != nil {
		return err
	}
	validators, err := traverseValidators(voteCtx.PackedBlock.validatorsTrie, voteCtx.RewardDynastyRoot)
	if err != nil {
		return err
	}
	size := util.NewUint128FromInt(int64(len(validators)))
	one := util.NewUint128FromInt(int64(1))
	averDynastyReward := util.NewUint128FromBigInt(deposit.Div(deposit.Int, size.Int))
	diff := averDynastyReward.Mul(averDynastyReward.Int, size.Sub(size.Int, one.Int))
	minerReward := util.NewUint128FromBigInt(deposit.Sub(deposit.Int, diff))
	packedBlock := voteCtx.PackedBlock
	for _, v := range validators {
		if !v.Equals(voteCtx.Voter) {
			packedBlock.accState.GetOrCreateUserAccount(v).AddBalance(averDynastyReward)
		}
	}
	coinbase := packedBlock.Coinbase().Bytes()
	packedBlock.accState.GetOrCreateUserAccount(coinbase).AddBalance(minerReward)
	return nil
}

// check slashing rule
// 0. V must be B's ancestor in Prepare(B, V)
// 1. cannot Prepare(B,V) before V’s prepare votes > 2/3 * MaxVotes
// 2. if V.height < B'.height < B.height, cannot Prepare(B’,V’) after Prepare(B,V)
// 3. if B1 != B2 but B1.height == B2.height, cannot Prepare(B1, V1) after Prepare(B2, V2)
// 4. if B' is B's child & B' is created by Proposer(N) after B, cannot Prepare(B', V) after Change(B, N)
// 5. if B belongs to Dynasty D, cannot Prepare in D any more after Abdicate(B)
func (payload *VotePayload) prepare(voteCtx *voteContext) error {
	// 0
	chain, err := proveAncestorOfBlock(payload.ViewHash, payload.BlockHash, voteCtx.PackedBlock)
	if err != nil {
		return err
	}
	if chain == nil {
		log.WithFields(log.Fields{
			"func":      "VotePayload.prepare",
			"BlockHash": payload.BlockHash.Hex(),
			"ViewHash":  payload.ViewHash.Hex(),
		}).Warn("Slash, Cannot find a chain from ViewHash to BlockHash.")
		return payload.slash(voteCtx)
	}
	// 1
	votedBlock := chain[0]
	voteCtx.RewardDynastyRoot = voteCtx.VotedBlock.CurDynastyRoot()
	maxVotes, err := DynastyMaxVotes(votedBlock.CurDynastyRoot(), votedBlock.storage)
	if err != nil {
		return err
	}
	viewPrepareVotes, err := countValidators(voteCtx.PackedBlock.prepareVotesTrie, payload.ViewHash)
	if err != nil {
		return err
	}
	if viewPrepareVotes < 2/3*maxVotes {
		log.WithFields(log.Fields{
			"func":       "VotePayload.prepare",
			"BlockHash":  payload.BlockHash.Hex(),
			"ViewHash":   payload.ViewHash.Hex(),
			"Prepare(V)": viewPrepareVotes,
			"MaxVotes":   maxVotes,
		}).Error("Slash, Prepare(B,V) before V’s prepare votes > 2/3 * MaxVotes.")
		return payload.slash(voteCtx)
	}
	// 2 and 3
	height := byteutils.FromUint64(votedBlock.Height())
	heightKey := append(height, voteCtx.Voter...)
	_, err = voteCtx.PackedBlock.heightPrepareVotesTrie.Get(heightKey)
	if err == nil {
		log.WithFields(log.Fields{
			"func":      "VotePayload.prepare",
			"BlockHash": payload.BlockHash.Hex(),
			"ViewHash":  payload.ViewHash.Hex(),
			"Height":    chain[0].Height(),
		}).Error(`Slash, V.height < B'.height < B.height, Prepare(B’,V’) after Prepare(B,V); 
		B1 != B2 but B1.height == B2.height, cannot Prepare(B1, V1) after Prepare(B2, V2)`)
		return payload.slash(voteCtx)
	}
	if err != storage.ErrKeyNotFound {
		return err
	}
	// 4
	changeKey := append(voteCtx.VotedBlockParent.Hash(), voteCtx.Voter...)
	nBytes, err := voteCtx.PackedBlock.changeVotesTrie.Get(changeKey)
	if err == nil {
		n := byteutils.Uint32(nBytes)
		if int(n) >= voteCtx.ProposerIndex {
			log.WithFields(log.Fields{
				"func":          "VotePayload.prepare",
				"BlockHash":     payload.BlockHash.Hex(),
				"ViewHash":      payload.ViewHash.Hex(),
				"N":             n,
				"ProposerIndex": voteCtx.ProposerIndex,
			}).Error("Slash,  Parent(B2) = B1 and B2 is created by Proposer(N), Prepare(B2, V) after Change(B1, N).")
			return payload.slash(voteCtx)
		}
	}
	if err != storage.ErrKeyNotFound {
		return err
	}
	// 5
	dynastyRoot, err := voteCtx.VotedBlockParent.NextBlockDynastyRoot()
	if err != nil {
		return err
	}
	abdicateKey := append(dynastyRoot, voteCtx.Voter...)
	_, err = voteCtx.PackedBlock.abdicateVotesTrie.Get(abdicateKey)
	if err == nil {
		log.WithFields(log.Fields{
			"func":      "VotePayload.prepare",
			"BlockHash": payload.BlockHash.Hex(),
			"ViewHash":  payload.ViewHash.Hex(),
		}).Error("Slash,  B belongs to Dynasty D, cannot Prepare in D any more after Abdicate(B).")
		return payload.slash(voteCtx)
	}
	if err != storage.ErrKeyNotFound {
		return err
	}

	prepareKey := append(voteCtx.PackedBlock.Hash(), voteCtx.Voter...)
	if _, err = voteCtx.PackedBlock.prepareVotesTrie.Put(prepareKey, vote); err != nil {
		return err
	}
	for _, v := range chain {
		height := byteutils.FromUint64(v.Height())
		heightKey = append(height, voteCtx.Voter...)
		if _, err = voteCtx.PackedBlock.heightPrepareVotesTrie.Put(heightKey, vote); err != nil {
			return err
		}
	}
	return nil
}

// check slashing rule
// 1. cannot Commit(B) before Prepare(B, V)
func (payload *VotePayload) commit(voteCtx *voteContext) error {
	key := append(voteCtx.VotedBlock.Hash(), voteCtx.Voter...)
	_, err := voteCtx.PackedBlock.prepareVotesTrie.Get(key)
	if err == nil {
		_, err = voteCtx.PackedBlock.commitVotesTrie.Put(key, vote)
		return err
	}
	if err != storage.ErrKeyNotFound {
		return err
	}
	log.WithFields(log.Fields{
		"func":      "VotePayload.commit",
		"BlockHash": payload.BlockHash.Hex(),
	}).Error("Slash, Commit(B) before Prepare(B, V).")
	voteCtx.RewardDynastyRoot = voteCtx.VotedBlock.CurDynastyRoot()
	return payload.slash(voteCtx)
}

// check slashing rule
// 1. cannot Change(B, N+1) before Change(B, N) > 2/3 * MaxVotes
func (payload *VotePayload) change(voteCtx *voteContext) error {
	if payload.Times > 1 {
		prefix := voteCtx.VotedBlockParent.Hash()
		iter, err := voteCtx.PackedBlock.changeVotesTrie.Iterator(prefix)
		if err != nil {
			return err
		}
		changeVotes := 0
		for iter.Next() {
			if byteutils.Uint32(iter.Value()) >= payload.Times-1 {
				changeVotes++
			}
		}
		if voteCtx.RewardDynastyRoot, err = voteCtx.VotedBlockParent.NextBlockDynastyRoot(); err != nil {
			return err
		}
		maxVotes, err := DynastyMaxVotes(voteCtx.RewardDynastyRoot, voteCtx.VotedBlock.storage)
		if err != nil {
			return err
		}
		if changeVotes < 2/3*maxVotes {
			log.WithFields(log.Fields{
				"func":         "VotePayload.change",
				"BlockHash":    payload.BlockHash.Hex(),
				"Times":        payload.Times,
				"Change(B, N)": changeVotes,
				"MaxVotes":     maxVotes,
			}).Error("Slash, Change(B, N+1) before Change(B, N) > 2/3 * MaxVotes.")
			if err != nil {
				return err
			}
			return payload.slash(voteCtx)
		}
	}
	key := append(voteCtx.VotedBlock.Hash(), byteutils.FromUint32(payload.Times)...)
	key = append(key, voteCtx.Voter...)
	_, err := voteCtx.PackedBlock.changeVotesTrie.Put(key, vote)
	return err
}

func (payload *VotePayload) abdicate(voteCtx *voteContext) error {
	key := append(voteCtx.VotedBlock.CurDynastyRoot(), voteCtx.Voter...)
	_, err := voteCtx.PackedBlock.abdicateVotesTrie.Put(key, vote)
	return err
}

// Execute the call payload in tx, call a function
func (payload *VotePayload) Execute(tx *Transaction, block *Block) error {
	voteCtx, err := payload.checkVoteValid(tx.from.Bytes(), block)
	if err != nil {
		return err
	}
	switch payload.Action {
	case PrepareAction:
		return payload.prepare(voteCtx)
	case CommitAction:
		return payload.commit(voteCtx)
	case ChangeAction:
		return payload.change(voteCtx)
	case AbdicateAction:
		return payload.abdicate(voteCtx)
	default:
		return ErrInvalidVoteAction
	}
}
