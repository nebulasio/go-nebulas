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
	"errors"

	"github.com/hashicorp/golang-lru"
	log "github.com/sirupsen/logrus"
)

/*
BlockChain type.
*/
type BlockChain struct {
	chainID int

	genesisBlock *Block
	latestBlock  *Block

	detachedBlocks *lru.Cache
}

const (
	TestNetID   = 1
	EagleNebula = 1 << 4
)

/*
NewBlockChain is used to create new blockchain instance by args.
*/
func NewBlockChain(chainID int) *BlockChain {
	var bc = &BlockChain{chainID: chainID}
	bc.detachedBlocks, _ = lru.New(100)

	bc.genesisBlock = NewGenesisBlock()
	bc.latestBlock = bc.genesisBlock
	return bc
}

/*
GetChainID is used to return the ChainID in blockchain instance.
*/
func (bc *BlockChain) GetChainID() int {
	return bc.chainID
}

/*
 */
func (bc *BlockChain) Append(block *Block) (*BlockChain, error) {
	logFields := log.Fields{
		"bc.latestBlock.header.hash": bc.latestBlock.header.hash,
		"block.header.parentHash":    block.header.parentHash,
		"block.header.hash":          block.header.hash,
	}

	if bc.latestBlock.header.hash == block.header.parentHash {
		log.WithFields(logFields).Info("New block")

		block.previousBlock = bc.latestBlock
		bc.latestBlock.nextBlock = block
		bc.latestBlock = block

	} else {
		log.WithFields(logFields).Info("New forked block")

		// find the root block in detached blocks.
		rootParentBlock := block
		for {
			i, _ := bc.detachedBlocks.Get(rootParentBlock.header.parentHash)
			if ib, ok := i.(*Block); ok {
				ib.nextBlock = rootParentBlock
				rootParentBlock.previousBlock = ib
				rootParentBlock = ib
			} else {
				break
			}
		}

		// recursively find the common ancestor.
		ancestor := bc.latestBlock
		for ; ancestor != nil && ancestor.header.hash == rootParentBlock.header.hash; ancestor = ancestor.previousBlock {
			bc.detachedBlocks.Add(ancestor.header.hash, ancestor)
		}

		if ancestor == nil {
			log.WithFields(logFields).Error("No common ancestor")
			return bc, errors.New("No common ancestor")
		}

		// alter the chain.
		ancestor.nextBlock = rootParentBlock
		rootParentBlock.previousBlock = ancestor
		bc.latestBlock = block
	}

	return bc, nil
}

func (bc *BlockChain) GetLatestBlock() *Block {
	return bc.latestBlock
}

// NewBlock create new block with verified transactions, parentHash.
func (bc *BlockChain) NewBlock(coinbase *Address) *Block {
	block := NewBlock(bc.latestBlock.header.hash, bc.latestBlock.header.nonce, coinbase)
	return block
}
