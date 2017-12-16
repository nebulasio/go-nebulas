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

	"github.com/nebulasio/go-nebulas/common/trie"
	"github.com/nebulasio/go-nebulas/util"
	log "github.com/sirupsen/logrus"
)

func checkDynasty(t *testing.T, dynasty *trie.BatchTrie) {
	delegatees, err := TraverseDynasty(dynasty)
	assert.Nil(t, err)
	for i := 0; i < DynastySize-1; i++ {
		assert.Equal(t, string(delegatees[i].Hex()), MockDynasty[i])
	}
}

func TestBlock_NextDynastyContext(t *testing.T) {
	neb := testNeb()
	chain, _ := NewBlockChain(neb)
	block, _ := LoadBlockFromStorage(GenesisHash, chain.storage, chain.txPool, neb.emitter)

	context, err := block.NextDynastyContext(BlockInterval)
	assert.Nil(t, err)
	validators, _ := TraverseDynasty(block.dposContext.dynastyTrie)
	assert.Equal(t, context.Proposer, validators[1])
	// check dynasty
	checkDynasty(t, context.DynastyTrie)
	checkDynasty(t, context.NextDynastyTrie)

	context, err = block.NextDynastyContext(BlockInterval + DynastyInterval)
	assert.Nil(t, err)
	validators, _ = TraverseDynasty(block.dposContext.dynastyTrie)
	assert.Equal(t, context.Proposer, validators[1])
	// check dynasty
	checkDynasty(t, context.DynastyTrie)
	checkDynasty(t, context.NextDynastyTrie)

	context, err = block.NextDynastyContext(DynastyInterval / 2)
	assert.Nil(t, err)
	validators, _ = TraverseDynasty(block.dposContext.dynastyTrie)
	assert.Equal(t, context.Proposer, validators[int(DynastyInterval/2/BlockInterval)%DynastySize])
	// check dynasty
	checkDynasty(t, context.DynastyTrie)
	checkDynasty(t, context.NextDynastyTrie)

	context, err = block.NextDynastyContext(DynastyInterval*2 + DynastyInterval/3)
	assert.Nil(t, err)
	validators, _ = TraverseDynasty(block.dposContext.dynastyTrie)
	index := int((DynastyInterval*2+DynastyInterval/3)%DynastyInterval) / int(BlockInterval) % DynastySize
	assert.Equal(t, context.Proposer, validators[index])
	// check dynasty
	checkDynasty(t, context.DynastyTrie)
	checkDynasty(t, context.NextDynastyTrie)

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
	neb := testNeb()
	chain, _ := NewBlockChain(neb)
	block, _ := LoadBlockFromStorage(GenesisHash, chain.storage, chain.txPool, neb.emitter)
	block.begin()
	kickout, _ := AddressParse(MockDynasty[0])
	v, _ := AddressParse(MockDynasty[DynastySize-1])
	block.accState.GetOrCreateUserAccount(v.Bytes()).AddBalance(util.NewUint128FromInt(2000000))
	block.accState.GetOrCreateUserAccount(kickout.Bytes()).AddBalance(util.NewUint128FromInt(2000000))
	delegatePayload := NewDelegatePayload(DelegateAction, v.String())
	bytes, _ := delegatePayload.ToBytes()
	tx := NewTransaction(0, v, v, util.NewUint128FromInt(1), 1, TxPayloadDelegateType, bytes, TransactionGasPrice, util.NewUint128FromInt(200000))
	_, err := block.executeTransaction(tx)
	candidatePayload := NewCandidatePayload(LogoutAction)
	bytes, _ = candidatePayload.ToBytes()
	tx = NewTransaction(0, kickout, kickout, util.NewUint128FromInt(1), 1, TxPayloadCandidateType, bytes, TransactionGasPrice, util.NewUint128FromInt(200000))
	_, err = block.executeTransaction(tx)
	assert.Nil(t, err)
	block.commit()
	context, err := block.NextDynastyContext(DynastyInterval)
	assert.Nil(t, err)
	log.Info(v.String())
	_, err = context.NextDynastyTrie.Get(v.Bytes())
	assert.Nil(t, err)
}

func TestBlock_Kickout(t *testing.T) {
	neb := testNeb()
	chain, _ := NewBlockChain(neb)
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
	checkDynasty(t, chain.tailBlock.dposContext.dynastyTrie)
	checkDynasty(t, chain.tailBlock.dposContext.nextDynastyTrie)

	block, _ = NewBlock(0, coinbase, block)
	block.header.timestamp = DynastyInterval * 2
	context, err = chain.tailBlock.NextDynastyContext(block.Timestamp() - chain.tailBlock.Timestamp())
	assert.Nil(t, err)
	block.LoadDynastyContext(context)
	block.SetMiner(coinbase)
	assert.Equal(t, block.Seal(), nil)
	block, _ = mockBlockFromNetwork(block)
	assert.Equal(t, block.LinkParentBlock(chain.tailBlock), true)
	block.SetMiner(coinbase)
	assert.Nil(t, block.Verify(0))
	chain.SetTailBlock(block)
	checkDynasty(t, chain.tailBlock.dposContext.dynastyTrie)
	checkDynasty(t, chain.tailBlock.dposContext.nextDynastyTrie)
}
