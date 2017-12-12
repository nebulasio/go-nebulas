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

	"github.com/stretchr/testify/assert"

	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util"
)

func TestBlock_NextDynastyContext(t *testing.T) {
	storage, _ := storage.NewMemoryStorage()
	chain, _ := NewBlockChain(0, storage)
	block, _ := LoadBlockFromStorage(GenesisHash, chain.storage, chain.txPool)

	block.begin()
	context, err := block.NextDynastyContext(BlockInterval)
	assert.Nil(t, err)
	validators, _ := TraverseDynasty(block.dposContext.dynastyTrie)
	assert.Equal(t, context.Proposer, validators[1])
	// check dynasty
	delegatees, err := TraverseDynasty(context.DynastyTrie)
	assert.Nil(t, err)
	for i := 0; i < DynastySize-1; i++ {
		assert.Equal(t, string(delegatees[i].Hex()), GenesisDynasty[i])
	}
	delegatees, err = TraverseDynasty(context.NextDynastyTrie)
	assert.Nil(t, err)
	for i := 0; i < DynastySize-1; i++ {
		assert.Equal(t, string(delegatees[i].Hex()), GenesisDynasty[i])
	}
	block.rollback()

	block.begin()
	context, err = block.NextDynastyContext(BlockInterval + DynastyInterval)
	assert.Nil(t, err)
	validators, _ = TraverseDynasty(block.dposContext.dynastyTrie)
	assert.Equal(t, context.Proposer, validators[1])
	// check dynasty
	delegatees, err = TraverseDynasty(context.DynastyTrie)
	assert.Nil(t, err)
	for i := 0; i < DynastySize-1; i++ {
		assert.Equal(t, string(delegatees[i].Hex()), GenesisDynasty[i])
	}
	delegatees, err = TraverseDynasty(context.NextDynastyTrie)
	assert.Nil(t, err)
	for i := 0; i < DynastySize-1; i++ {
		assert.Equal(t, string(delegatees[i].Hex()), GenesisDynasty[i])
	}
	block.rollback()

	block.begin()
	context, err = block.NextDynastyContext(DynastyInterval / 2)
	assert.Nil(t, err)
	validators, _ = TraverseDynasty(block.dposContext.dynastyTrie)
	assert.Equal(t, context.Proposer, validators[int(DynastyInterval/2/BlockInterval)%len(GenesisDynasty)])
	// check dynasty
	delegatees, err = TraverseDynasty(context.DynastyTrie)
	assert.Nil(t, err)
	for i := 0; i < DynastySize-1; i++ {
		assert.Equal(t, string(delegatees[i].Hex()), GenesisDynasty[i])
	}
	delegatees, err = TraverseDynasty(context.NextDynastyTrie)
	assert.Nil(t, err)
	for i := 0; i < DynastySize-1; i++ {
		assert.Equal(t, string(delegatees[i].Hex()), GenesisDynasty[i])
	}
	block.rollback()

	block.begin()
	context, err = block.NextDynastyContext(DynastyInterval*2 + DynastyInterval/3)
	assert.Nil(t, err)
	validators, _ = TraverseDynasty(block.dposContext.dynastyTrie)
	index := int((DynastyInterval*2+DynastyInterval/3)%DynastyInterval) / int(BlockInterval) % len(GenesisDynasty)
	assert.Equal(t, context.Proposer, validators[index])
	// check dynasty
	delegatees, err = TraverseDynasty(context.DynastyTrie)
	assert.Nil(t, err)
	for i := 0; i < DynastySize-1; i++ {
		assert.Equal(t, string(delegatees[i].Hex()), GenesisDynasty[i])
	}
	delegatees, err = TraverseDynasty(context.NextDynastyTrie)
	assert.Nil(t, err)
	for i := 0; i < DynastySize-1; i++ {
		assert.Equal(t, string(delegatees[i].Hex()), GenesisDynasty[i])
	}
	block.rollback()

	// new block
	coinbase := &Address{validators[1]}
	newBlock, _ := NewBlock(chain.ChainID(), coinbase, chain.tailBlock)
	newBlock.LoadDynastyContext(context)
	newBlock.CollectTransactions(500)
	newBlock.SetMiner(coinbase)
	newBlock.Seal()
	newBlock, _ = mockBlockFromNetwork(newBlock)
	newBlock.LinkParentBlock(chain.tailBlock)
	newBlock.SetMiner(coinbase)
	assert.Nil(t, newBlock.Verify(chain.ChainID()))
}

func TestBlock_ElectNewDynasty(t *testing.T) {
	storage, _ := storage.NewMemoryStorage()
	chain, _ := NewBlockChain(0, storage)
	block, _ := LoadBlockFromStorage(GenesisHash, chain.storage, chain.txPool)
	validators, _ := TraverseDynasty(block.dposContext.dynastyTrie)
	block.begin()
	v := &Address{validators[DynastySize-1]}
	block.accState.GetOrCreateUserAccount(v.Bytes()).AddBalance(util.NewUint128FromInt(2000000))
	payload, _ := NewDelegatePayload(DelegateAction, v.ToHex())
	bytes, _ := payload.ToBytes()
	tx := NewTransaction(0, v, v, util.NewUint128FromInt(1), 1, TxPayloadDelegateType, bytes, TransactionGasPrice, util.NewUint128FromInt(200000))
	_, err := block.executeTransaction(tx)
	assert.Nil(t, err)
	block.commit()
	dynasty, err := block.electNewDynasty(0, 1)
	assert.Nil(t, err)
	_, err = dynasty.Get(validators[ReserveSize+1])
	assert.Nil(t, err)
}
