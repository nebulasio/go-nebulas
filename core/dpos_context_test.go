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
	log "github.com/sirupsen/logrus"
)

func TestBlock_NextDynastyContext(t *testing.T) {
	storage, _ := storage.NewMemoryStorage()
	chain, _ := NewBlockChain(0, storage)
	block, _ := LoadBlockFromStorage(GenesisHash, chain.storage, chain.txPool)

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

	context, err = block.NextDynastyContext(DynastyInterval / 2)
	assert.Nil(t, err)
	validators, _ = TraverseDynasty(block.dposContext.dynastyTrie)
	assert.Equal(t, context.Proposer, validators[int(DynastyInterval/2/BlockInterval)%DynastySize])
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

	context, err = block.NextDynastyContext(DynastyInterval*2 + DynastyInterval/3)
	assert.Nil(t, err)
	validators, _ = TraverseDynasty(block.dposContext.dynastyTrie)
	index := int((DynastyInterval*2+DynastyInterval/3)%DynastyInterval) / int(BlockInterval) % DynastySize
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
	context, err := block.NextDynastyContext(DynastyInterval)
	assert.Nil(t, err)
	_, err = context.NextDynastyTrie.Get(validators[ReserveSize+1])
	assert.Nil(t, err)
}

func TestBlock_Kickout(t *testing.T) {
	storage, _ := storage.NewMemoryStorage()
	chain, _ := NewBlockChain(0, storage)
	validators, _ := TraverseDynasty(chain.tailBlock.dposContext.dynastyTrie)
	coinbase := &Address{validators[2]}

	block, _ := NewBlock(0, coinbase, chain.tailBlock)
	block.header.timestamp = DynastyInterval
	context, err := chain.tailBlock.NextDynastyContext(block.Timestamp() - chain.tailBlock.Timestamp())
	assert.Nil(t, err)
	block.LoadDynastyContext(context)
	block.SetMiner(coinbase)
	assert.Equal(t, block.Seal(), nil)
	block, _ = mockBlockFromNetwork(block)
	assert.Equal(t, block.LinkParentBlock(chain.tailBlock), true)
	block.SetMiner(coinbase)
	assert.Nil(t, block.Verify(0))
	chain.SetTailBlock(block)
	delegatees, _ := TraverseDynasty(chain.tailBlock.dposContext.dynastyTrie)
	for i := 0; i < DynastySize-1; i++ {
		assert.Equal(t, string(delegatees[i].Hex()), GenesisDynasty[i])
	}
	delegatees, _ = TraverseDynasty(chain.tailBlock.dposContext.nextDynastyTrie)
	for i := 0; i < DynastySize-1; i++ {
		assert.Equal(t, string(delegatees[i].Hex()), GenesisDynasty[i])
	}

	log.Info(chain.tailBlock.DposContextHash())
	log.Info(chain.tailBlock.dposContext.RootHash())
	block, _ = NewBlock(0, coinbase, block)
	block.header.timestamp = DynastyInterval * 2
	context, err = chain.tailBlock.NextDynastyContext(block.Timestamp() - chain.tailBlock.Timestamp())
	assert.Nil(t, err)
	block.LoadDynastyContext(context)
	block.SetMiner(coinbase)
	assert.Equal(t, block.Seal(), nil)
	log.Info(chain.tailBlock.DposContextHash())
	log.Info(chain.tailBlock.dposContext.RootHash())
	block, _ = mockBlockFromNetwork(block)
	assert.Equal(t, block.LinkParentBlock(chain.tailBlock), true)
	block.SetMiner(coinbase)
	assert.Nil(t, block.Verify(0))
	chain.SetTailBlock(block)
	delegatees, _ = TraverseDynasty(chain.tailBlock.dposContext.dynastyTrie)
	for i := 0; i < DynastySize-1; i++ {
		assert.Equal(t, string(delegatees[i].Hex()), GenesisDynasty[i])
	}
	delegatees, _ = TraverseDynasty(chain.tailBlock.dposContext.nextDynastyTrie)
	for i := 0; i <= ReserveSize; i++ {
		assert.Equal(t, string(delegatees[i].Hex()), GenesisDynasty[i])
	}
	cnt := 0
	for i := ReserveSize + 1; i < DynastySize; i++ {
		for j := DynastySize; j < len(GenesisDynasty); j++ {
			if delegatees[i].String() == GenesisDynasty[j] {
				cnt++
			}
		}
	}
	assert.Equal(t, cnt, DynastySize-ReserveSize-1)
}
