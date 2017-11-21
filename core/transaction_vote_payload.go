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
	log "github.com/sirupsen/logrus"

	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/nebulasio/go-nebulas/util/byteutils"
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
	ErrInvalidVoteAction              = errors.New("invalid vote action")
	ErrInvalidVotedBlock              = errors.New("cannot vote for a block which isn't created by validators")
	ErrInvalidBlockFromNonValidators  = errors.New("invalid block created by non-validators")
	ErrInvalidVoterNotAcitve          = errors.New("invalid voter, not in current active validators set")
	ErrInvalidChangeSmallerN          = errors.New("invalid change vote, pls vote for a bigger N on a same block")
	ErrInvalidPrepareBiggerViewHeight = errors.New("invalid prepare vote, view height should be smaller than current height")
)

var (
	vote = []byte{1}
)

// VotePayload carry vote information
// 1. Action: prepare, Hash: voted block hash,
//    CurrentHeight: voted block height,
//    ViewHeight: view block height.
// 2. Action: commit, Hash: voted block hash
// 3. Action: change, Hash: voted block hash, Times: change times
// 4. Action: abdicate, Hash: voted dynasty root hash
type VotePayload struct {
	Action        string
	Hash          byteutils.Hash
	CurrentHeight uint64
	ViewHeight    uint64
	Times         uint32
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
func NewPrepareVotePayload(action string, votedBlockHash byteutils.Hash,
	currentHeight uint64, viewHeight uint64) *VotePayload {
	return &VotePayload{
		Action:        action,
		Hash:          votedBlockHash,
		CurrentHeight: currentHeight,
		ViewHeight:    viewHeight,
	}
}

// NewCommitVotePayload create a new commit vote payload
func NewCommitVotePayload(action string, votedBlockHash byteutils.Hash) *VotePayload {
	return &VotePayload{
		Action: action,
		Hash:   votedBlockHash,
	}
}

// NewChangeVotePayload create a new change vote payload
func NewChangeVotePayload(action string, votedBlockHash byteutils.Hash, times uint32) *VotePayload {
	return &VotePayload{
		Action: action,
		Hash:   votedBlockHash,
		Times:  times,
	}
}

// NewAbdicateVotePayload create a new abdicate vote payload
func NewAbdicateVotePayload(action string, dynastyRootHash byteutils.Hash) *VotePayload {
	return &VotePayload{
		Action: action,
		Hash:   dynastyRootHash,
	}
}

// ToBytes serialize payload
func (payload *VotePayload) ToBytes() ([]byte, error) {
	return json.Marshal(payload)
}

type voteContext struct {
	Voter         byteutils.Hash
	ProposerIndex int
	VotedBlock    *Block
	PackedBlock   *Block
	DynastyTrie   *trie.BatchTrie
}

// TODO(roy): make sure the voted blocks are all on canonical chain
func checkAndBuildVoteContext(from []byte, votedBlockHash byteutils.Hash, packedBlock *Block) (*voteContext, error) {
	// find votedBlock on canonical chain
	votedBlock, err := LoadBlockFromStorage(votedBlockHash, packedBlock.storage, packedBlock.txPool)
	if err != nil {
		log.WithFields(log.Fields{
			"func":           "VotePayload.buildVoteContext",
			"VotedBlockHash": votedBlockHash,
		}).Warn("cannot find the voted block in canonical chain.")
		return nil, nil
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
	index := -1
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
	// record vote context
	dynastyRoot, err := votedBlockParent.NextBlockDynastyRoot()
	if err != nil {
		return nil, err
	}
	dynastyTrie, err := trie.NewBatchTrie(dynastyRoot, packedBlock.storage)
	if err != nil {
		return nil, err
	}
	return &voteContext{
		Voter:         from,
		ProposerIndex: index,
		VotedBlock:    votedBlock,
		DynastyTrie:   dynastyTrie,
		PackedBlock:   packedBlock,
	}, nil
}

// TODO(roy): how to slash voters who always vote on non-canonical chain
func (payload *VotePayload) slash(ctx *voteContext) error {
	if ctx == nil {
		return nil
	}

	depositBytes, err := ctx.PackedBlock.depositTrie.Get(ctx.Voter)
	if err != nil {
		return err
	}
	deposit, err := util.NewUint128FromFixedSizeByteSlice(depositBytes)
	if err != nil {
		return err
	}
	if _, err = ctx.PackedBlock.depositTrie.Del(ctx.Voter); err != nil {
		return err
	}
	key := append(ctx.DynastyTrie.RootHash(), ctx.Voter...)
	if _, err = ctx.PackedBlock.validatorsTrie.Del(key); err != nil && err != storage.ErrKeyNotFound {
		return err
	}
	validators, err := traverseValidators(ctx.PackedBlock.validatorsTrie, ctx.DynastyTrie.RootHash())
	if err != nil {
		return err
	}
	size := util.NewUint128FromInt(int64(len(validators)))
	one := util.NewUint128FromInt(int64(1))
	averDynastyReward := util.NewUint128FromBigInt(deposit.Div(deposit.Int, size.Int))
	diff := averDynastyReward.Mul(averDynastyReward.Int, size.Sub(size.Int, one.Int))
	minerReward := util.NewUint128FromBigInt(deposit.Sub(deposit.Int, diff))
	packedBlock := ctx.PackedBlock
	for _, v := range validators {
		packedBlock.addDeposit(v, averDynastyReward)
	}
	coinbase := packedBlock.Coinbase().Bytes()
	packedBlock.addDeposit(coinbase, minerReward)
	return nil
}

// check slashing rule
// 0. cannot Prepare(B,CH,VH) while B.height() != CH
// 1. cannot Prepare(B,CH,VH) before V’s prepare votes > 2/3 * MaxVotes
// 2. if B' is B's child & B' is created by Proposer(N) after B, cannot Prepare(B',CH,VH) after Change(B, N)
// 3. if B belongs to Dynasty D, cannot Prepare in D any more after Abdicate(B)
// 4. if V.height < B'.height < B.height, cannot Prepare(B’,V’) after Prepare(B,V)
// 5. if B1 != B2 but B1.height == B2.height, cannot Prepare(B1, V1) after Prepare(B2, V2)
func (payload *VotePayload) prepare(from []byte, packedBlock *Block) error {
	ctx, err := checkAndBuildVoteContext(from, payload.Hash, packedBlock)
	if err != nil {
		return err
	}
	if payload.ViewHeight >= payload.CurrentHeight {
		return ErrInvalidPrepareBiggerViewHeight
	}
	if ctx != nil {
		// 0. cannot Prepare(B,CH,VH) while B.height() != CH
		if ctx.VotedBlock.Height() != payload.CurrentHeight {
			log.WithFields(log.Fields{
				"func":          "VotePayload.prepare",
				"BlockHash":     payload.Hash,
				"BlockHeight":   ctx.VotedBlock.Height(),
				"CurrentHeight": payload.CurrentHeight,
			}).Error("Slash, Cannot Prepare(B,CH,VH) while B.Height() != CH.")
			return payload.slash(ctx)
		}
		// 1. cannot Prepare(B,CH,VH) before V’s prepare votes > 2/3 * MaxVotes
		if payload.ViewHeight > 1 {
			curBlock := ctx.VotedBlock
			for curBlock.Height() > payload.ViewHeight {
				curBlock, err = curBlock.ParentBlock()
				if err != nil {
					return err
				}
			}
			maxVotes, err := countValidators(curBlock.curDynastyTrie, nil)
			if err != nil {
				return err
			}
			viewPrepareVotes, err := countValidators(packedBlock.prepareVotesTrie, curBlock.Hash())
			if err != nil {
				return err
			}
			if viewPrepareVotes <= maxVotes*2/3 {
				log.WithFields(log.Fields{
					"func":                  "VotePayload.prepare",
					"BlockHash":             payload.Hash,
					"CurrentHeight":         payload.CurrentHeight,
					"ViewHeight":            payload.ViewHeight,
					"ViewHash":              curBlock.Hash(),
					"ViewBlockPrepareVotes": viewPrepareVotes,
					"MaxVotes":              maxVotes,
				}).Error("Slash, Prepare(B,CH,VH) before view block's prepare votes > 2/3 * MaxVotes.")
				return payload.slash(ctx)
			}
		}
		// 2. if B' is B's child & B' is created by Proposer(N) after B, cannot Prepare(B',CH,VH) after Change(B, N)
		changeKey := append(ctx.VotedBlock.ParentHash(), from...)
		bytes, err := packedBlock.changeVotesTrie.Get(changeKey)
		if bytes != nil {
			changeN := byteutils.Uint32(bytes)
			if int(changeN) >= ctx.ProposerIndex {
				log.WithFields(log.Fields{
					"func":          "VotePayload.prepare",
					"BlockHash":     payload.Hash,
					"ChangeN":       changeN,
					"ProposerIndex": ctx.ProposerIndex,
				}).Error("Slash,  Parent(B2) = B1 and B2 is created by Proposer(N), Prepare(B2, CH2, VH2) after Change(B1, CH1, VH1).")
				return payload.slash(ctx)
			}
		}
		if err != storage.ErrKeyNotFound {
			return err
		}
		// 3. if B belongs to Dynasty D, cannot Prepare in D any more after Abdicate(B)
		abdicateKey := append(ctx.VotedBlock.CurDynastyRoot(), from...)
		bytes, err = packedBlock.abdicateVotesTrie.Get(abdicateKey)
		if bytes != nil {
			log.WithFields(log.Fields{
				"func":        "VotePayload.prepare",
				"BlockHash":   payload.Hash,
				"DynastyHash": ctx.VotedBlock.CurDynastyRoot(),
			}).Error("Slash,  B belongs to Dynasty D, cannot Prepare in D any more after Abdicate(B).")
			return payload.slash(ctx)
		}
		if err != storage.ErrKeyNotFound {
			return err
		}
	}
	// 4 and 5
	height := byteutils.FromUint64(payload.CurrentHeight)
	heightKey := append(height, from...)
	bytes, err := packedBlock.heightPrepareVotesTrie.Get(heightKey)
	if bytes != nil {
		log.WithFields(log.Fields{
			"func":          "VotePayload.prepare",
			"BlockHash":     payload.Hash,
			"CurrentHeight": payload.CurrentHeight,
		}).Error(`Slash, V.height < B'.height < B.height, Prepare(B’,V’) after Prepare(B,V); 
		B1 != B2 but B1.height == B2.height, cannot Prepare(B1, V1) after Prepare(B2, V2)`)
		return payload.slash(ctx)
	}
	if err != storage.ErrKeyNotFound {
		return err
	}

	// record vote
	prepareKey := append(payload.Hash, from...)
	if _, err = packedBlock.prepareVotesTrie.Put(prepareKey, vote); err != nil {
		return err
	}
	for i := payload.ViewHeight; i <= payload.CurrentHeight; i++ {
		height := byteutils.FromUint64(i)
		heightKey = append(height, from...)
		if _, err = packedBlock.heightPrepareVotesTrie.Put(heightKey, vote); err != nil {
			return err
		}
	}
	if ctx != nil {
		packedBlock.addDeposit(from, VoteBlockReward)
	}

	log.WithFields(log.Fields{
		"func":      "VotePayload.prepare",
		"BlockHash": payload.Hash,
		"Height":    payload.CurrentHeight,
	}).Info("prepare vote")
	return nil
}

// check slashing rule
// 1. cannot Commit(B) before Prepare(B, CH, VH)
func (payload *VotePayload) commit(from []byte, packedBlock *Block) error {
	ctx, err := checkAndBuildVoteContext(from, payload.Hash, packedBlock)
	if err != nil {
		return err
	}
	// check slash rule
	key := append(payload.Hash, from...)
	_, err = packedBlock.prepareVotesTrie.Get(key)
	if err == storage.ErrKeyNotFound {
		log.WithFields(log.Fields{
			"func":           "VotePayload.commit",
			"VotedBlockHash": payload.Hash,
		}).Error("Slash, Commit(B) before Prepare(B, CH, VH).")
		return payload.slash(ctx)
	}
	if err != nil {
		return err
	}
	// record vote
	_, err = packedBlock.commitVotesTrie.Put(key, vote)
	if err != nil {
		return err
	}
	if ctx != nil {
		packedBlock.addDeposit(from, VoteBlockReward)
	}
	// check finality reward
	if ctx != nil {
		n, err := countValidators(packedBlock.commitVotesTrie, payload.Hash)
		if err != nil {
			return err
		}
		maxVotes, err := countValidators(ctx.VotedBlock.curDynastyTrie, nil)
		if err != nil {
			return err
		}
		if n-1 <= maxVotes*2/3 && n > maxVotes*2/3 {
			validators, err := traverseValidators(packedBlock.validatorsTrie, ctx.VotedBlock.CurDynastyRoot())
			if err != nil {
				return err
			}
			for _, v := range validators {
				packedBlock.addDeposit(v, FinalityBlockReward)
			}
		}
	}

	log.WithFields(log.Fields{
		"func":      "VotePayload.commit",
		"BlockHash": payload.Hash,
	}).Info("commit vote")
	return nil
}

// check slashing rule
// 1. cannot Change(B, N+1) before Change(B, N) > 2/3 * MaxVotes
func (payload *VotePayload) change(from []byte, packedBlock *Block) error {
	ctx, err := checkAndBuildVoteContext(from, payload.Hash, packedBlock)
	if err != nil {
		return err
	}
	// check N is increasing
	key := append(payload.Hash, from...)
	nBytes, err := packedBlock.changeVotesTrie.Get(key)
	if nBytes != nil {
		n := byteutils.Uint32(nBytes)
		if n >= payload.Times {
			return ErrInvalidChangeSmallerN
		}
	}
	// check slash rules
	if ctx != nil && payload.Times > 1 {
		iter, err := packedBlock.changeVotesTrie.Iterator(payload.Hash)
		if err != nil {
			return err
		}
		changeVotes := 0
		for iter.Next() {
			if byteutils.Uint32(iter.Value()) >= payload.Times-1 {
				changeVotes++
			}
		}
		maxVotes, err := countValidators(ctx.DynastyTrie, nil)
		if err != nil {
			return err
		}
		if changeVotes <= maxVotes*2/3 {
			log.WithFields(log.Fields{
				"func":         "VotePayload.change",
				"BlockHash":    payload.Hash,
				"Times":        payload.Times,
				"Change(B, N)": changeVotes,
				"MaxVotes":     maxVotes,
			}).Error("Slash, Change(B, N+1) before Change(B, N) > 2/3 * MaxVotes.")
			return payload.slash(ctx)
		}
	}
	// record vote
	nBytes = byteutils.FromUint32(payload.Times)
	_, err = packedBlock.changeVotesTrie.Put(key, nBytes)
	// TODO(roy): kickout the changed validator if change vote > 2/3
	return err
}

func (payload *VotePayload) abdicate(from []byte, packedBlock *Block) error {
	key := append(payload.Hash, from...)
	_, err := packedBlock.validatorsTrie.Get(key)
	if err != nil {
		return err
	}
	_, err = packedBlock.validatorsTrie.Del(key)
	if err != nil {
		return err
	}
	_, err = packedBlock.abdicateVotesTrie.Put(key, vote)
	return err
}

// Execute the call payload in tx, call a function
func (payload *VotePayload) Execute(tx *Transaction, block *Block) error {
	from := tx.from.Bytes()
	switch payload.Action {
	case PrepareAction:
		return payload.prepare(from, block)
	case CommitAction:
		return payload.commit(from, block)
	case ChangeAction:
		return payload.change(from, block)
	case AbdicateAction:
		return payload.abdicate(from, block)
	default:
		return ErrInvalidVoteAction
	}
}
