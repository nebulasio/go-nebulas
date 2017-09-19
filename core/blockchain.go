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
	"github.com/hashicorp/golang-lru"
)

// BlockChain the BlockChain core type.
type BlockChain struct {
	chainID int

	genesisBlock *Block
	tailBlock    *Block

	// block pool.
	bkPool *BlockPool

	detachedBlocks *lru.Cache
}

const (
	// TestNetID chain id for test net.
	TestNetID = 1

	// EagleNebula chain id for 1.x
	EagleNebula = 1 << 4
)

// NewBlockChain create new #BlockChain instance.
func NewBlockChain(chainID int) *BlockChain {
	var bc = &BlockChain{
		chainID:      chainID,
		genesisBlock: NewGenesisBlock(),
		bkPool:       NewBlockPool(),
	}

	bc.detachedBlocks, _ = lru.New(100)
	bc.tailBlock = bc.genesisBlock

	return bc
}

// ChainID return the chainID.
func (bc *BlockChain) ChainID() int {
	return bc.chainID
}

// TailBlock return the tail block.
func (bc *BlockChain) TailBlock() *Block {
	return bc.tailBlock
}

// SetTailBlock set tail block.
func (bc *BlockChain) SetTailBlock(block *Block) {
	block.previousBlock = bc.tailBlock
	bc.tailBlock.nextBlock = block
	bc.tailBlock = block
}

// BlockPool return block pool.
func (bc *BlockChain) BlockPool() *BlockPool {
	return bc.bkPool
}

// NewBlock create new #Block instance.
func (bc *BlockChain) NewBlock(coinbase *Address) *Block {
	block := NewBlock(bc.tailBlock.header.hash, coinbase)
	return block
}
