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

func NewBlockWithValidDynasty(t *testing.T, size int) ([]byteutils.Hash, *Block) {
	storage, _ := storage.NewMemoryStorage()
	genesis := NewGenesisBlock(0, storage, nil)
	genesis.begin()
	loginPayload, _ := NewElectPayload(LoginAction).ToBytes()
	validators := []byteutils.Hash{}
	for i := 0; i < size; i++ {
		v := GenerateNewAddress()
		validators = append(validators, v.Bytes())
		account := genesis.accState.GetOrCreateUserAccount(v.Bytes())
		account.AddBalance(StandardDeposit)
		tx := NewTransaction(genesis.header.chainID, v, v, zero, 1, TxPayloadElectType, loginPayload)
		giveback, err := genesis.executeTransaction(tx)
		assert.Equal(t, giveback, false)
		assert.Nil(t, err)
	}
	genesis.commit()
	genesis.Seal()

	coinbase := validators[0]
	block1 := NewBlock(genesis.header.chainID, &Address{coinbase}, genesis)
	block1.Seal()
	block2 := NewBlock(block1.header.chainID, &Address{coinbase}, block1)
	block2.Seal()
	block3 := NewBlock(block2.header.chainID, &Address{coinbase}, block2)
	block3.Seal()

	assert.Nil(t, storeBlockToStorage(genesis))
	assert.Nil(t, storeBlockToStorage(block1))
	assert.Nil(t, storeBlockToStorage(block2))
	assert.Nil(t, storeBlockToStorage(block3))
	return validators, block3
}

func TestVotePayload_RightFlow(t *testing.T) {
	size := 3
	validators, block := NewBlockWithValidDynasty(t, size)
	block.begin()
	zero := util.NewUint128()
	for i := 0; i < size; i++ {
		preparePayload, err := NewPrepareVotePayload(
			PrepareAction, block.ParentHash(), block.parentBlock.Height(),
			block.parentBlock.parentBlock.parentBlock.Height()).ToBytes()
		assert.Nil(t, err)
		prepareTx := NewTransaction(
			block.header.chainID, &Address{validators[i]}, &Address{validators[i]},
			zero, 2, TxPayloadVoteType, preparePayload)
		giveback, err := block.executeTransaction(prepareTx)
		assert.Equal(t, giveback, false)
		assert.Nil(t, err)
	}
	for i := 0; i < size; i++ {
		commitPayload, err := NewCommitVotePayload(CommitAction, block.ParentHash()).ToBytes()
		assert.Nil(t, err)
		commitTx := NewTransaction(
			block.header.chainID, &Address{validators[i]}, &Address{validators[i]},
			zero, 3, TxPayloadVoteType, commitPayload)
		giveback, err := block.executeTransaction(commitTx)
		assert.Equal(t, giveback, false)
		assert.Nil(t, err)
	}
	reward := util.NewUint128()
	reward.Add(reward.Int, FinalityBlockReward.Int)
	reward.Add(reward.Int, StandardDeposit.Int)
	reward.Sub(reward.Int, VoteBlockReward.Int)
	for i := 0; i < size; i++ {
		depositBytes, err := block.depositTrie.Get(validators[i])
		assert.Nil(t, err)
		deposit, err := util.NewUint128FromFixedSizeByteSlice(depositBytes)
		assert.Nil(t, err)
		assert.Equal(t, deposit, reward)
	}
	block.commit()
}

