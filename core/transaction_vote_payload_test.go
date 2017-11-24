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
	"testing"

	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/stretchr/testify/assert"
)

func NewBlockWithValidDynasty(t *testing.T) ([]byteutils.Hash, *Block) {
	storage, _ := storage.NewMemoryStorage()
	genesis := NewGenesisBlock(0, storage, nil)
	validators, _ := genesis.NextBlockSortedValidators()
	coinbase := &Address{validators[0]}
	block1 := NewBlock(genesis.header.chainID, coinbase, genesis)
	block1.Seal()
	validators, _ = block1.NextBlockSortedValidators()
	coinbase = &Address{validators[0]}
	block2 := NewBlock(block1.header.chainID, coinbase, block1)
	block2.Seal()

	validators, _ = block2.NextBlockSortedValidators()
	assert.Nil(t, storeBlockToStorage(genesis))
	assert.Nil(t, storeBlockToStorage(block1))
	assert.Nil(t, storeBlockToStorage(block2))
	return validators, block2
}

func TestVotePayload_RightFlow(t *testing.T) {
	validators, block := NewBlockWithValidDynasty(t)
	size := len(validators)
	zero := util.NewUint128()
	coinbase := &Address{validators[0]}
	newBlock := NewBlock(block.header.chainID, coinbase, block)
	newBlock.begin()
	for i := 0; i < size; i++ {
		preparePayload, err := NewPrepareVotePayload(
			PrepareAction, block.Hash(), block.Height(), 1).ToBytes()
		assert.Nil(t, err)
		prepareTx := NewTransaction(
			block.header.chainID, &Address{validators[i]}, &Address{validators[i]},
			zero, 1, TxPayloadVoteType, preparePayload)
		giveback, err := newBlock.executeTransaction(prepareTx)
		assert.Equal(t, giveback, false)
		assert.Nil(t, err)
	}
	for i := 0; i < size; i++ {
		commitPayload, err := NewCommitVotePayload(CommitAction, block.Hash()).ToBytes()
		assert.Nil(t, err)
		commitTx := NewTransaction(
			block.header.chainID, &Address{validators[i]}, &Address{validators[i]},
			zero, 2, TxPayloadVoteType, commitPayload)
		giveback, err := newBlock.executeTransaction(commitTx)
		assert.Equal(t, giveback, false)
		assert.Nil(t, err)
	}
	newBlock.commit()
	newBlock.Seal()
	newScores, err := CalScoresOnChain(newBlock)
	assert.Nil(t, err)
	blockScores, err := CalScoresOnChain(block)
	assert.Nil(t, err)
	assert.Equal(t, newScores, uint64(3))
	assert.Equal(t, blockScores, uint64(0))
	reward := util.NewUint128()
	reward.Add(reward.Int, FinalityBlockReward.Int)
	reward.Add(reward.Int, StandardDeposit.Int)
	reward.Sub(reward.Int, VoteBlockReward.Int)
	reward.Sub(reward.Int, VoteBlockReward.Int)
	for i := 0; i < size; i++ {
		depositBytes, err := newBlock.depositTrie.Get(validators[i])
		assert.Nil(t, err)
		deposit, err := util.NewUint128FromFixedSizeByteSlice(depositBytes)
		assert.Nil(t, err)
		assert.Equal(t, deposit, reward)
	}
}

func TestVotePayload_PrepareWrongHeight(t *testing.T) {
	validators, block := NewBlockWithValidDynasty(t)
	zero := util.NewUint128()
	coinbase := &Address{validators[0]}
	newBlock := NewBlock(block.header.chainID, coinbase, block)
	newBlock.begin()
	// wrong current height
	preparePayload, err := NewPrepareVotePayload(
		PrepareAction, block.Hash(), block.Height()+1, 1).ToBytes()
	assert.Nil(t, err)
	prepareTx := NewTransaction(
		block.header.chainID, &Address{validators[1]}, &Address{validators[1]},
		zero, 1, TxPayloadVoteType, preparePayload)
	giveback, err := newBlock.executeTransaction(prepareTx)
	assert.Equal(t, giveback, false)
	assert.Nil(t, err)
	_, err = newBlock.depositTrie.Get(validators[1])
	assert.NotNil(t, err)
	// wrong view height
	preparePayload, err = NewPrepareVotePayload(
		PrepareAction, block.Hash(), block.Height(),
		block.Height()).ToBytes()
	assert.Nil(t, err)
	prepareTx = NewTransaction(
		block.header.chainID, &Address{validators[2]}, &Address{validators[2]},
		zero, 1, TxPayloadVoteType, preparePayload)
	giveback, err = newBlock.executeTransaction(prepareTx)
	assert.Equal(t, giveback, false)
	assert.NotNil(t, err)
	block.commit()
}

