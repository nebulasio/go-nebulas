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
	"testing"

	"github.com/nebulasio/go-nebulas/consensus/pb"

	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/crypto/keystore"

	"github.com/stretchr/testify/assert"

	"github.com/nebulasio/go-nebulas/common/trie"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util/byteutils"
)

func checkDynasty(t *testing.T, consensus core.Consensus, consensusRoot *consensuspb.ConsensusRoot, storage storage.Storage) {
	consensusState, err := consensus.NewState(consensusRoot, storage)
	assert.Nil(t, err)
	dynasty, err := consensusState.Dynasty()
	assert.Nil(t, err)
	for i := 0; i < DynastySize-1; i++ {
		assert.Equal(t, string(dynasty[i].Hex()), DefaultOpenDynasty[i])
	}
}

func TestBlock_NextDynastyContext(t *testing.T) {
	neb := mockNeb(t)
	block := neb.chain.GenesisBlock()

	context, err := block.WorldState().NextConsensusState(BlockIntervalInMs / SecondInMs)
	assert.Nil(t, err)
	validators, _ := block.WorldState().Dynasty()
	assert.Equal(t, context.Proposer(), validators[1])
	// check dynasty
	consensusRoot, err := context.RootHash()
	assert.Nil(t, err)
	checkDynasty(t, neb.consensus, consensusRoot, neb.Storage())

	context, err = block.WorldState().NextConsensusState((BlockIntervalInMs + DynastyIntervalInMs) / SecondInMs)
	assert.Nil(t, err)
	validators, _ = block.WorldState().Dynasty()
	assert.Equal(t, context.Proposer(), validators[1])
	// check dynasty
	consensusRoot, err = context.RootHash()
	assert.Nil(t, err)
	checkDynasty(t, neb.consensus, consensusRoot, neb.Storage())

	context, err = block.WorldState().NextConsensusState(DynastyIntervalInMs / SecondInMs / 2)
	assert.Nil(t, err)
	validators, _ = block.WorldState().Dynasty()
	assert.Equal(t, context.Proposer(), validators[int(DynastyIntervalInMs/2/BlockIntervalInMs)%DynastySize])
	// check dynasty
	consensusRoot, err = context.RootHash()
	assert.Nil(t, err)
	checkDynasty(t, neb.consensus, consensusRoot, neb.Storage())

	context, err = block.WorldState().NextConsensusState((DynastyIntervalInMs*2 + DynastyIntervalInMs/3) / SecondInMs)
	assert.Nil(t, err)
	validators, _ = block.WorldState().Dynasty()
	index := int((DynastyIntervalInMs*2+DynastyIntervalInMs/3)%DynastyIntervalInMs) / int(BlockIntervalInMs) % DynastySize
	assert.Equal(t, context.Proposer(), validators[index])
	// check dynasty
	consensusRoot, err = context.RootHash()
	assert.Nil(t, err)
	checkDynasty(t, neb.consensus, consensusRoot, neb.Storage())

	// new block
	coinbase, err := core.AddressParseFromBytes(validators[4])
	assert.Nil(t, err)
	assert.Nil(t, neb.am.Unlock(coinbase, []byte("passphrase"), keystore.DefaultUnlockDuration))

	newBlock, _ := core.NewBlock(neb.chain.ChainID(), coinbase, neb.chain.TailBlock())
	newBlock.SetTimestamp((DynastyIntervalInMs*2 + DynastyIntervalInMs/3) / SecondInMs)
	newBlock.WorldState().SetConsensusState(context)
	assert.Equal(t, newBlock.Seal(), nil)
	assert.Nil(t, neb.am.SignBlock(coinbase, newBlock))
	newBlock, _ = mockBlockFromNetwork(newBlock)
	newBlock.LinkParentBlock(neb.chain, neb.chain.TailBlock())
	assert.Nil(t, newBlock.VerifyExecution()) //neb.chain.TailBlock(), neb.chain.ConsensusHandler()
}

func TestTraverseDynasty(t *testing.T) {
	stor, err := storage.NewMemoryStorage()
	assert.Nil(t, err)
	dynasty, err := trie.NewTrie(nil, stor)
	assert.Nil(t, err)
	members, err := TraverseDynasty(dynasty)
	assert.Nil(t, err)
	assert.Equal(t, members, []byteutils.Hash{})
}

func TestInitialDynastyNotEnough(t *testing.T) {
	neb := mockNeb(t)
	neb.genesis.Consensus.Dpos.Dynasty = []string{}
	chain, err := core.NewBlockChain(neb)
	assert.Nil(t, err)
	assert.Equal(t, chain.Setup(neb), core.ErrGenesisConfNotMatch)
	neb.storage, _ = storage.NewMemoryStorage()
	chain, err = core.NewBlockChain(neb)
	assert.Nil(t, err)
	assert.Equal(t, chain.Setup(neb), ErrInitialDynastyNotEnough)
}
