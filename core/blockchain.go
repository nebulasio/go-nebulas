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
	"fmt"
	"strings"

	"github.com/hashicorp/golang-lru"
)

// BlockChain the BlockChain core type.
type BlockChain struct {
	chainID uint32

	genesisBlock *Block
	tailBlock    *Block

	bkPool           *BlockPool
	txPool           *TransactionPool
	consensusHandler Consensus

	cachedBlocks       *lru.Cache
	detachedTailBlocks *lru.Cache
}

const (
	// TestNetID chain id for test net.
	TestNetID = 1

	// EagleNebula chain id for 1.x
	EagleNebula = 1 << 4
)

// NewBlockChain create new #BlockChain instance.
func NewBlockChain(chainID uint32) *BlockChain {
	var bc = &BlockChain{
		chainID: chainID,
		bkPool:  NewBlockPool(),
		txPool:  NewTransactionPool(4096),
	}

	bc.cachedBlocks, _ = lru.New(1024)
	bc.detachedTailBlocks, _ = lru.New(64)

	bc.genesisBlock = NewGenesisBlock(chainID)
	bc.tailBlock = bc.genesisBlock

	bc.bkPool.setBlockChain(bc)
	bc.txPool.setBlockChain(bc)
	return bc
}

// ChainID return the chainID.
func (bc *BlockChain) ChainID() uint32 {
	return bc.chainID
}

// TailBlock return the tail block.
func (bc *BlockChain) TailBlock() *Block {
	return bc.tailBlock
}

// SetTailBlock set tail block.
func (bc *BlockChain) SetTailBlock(block *Block) {
	bc.detachedTailBlocks.Remove(block.Hash().Hex())
	bc.tailBlock = block
}

// BlockPool return block pool.
func (bc *BlockChain) BlockPool() *BlockPool {
	return bc.bkPool
}

// TransactionPool return block pool.
func (bc *BlockChain) TransactionPool() *TransactionPool {
	return bc.txPool
}

// SetConsensusHandler set consensus handler.
func (bc *BlockChain) SetConsensusHandler(handler Consensus) {
	bc.consensusHandler = handler
}

// ConsensusHandler return consensus handler.
func (bc *BlockChain) ConsensusHandler() Consensus {
	return bc.consensusHandler
}

// NewBlock create new #Block instance.
func (bc *BlockChain) NewBlock(coinbase *Address) *Block {
	return bc.NewBlockFromParent(coinbase, bc.tailBlock)
}

// NewBlockFromParent create new block from parent block and return it.
func (bc *BlockChain) NewBlockFromParent(coinbase *Address, parentBlock *Block) *Block {
	return NewBlock(bc.chainID, coinbase, parentBlock, bc.txPool)
}

// PutVerifiedNewBlocks put verified new blocks and tails.
func (bc *BlockChain) PutVerifiedNewBlocks(allBlocks, tailBlocks []*Block) {
	for _, v := range allBlocks {
		bc.cachedBlocks.ContainsOrAdd(v.Hash().Hex(), v)
	}
	for _, v := range tailBlocks {
		bc.detachedTailBlocks.ContainsOrAdd(v.Hash().Hex(), v)
	}
}

// DetachedTailBlocks return detached tail blocks, used by Fork Choice algorithm.
func (bc *BlockChain) DetachedTailBlocks() []*Block {
	ret := make([]*Block, 0)
	for _, k := range bc.detachedTailBlocks.Keys() {
		v, _ := bc.detachedTailBlocks.Get(k)
		if v != nil {
			block := v.(*Block)
			ret = append(ret, block)
		}
	}
	return ret
}

// GetBlock return block of given hash from local storage and detachedBlocks.
func (bc *BlockChain) GetBlock(hash Hash) *Block {
	// TODO: get block from local storage.
	v, _ := bc.cachedBlocks.Get(hash.Hex())
	if v == nil {
		if hash.Equals(bc.genesisBlock.Hash()) {
			return bc.genesisBlock
		}
		// TODO: load from storage for previous block.
		return nil
	}

	block := v.(*Block)
	return block
}

// Dump dump full chain.
func (bc *BlockChain) Dump() string {
	rl := make([]string, 1)
	for block := bc.tailBlock; block != nil; block = block.parenetBlock {
		rl = append(rl,
			fmt.Sprintf(
				"{%d, hash: %s, parent: %s, stateRoot: %s}",
				block.height,
				block.Hash().Hex(),
				block.ParentHash().Hex(),
				block.StateRoot().Hex(),
			))
	}
	rls := strings.Join(rl, " --> ")
	return rls
}