func TestVotePayload_InvalidViewBlock(t *testing.T) {
	validators, block := NewBlockWithValidDynasty(t)
	size := len(validators)
	newBlock := NewBlock(block.header.chainID, &Address{validators[0]}, block)
	newBlock.begin()
	zero := util.NewUint128()
	for i := 0; i < size*2/3; i++ {
		preparePayload, err := NewPrepareVotePayload(
			PrepareAction, block.ParentHash(), block.parentBlock.Height(), 1).ToBytes()
		assert.Nil(t, err)
		prepareTx := NewTransaction(
			block.header.chainID, &Address{validators[i]}, &Address{validators[i]},
			zero, 1, TxPayloadVoteType, preparePayload)
		giveback, err := newBlock.executeTransaction(prepareTx)
		assert.Equal(t, giveback, false)
		assert.Nil(t, err)
	}
	newBlock.commit()
	newBlock.Seal()
	assert.Nil(t, storeBlockToStorage(newBlock))
	maxVotes, err := countValidators(newBlock.curDynastyTrie, nil)
	assert.Nil(t, err)
	assert.Equal(t, maxVotes, uint32(size))
	prepareVotes, err := countValidators(newBlock.prepareVotesTrie, block.ParentHash())
	assert.Nil(t, err)
	assert.Equal(t, prepareVotes, uint32(size*2/3))

	newBlock2 := NewBlock(newBlock.header.chainID, newBlock.Coinbase(), newBlock)
	newBlock2.begin()
	preparePayload, err := NewPrepareVotePayload(
		PrepareAction, block.Hash(), block.Height(),
		block.parentBlock.Height()).ToBytes()
	assert.Nil(t, err)
	prepareTx := NewTransaction(
		block.header.chainID, &Address{validators[1]}, &Address{validators[1]},
		zero, 2, TxPayloadVoteType, preparePayload)
	giveback, err := newBlock2.executeTransaction(prepareTx)
	assert.Equal(t, giveback, false)
	assert.Nil(t, err)
	newBlock2.commit()
	newBlock2.Seal()
	assert.Nil(t, storeBlockToStorage(newBlock2))
	_, err = newBlock2.depositTrie.Get(validators[1])
	assert.NotNil(t, err)
}

func TestVotePayload_PrepareChangedBlock(t *testing.T) {
	validators, block := NewBlockWithValidDynasty(t)
	zero := util.NewUint128()
	newBlock := NewBlock(block.header.chainID, &Address{validators[0]}, block)
	newBlock.begin()
	changePayload, err := NewChangeVotePayload(ChangeAction, block.Hash(), 1).ToBytes()
	assert.Nil(t, err)
	changeTx := NewTransaction(
		block.header.chainID, &Address{validators[1]}, &Address{validators[1]},
		zero, 1, TxPayloadVoteType, changePayload)
	giveback, err := newBlock.executeTransaction(changeTx)
	assert.Equal(t, giveback, false)
	assert.Nil(t, err)
	newBlock.commit()
	newBlock.Seal()
	assert.Nil(t, storeBlockToStorage(newBlock))

	newBlock1 := NewBlock(newBlock.header.chainID, &Address{validators[0]}, newBlock)
	newBlock1.begin()
	preparePayload, err := NewPrepareVotePayload(PrepareAction, newBlock.Hash(), newBlock.Height(), 1).ToBytes()
	assert.Nil(t, err)
	prepareTx := NewTransaction(block.header.chainID, &Address{validators[1]}, &Address{validators[1]},
		zero, 2, TxPayloadVoteType, preparePayload)
	giveback, err = newBlock1.executeTransaction(prepareTx)
	assert.Equal(t, giveback, false)
	assert.Nil(t, err)
	newBlock1.commit()
	newBlock1.Seal()
	assert.Nil(t, storeBlockToStorage(newBlock1))

	_, err = newBlock1.depositTrie.Get(validators[1])
	assert.NotNil(t, err)
}

func TestVotePayload_PrepareAfterAbdicate(t *testing.T) {
	validators, block := NewBlockWithValidDynasty(t)
	zero := util.NewUint128()

	newBlock := NewBlock(block.header.chainID, &Address{validators[0]}, block)
	newBlock.begin()
	abdicatePayload, err := NewAbdicateVotePayload(AbdicateAction, block.Hash(), block.CurDynastyRoot()).ToBytes()
	assert.Nil(t, err)
	abdicateTx := NewTransaction(
		block.header.chainID, &Address{validators[1]}, &Address{validators[1]},
		zero, 1, TxPayloadVoteType, abdicatePayload)
	giveback, err := newBlock.executeTransaction(abdicateTx)
	assert.Equal(t, giveback, false)
	assert.Nil(t, err)
	newBlock.commit()
	newBlock.Seal()
	assert.Nil(t, storeBlockToStorage(newBlock))

	newBlock1 := NewBlock(newBlock.header.chainID, &Address{validators[0]}, newBlock)
	newBlock1.begin()
	preparePayload, err := NewPrepareVotePayload(PrepareAction, newBlock.Hash(), newBlock.Height(), 1).ToBytes()
	assert.Nil(t, err)
	prepareTx := NewTransaction(block.header.chainID, &Address{validators[1]}, &Address{validators[1]},
		zero, 2, TxPayloadVoteType, preparePayload)
	giveback, err = newBlock1.executeTransaction(prepareTx)
	assert.Equal(t, giveback, false)
	assert.Nil(t, err)
	newBlock1.commit()
	newBlock1.Seal()
	assert.Nil(t, storeBlockToStorage(newBlock1))

	_, err = newBlock1.depositTrie.Get(validators[1])
	assert.NotNil(t, err)
}

