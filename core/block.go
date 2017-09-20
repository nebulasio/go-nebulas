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

	"github.com/nebulasio/go-nebulas/common/trie"
	"golang.org/x/crypto/sha3"

	"github.com/nebulasio/go-nebulas/utils/byteutils"
	log "github.com/sirupsen/logrus"
)

const (
	// BlockHashLength define a const of the length of Hash of Block in byte.
	BlockHashLength = 32

	// BlockReward. TODO: block reward should calculates dynamic.
	BlockReward = 16
)

/*
BlockHeader type.
*/
type BlockHeader struct {
	hash       Hash
	parentHash Hash
	stateRoot  Hash
	nonce      uint64
	coinbase   *Address
	timestamp  time.Time
}

type BHStream struct {
	Hash       []byte
	ParentHash []byte
	StateRoot  []byte
	Nonce      uint64
	CoinBase   []byte
	Time       int64
}

// Serialize Block to bytes
func (b *BlockHeader) Serialize() ([]byte, error) {
	serializer := &byteutils.JSONSerializer{}
	data := BHStream{
		b.hash,
		b.parentHash,
		b.stateRoot,
		b.nonce,
		b.coinbase.address,
		b.timestamp.UnixNano(),
	}
	return serializer.Serialize(data)
}

// Deserialize a block
func (b *BlockHeader) Deserialize(blob []byte) error {
	serializer := &byteutils.JSONSerializer{}
	var data BHStream
	if err := serializer.Deserialize(blob, &data); err != nil {
		return err
	}
	b.hash = data.Hash
	b.parentHash = data.ParentHash
	b.stateRoot = data.StateRoot
	b.nonce = data.Nonce
	b.coinbase = &Address{data.CoinBase}
	b.timestamp = time.Unix(0, data.Time)
	return nil
}

/*
Block type.
*/
type Block struct {
	header       *BlockHeader
	transactions Transactions

	sealed       bool
	height       uint64
	parenetBlock *Block
	stateTrie    *trie.Trie
	txPool       *TransactionPool
}

// Serialize Block to bytes
func (b *Block) Serialize() ([]byte, error) {
	var data [][]byte
	serializer := &byteutils.JSONSerializer{}
	hir, err := b.header.Serialize()
	if err != nil {
		return nil, err
	}
	data = append(data, hir)
	tir, err := (&b.transactions).Serialize()
	if err != nil {
		return nil, err
	}
	data = append(data, tir)
	return serializer.Serialize(data)
}

// Deserialize a block
func (b *Block) Deserialize(blob []byte) error {
	var data [][]byte
	serializer := &byteutils.JSONSerializer{}
	if err := serializer.Deserialize(blob, &data); err != nil {
		return err
	}
	b.sealed = true
	b.header = &BlockHeader{}
	if err := b.header.Deserialize(data[0]); err != nil {
		return err
	}
	if err := b.transactions.Deserialize(data[1]); err != nil {
		return err
	}
	return nil
}

// NewBlock return new block.
func NewBlock(parentHash Hash, coinbase *Address, stateTrie *trie.Trie, txPool *TransactionPool) *Block {
	block := &Block{
		header: &BlockHeader{
			parentHash: parentHash,
			coinbase:   coinbase,
			timestamp:  time.Now(),
		},
		transactions: make(Transactions, 0),
		stateTrie:    stateTrie,
		txPool:       txPool,
		sealed:       false,
	}
	return block
}

// Nonce return nonce.
func (block *Block) Nonce() uint64 {
	return block.header.nonce
}

// SetNonce set nonce.
func (block *Block) SetNonce(nonce uint64) {
	if block.sealed {
		panic("Sealed block can't be changed.")
	}
	block.header.nonce = nonce
}

// Hash return block hash.
func (block *Block) Hash() Hash {
	return block.header.hash
}

// StateRoot return state root hash.
func (block *Block) StateRoot() Hash {
	return block.header.stateRoot
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

// AddTransactions add transactions to block.
func (block *Block) AddTransactions(txs ...*Transaction) *Block {
	if block.sealed {
		panic("Sealed block can't be changed.")
	}

	// TODO: dedup the transaction from chain.
	block.transactions = append(block.transactions, txs...)
	return block
}

// Seal seal block, calculate stateRoot and block hash.
func (block *Block) Seal() {
	if block.sealed {
		return
	}

	// 1st, reward coinbase.
	block.rewardCoinbase()

	// 2nd, execute transactions.
	block.executeTransactions()

	block.header.stateRoot = block.stateTrie.RootHash()
	block.header.hash = HashBlock(block)
	block.sealed = true
}

// Verify return the signature verification result.
func (block *Block) Verify() (bool, error) {
	// TODO: verify hash and state root hash.
	return true, nil
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

func (block *Block) rewardCoinbase() {
	stateTrie := block.stateTrie
	coinbaseAddr := block.header.coinbase.address
	origBalance := uint64(0)
	if v, _ := stateTrie.Get(coinbaseAddr); v != nil {
		origBalance = byteutils.Uint64(v)
	}
	balance := origBalance + BlockReward
	stateTrie.Put(coinbaseAddr, byteutils.FromUint64(balance))

	log.WithFields(log.Fields{
		"func":        "Block.rewardCoinbase",
		"coinbase":    block.header.coinbase,
		"origBalance": origBalance,
		"balance":     balance,
	}).Debug("assign block reward.")
}

func (block *Block) executeTransactions() {
	stateTrie := block.stateTrie

	// TODO: transaction nonce for address should be added to prevent transaction record-replay attack.
	invalidTxs := make([]int, 0)
	for idx, tx := range block.transactions {
		err := tx.Execute(stateTrie)
		if err != nil {
			log.WithFields(log.Fields{
				"err":         err,
				"func":        "Block.executeTransactions",
				"transaction": tx,
			}).Warn("execute transaction fail, remove it from block.")
			invalidTxs = append(invalidTxs, idx)
		}
	}

	// remove invalid transactions.
	txs := block.transactions
	lenOfTxs := len(block.transactions)
	for i := len(invalidTxs) - 1; i >= 0; i-- {
		idx := invalidTxs[i]

		// Put transaction back to pool.
		block.txPool.Put(txs[idx])

		// remove it from block.
		if idx == lenOfTxs-1 {
			txs = txs[:idx]
			continue
		} else if idx == 0 {
			txs = txs[0:]
			continue
		}
		txs = append(txs[:idx], txs[idx+1:]...)
	}

	block.transactions = txs
}

// HashBlock return the hash of block.
func HashBlock(block *Block) Hash {
	hasher := sha3.New256()

	hasher.Write(block.header.parentHash)
	hasher.Write(block.header.stateRoot)
	hasher.Write(byteutils.FromUint64(block.header.nonce))
	hasher.Write(block.header.coinbase.address)
	hasher.Write(byteutils.FromInt64(block.header.timestamp.UnixNano()))

	for _, tx := range block.transactions {
		hasher.Write(tx.Hash())
	}

	return hasher.Sum(nil)
}
