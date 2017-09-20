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

	"github.com/nebulasio/go-nebulas/common/trie"
	log "github.com/sirupsen/logrus"

	"github.com/nebulasio/go-nebulas/utils/byteutils"

	"github.com/hashicorp/golang-lru"
)

// BlockChain the BlockChain core type.
type BlockChain struct {
	chainID int

	genesisBlock *Block
	tailBlock    *Block

	// block pool.
	bkPool *BlockPool
	txPool *TransactionPool

	stateTrie      *trie.Trie
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
		chainID: chainID,
		bkPool:  NewBlockPool(),
		txPool:  NewTransactionPool(),
	}

	bc.stateTrie, _ = trie.NewTrie(nil)
	bc.detachedBlocks, _ = lru.New(1024)

	bc.genesisBlock = NewGenesisBlock(bc.stateTrie)
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
	block.LinkParentBlock(bc.tailBlock)
	bc.tailBlock = block
}

// BlockPool return block pool.
func (bc *BlockChain) BlockPool() *BlockPool {
	return bc.bkPool
}

// NewBlock create new #Block instance.
func (bc *BlockChain) NewBlock(coinbase *Address) *Block {
	stateTrie, err := bc.tailBlock.stateTrie.Clone()
	if err != nil {
		log.WithFields(log.Fields{
			"err":  err,
			"func": "BlockChain.NewBlock",
		}).Fatal("clone state trie fail.")
		panic("BlockChain.NewBlock: clone state trie fail.")
	}

	block := NewBlock(bc.tailBlock.header.hash, coinbase, stateTrie, bc.txPool)
	return block
}

// PutUnattachedBlocks put unattached blocks to LRU cache for furthur process.
// Unattached block is the block not yet attach to chain, eg. new block from network, local minted block.
func (bc *BlockChain) PutUnattachedBlocks(blocks ...*Block) {
	for _, v := range blocks {
		bc.detachedBlocks.Add(v.Hash().Hex(), v)
	}
}

// PutUnattachedBlockMap put unattached blocks to LRU cache for furthur process.
// Unattached block is the block not yet attach to chain, eg. new block from network, local minted block.
func (bc *BlockChain) PutUnattachedBlockMap(blocks map[HexHash]*Block) {
	for k, v := range blocks {
		bc.detachedBlocks.Add(k, v)
	}
}

// GetBlock return block of given hash from local storage and detachedBlocks.
func (bc *BlockChain) GetBlock(hash Hash) *Block {
	// TODO: get block from local storage.
	v, _ := bc.detachedBlocks.Get(hash.Hex())
	if v == nil {
		if hash.Equals(bc.genesisBlock.Hash()) {
			return bc.genesisBlock
		}
		return nil
	}

	block := v.(*Block)
	return block
}

// Dump dump full chain.
func (bc *BlockChain) Dump() string {
	rl := make([]string, 1)
	for block := bc.tailBlock; block != nil; block = block.parenetBlock {
		rl = append(rl, fmt.Sprintf("{%d, hash: %s, parent: %s}", block.height, byteutils.Hex(block.Hash()), byteutils.Hex(block.ParentHash())))
	}
	rls := strings.Join(rl, " --> ")
	return rls
}
