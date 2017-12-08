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
	context, err := block.NextDynastyContext(5)
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

	context, err = block.NextDynastyContext(3605)
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

	context, err = block.NextDynastyContext(110)
	assert.Nil(t, err)
	validators, _ = TraverseDynasty(block.dposContext.dynastyTrie)
	assert.Equal(t, context.Proposer, validators[110/int(BlockInterval)%len(GenesisDynasty)])
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

	context, err = block.NextDynastyContext(7310)
	assert.Nil(t, err)
	validators, _ = TraverseDynasty(block.dposContext.dynastyTrie)
	assert.Equal(t, context.Proposer, validators[7310%int(DynastyInterval)/int(BlockInterval)%len(GenesisDynasty)])
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

	// new block
	newBlock, _ := NewBlock(chain.ChainID(), &Address{validators[1]}, chain.tailBlock)
	newBlock.LoadDynastyContext(context)
	newBlock.CollectTransactions(500)
	newBlock.Seal()
	newBlock, _ = mockBlockFromNetwork(newBlock)
	newBlock.LinkParentBlock(chain.tailBlock)
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
	dynasty, err := block.electNewDynasty(0)
	assert.Nil(t, err)
	_, err = dynasty.Get(validators[ReserveSize+1])
	assert.Nil(t, err)
}