func TestVotePayload_PrepareWrongHeight(t *testing.T) {
	size := 3
	validators, block := NewBlockWithValidDynasty(t, size)
	block.begin()
	zero := util.NewUint128()
	// wrong current height
	preparePayload, err := NewPrepareVotePayload(
		PrepareAction, block.ParentHash(), block.parentBlock.Height()+1,
		block.parentBlock.parentBlock.Height()).ToBytes()
	assert.Nil(t, err)
	prepareTx := NewTransaction(
		block.header.chainID, &Address{validators[1]}, &Address{validators[1]},
		zero, 2, TxPayloadVoteType, preparePayload)
	giveback, err := block.executeTransaction(prepareTx)
	assert.Equal(t, giveback, false)
	assert.Nil(t, err)
	_, err = block.depositTrie.Get(validators[1])
	assert.NotNil(t, err)
	// wrong view height
	preparePayload, err = NewPrepareVotePayload(
		PrepareAction, block.ParentHash(), block.parentBlock.Height(),
		block.parentBlock.Height()).ToBytes()
	assert.Nil(t, err)
	prepareTx = NewTransaction(
		block.header.chainID, &Address{validators[1]}, &Address{validators[1]},
		zero, 2, TxPayloadVoteType, preparePayload)
	giveback, err = block.executeTransaction(prepareTx)
	assert.Equal(t, giveback, false)
	assert.NotNil(t, err)
	block.commit()
}

func TestVotePayload_InvalidViewBlock(t *testing.T) {
	size := 3
	validators, block := NewBlockWithValidDynasty(t, size)
	newBlock := NewBlock(block.header.chainID, block.Coinbase(), block)
	newBlock.begin()
	zero := util.NewUint128()
	for i := 0; i < size*2/3; i++ {
		preparePayload, err := NewPrepareVotePayload(
			PrepareAction, block.ParentHash(), block.parentBlock.Height(),
			block.parentBlock.parentBlock.parentBlock.Height()).ToBytes()
		assert.Nil(t, err)
		prepareTx := NewTransaction(
			block.header.chainID, &Address{validators[i]}, &Address{validators[i]},
			zero, 2, TxPayloadVoteType, preparePayload)
		giveback, err := newBlock.executeTransaction(prepareTx)
		assert.Equal(t, giveback, false)
		assert.Nil(t, err)
	}
	newBlock.commit()
	newBlock.Seal()
	assert.Nil(t, storeBlockToStorage(newBlock))
	maxVotes, err := countValidators(newBlock.curDynastyTrie, nil)
	assert.Nil(t, err)
	assert.Equal(t, maxVotes, size)
	prepareVotes, err := countValidators(newBlock.prepareVotesTrie, block.ParentHash())
	assert.Nil(t, err)
	assert.Equal(t, prepareVotes, size*2/3)

	newBlock2 := NewBlock(newBlock.header.chainID, newBlock.Coinbase(), newBlock)
	newBlock2.begin()
	preparePayload, err := NewPrepareVotePayload(
		PrepareAction, block.Hash(), block.Height(),
		block.parentBlock.Height()).ToBytes()
	assert.Nil(t, err)
	prepareTx := NewTransaction(
		block.header.chainID, &Address{validators[1]}, &Address{validators[1]},
		zero, 3, TxPayloadVoteType, preparePayload)
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
	size := 3
	validators, block := NewBlockWithValidDynasty(t, size)
	validators, err := block.NextBlockSortedValidators()
	assert.Nil(t, err)
	zero := util.NewUint128()

	newBlock := NewBlock(block.header.chainID, &Address{validators[0]}, block)
	newBlock.begin()
	changePayload, err := NewChangeVotePayload(ChangeAction, block.Hash(), 1).ToBytes()
	assert.Nil(t, err)
	changeTx := NewTransaction(
		block.header.chainID, &Address{validators[1]}, &Address{validators[1]},
		zero, 2, TxPayloadVoteType, changePayload)
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
		zero, 3, TxPayloadVoteType, preparePayload)
	giveback, err = newBlock1.executeTransaction(prepareTx)
	assert.Equal(t, giveback, false)
	assert.Nil(t, err)
	newBlock1.commit()
	newBlock1.Seal()
	assert.Nil(t, storeBlockToStorage(newBlock1))

	_, err = newBlock1.depositTrie.Get(validators[1])
	assert.NotNil(t, err)
}
