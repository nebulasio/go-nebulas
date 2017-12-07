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
)

func TestBlock_NextDynastyContext(t *testing.T) {
	storage, _ := storage.NewMemoryStorage()
	chain, _ := NewBlockChain(0, storage)
	block, _ := LoadBlockFromStorage(GenesisHash, chain.storage, chain.txPool)
	context, err := block.NextDynastyContext(5)
	assert.Nil(t, err)
	validators, _ := TraverseDynasty(block.dposContext.dynastyTrie)
	assert.Equal(t, context.Proposer, validators[1])
	context, err = block.NextDynastyContext(3605)
	assert.Nil(t, err)
	validators, _ = TraverseDynasty(block.dposContext.dynastyTrie)
	assert.Equal(t, context.Proposer, validators[1])
	context, err = block.NextDynastyContext(110)
	assert.Nil(t, err)
	validators, _ = TraverseDynasty(block.dposContext.dynastyTrie)
	assert.Equal(t, context.Proposer, validators[110/int(BlockInterval)%len(GenesisDynasty)])
	context, err = block.NextDynastyContext(7310)
	assert.Nil(t, err)
	validators, _ = TraverseDynasty(block.dposContext.dynastyTrie)
	assert.Equal(t, context.Proposer, validators[7310%int(DynastyInterval)/int(BlockInterval)%len(GenesisDynasty)])

	// new block
	newBlock, _ := NewBlock(chain.ChainID(), &Address{validators[1]}, chain.tailBlock)
	newBlock.LoadDynastyContext(context)
	newBlock.CollectTransactions(500)
	newBlock.Seal()
	newBlock, _ = mockBlockFromNetwork(newBlock)
	newBlock.LinkParentBlock(chain.tailBlock)
	assert.Nil(t, newBlock.Verify(chain.ChainID()))

	// check dynasty
	delegatees, err := TraverseDynasty(newBlock.dposContext.dynastyTrie)
	assert.Nil(t, err)
	for i := 0; i < DynastySize-1; i++ {
		assert.Equal(t, string(delegatees[i].Hex()), GenesisDynasty[i])
	}
}