func TestVotePayload_PrepareOnJumpedHeight(t *testing.T) {
	validators, block := NewBlockWithValidDynasty(t)
	zero := util.NewUint128()

	newBlock := NewBlock(block.header.chainID, &Address{validators[0]}, block)
	newBlock.begin()
	preparePayload, err := NewPrepareVotePayload(PrepareAction, block.Hash(), block.Height(), 1).ToBytes()
	assert.Nil(t, err)
	prepareTx := NewTransaction(block.header.chainID, &Address{validators[1]}, &Address{validators[1]},
		zero, 1, TxPayloadVoteType, preparePayload)
	giveback, err := newBlock.executeTransaction(prepareTx)
	assert.Equal(t, giveback, false)
	assert.Nil(t, err)
	newBlock.commit()
	newBlock.Seal()
	assert.Nil(t, storeBlockToStorage(newBlock))

	newBlock1 := NewBlock(newBlock.header.chainID, &Address{validators[0]}, newBlock)
	newBlock1.begin()
	preparePayload, err = NewPrepareVotePayload(PrepareAction, block.ParentHash(), block.parentBlock.Height(), 1).ToBytes()
	assert.Nil(t, err)
	prepareTx = NewTransaction(block.header.chainID, &Address{validators[1]}, &Address{validators[1]},
		zero, 2, TxPayloadVoteType, preparePayload)
	giveback, err = newBlock1.executeTransaction(prepareTx)
	assert.Equal(t, giveback, false)
	assert.Nil(t, err)
	newBlock1.commit()
	newBlock1.Seal()
	assert.Nil(t, storeBlockToStorage(newBlock1))

	_, err = newBlock1.depositTrie.Get(validators[1])
	assert.NotNil(t, err)
}

func TestVotePayload_CommitBeforePrepare(t *testing.T) {
	validators, block := NewBlockWithValidDynasty(t)
	zero := util.NewUint128()

	newBlock := NewBlock(block.header.chainID, &Address{validators[0]}, block)
	newBlock.begin()
	commitPayload, err := NewCommitVotePayload(CommitAction, block.Hash()).ToBytes()
	assert.Nil(t, err)
	commitTx := NewTransaction(block.header.chainID, &Address{validators[1]}, &Address{validators[1]},
		zero, 1, TxPayloadVoteType, commitPayload)
	giveback, err := newBlock.executeTransaction(commitTx)
	assert.Equal(t, giveback, false)
	assert.Nil(t, err)
	newBlock.commit()
	newBlock.Seal()
	assert.Nil(t, storeBlockToStorage(newBlock))

	_, err = newBlock.depositTrie.Get(validators[1])
	assert.NotNil(t, err)
}

func TestVotePayload_ChangeBeforeLastChanged(t *testing.T) {
	validators, block := NewBlockWithValidDynasty(t)
	size := len(validators)
	newBlock := NewBlock(block.header.chainID, &Address{validators[0]}, block)
	newBlock.begin()
	zero := util.NewUint128()
	for i := 0; i < size*2/3; i++ {
		changePayload, err := NewChangeVotePayload(
			ChangeAction, block.Hash(), 1).ToBytes()
		assert.Nil(t, err)
		changeTx := NewTransaction(
			block.header.chainID, &Address{validators[i]}, &Address{validators[i]},
			zero, 1, TxPayloadVoteType, changePayload)
		giveback, err := newBlock.executeTransaction(changeTx)
		assert.Equal(t, giveback, false)
		assert.Nil(t, err)
	}
	newBlock.commit()
	newBlock.Seal()
	assert.Nil(t, storeBlockToStorage(newBlock))

	newBlock2 := NewBlock(newBlock.header.chainID, newBlock.Coinbase(), newBlock)
	newBlock2.begin()
	changePayload, err := NewChangeVotePayload(
		ChangeAction, block.Hash(), 2).ToBytes()
	assert.Nil(t, err)
	changeTx := NewTransaction(
		newBlock.header.chainID, &Address{validators[1]}, &Address{validators[1]},
		zero, 2, TxPayloadVoteType, changePayload)
	giveback, err := newBlock2.executeTransaction(changeTx)
	assert.Equal(t, giveback, false)
	assert.Nil(t, err)
	newBlock2.commit()
	newBlock2.Seal()
	assert.Nil(t, storeBlockToStorage(newBlock2))

	_, err = newBlock2.depositTrie.Get(validators[1])
	assert.NotNil(t, err)
}
