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
	"time"

	"github.com/nebulasio/go-nebulas/crypto/hash"
	"github.com/nebulasio/go-nebulas/utils/byteutils"
	log "github.com/sirupsen/logrus"
)

const (
	// BlockHashLength define a const of the length of Hash of Block in byte.
	BlockHashLength = 32
)

/*
BlockHeader type.
*/
type BlockHeader struct {
	hash       Hash
	parentHash Hash
	nonce      uint64
	coinbase   *Address
	timestamp  time.Time
}

/*
Block type.
*/
type Block struct {
	header       *BlockHeader
	transactions Transactions

	height       uint64
	parenetBlock *Block
}

// NewBlock return new block.
func NewBlock(parentHash Hash, coinbase *Address) *Block {
	block := &Block{
		header: &BlockHeader{
			parentHash: parentHash,
			coinbase:   coinbase,
			timestamp:  time.Now(),
		},
		transactions: make(Transactions, 10, 20),
	}
	return block
}

func (block *Block) AddTransactions(txs ...*Transaction) *Block {
	// TODO: dedup the transaction from chain.
	block.transactions = append(block.transactions, txs...)
	return block
}

// Sign signature this block.
func (block *Block) Sign() *Block {
	// TODO: Use Cipher/Key from #KeyStore by coinbase to signature this block.
	block.header.hash = HashBlock(block)
	return block
}

// VerifySign return the signature verification result.
func (block *Block) VerifySign() bool {
	// TODO: implement ECDSA verify, only verify the singature.
	return true
}

// Nonce return nonce.
func (block *Block) Nonce() uint64 {
	return block.header.nonce
}

// SetNonce set nonce.
func (block *Block) SetNonce(nonce uint64) {
	block.header.nonce = nonce
}

// Hash return block hash.
func (block *Block) Hash() Hash {
	return block.header.hash
}

// ParentHash return parent hash.
func (block *Block) ParentHash() Hash {
	return block.header.parentHash
}

// ParentBlock return parent block.
func (block *Block) ParentBlock() *Block {
	return block.parenetBlock
}

// Height return height from genesis block.
func (block *Block) Height() uint64 {
	return block.height
}

// LinkParentBlock link parent block, return true if hash is the same; false otherwise.
func (block *Block) LinkParentBlock(parentBlock *Block) bool {
	if block.ParentHash().Equals(parentBlock.Hash()) == false {
		return false
	}

	log.Infof("Block.LinkParentBlock: parentBlock %s <- block %s", parentBlock.Hash(), block.Hash())

	block.parenetBlock = parentBlock

	// travel to calculate block height.
	depth := uint64(0)
	ancestorHeight := uint64(0)
	for ancestor := block; ancestor != nil; ancestor = ancestor.parenetBlock {
		depth++
		ancestorHeight = ancestor.height
		if ancestor.height > 0 {
			break
		}
	}

	for ancestor := block; ancestor != nil && depth > 1; ancestor = ancestor.parenetBlock {
		depth--
		ancestor.height = ancestorHeight + depth
	}

	return true
}

func (block *Block) String() string {
	return fmt.Sprintf("Block {height:%d; hash:%s; parentHash:%s; nonce:%d, timestamp: %d}",
		block.height,
		byteutils.Hex(block.header.hash),
		byteutils.Hex(block.header.parentHash),
		block.header.nonce,
		block.header.timestamp.UnixNano(),
	)
}

// HashBlock return the hash of block.
func HashBlock(block *Block) []byte {
	// TODO: block.txs should be included in hash procedure.
	return hash.Sha3256(
		block.header.parentHash,
		block.header.coinbase.address,
		byteutils.FromUint64(block.header.nonce),
		byteutils.FromInt64(block.header.timestamp.UnixNano()),
	)
}
