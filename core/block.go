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
	"fmt"
	"time"

	"github.com/nebulasio/go-nebulas/common/trie"
	"golang.org/x/crypto/sha3"

	"github.com/nebulasio/go-nebulas/util/byteutils"
	log "github.com/sirupsen/logrus"
)

const (
	// BlockHashLength define a const of the length of Hash of Block in byte.
	BlockHashLength = 32

	// BlockReward given to coinbase
	// TODO: block reward should calculates dynamic.
	BlockReward = 16
)

// Errors in block
var (
	ErrInvalidBlockHash      = errors.New("invalid block hash")
	ErrInvalidBlockStateRoot = errors.New("invalid block state root hash")
)

// BlockHeader of a block
type BlockHeader struct {
	hash       Hash
	parentHash Hash
	stateRoot  Hash
	nonce      uint64
	coinbase   *Address
	timestamp  int64
}

type blockHeaderStream struct {
	Hash       []byte
	ParentHash []byte
	StateRoot  []byte
	Nonce      uint64
	CoinBase   []byte
	TimeStamp  int64
}

// Serialize Block to bytes
func (b *BlockHeader) Serialize() ([]byte, error) {
	serializer := &byteutils.JSONSerializer{}
	data := blockHeaderStream{
		b.hash,
		b.parentHash,
		b.stateRoot,
		b.nonce,
		b.coinbase.address,
		b.timestamp,
	}
	return serializer.Serialize(data)
}

// Deserialize a block
func (b *BlockHeader) Deserialize(blob []byte) error {
	serializer := &byteutils.JSONSerializer{}
	var data blockHeaderStream
	if err := serializer.Deserialize(blob, &data); err != nil {
		return err
	}
	b.hash = data.Hash
	b.parentHash = data.ParentHash
	b.stateRoot = data.StateRoot
	b.nonce = data.Nonce
	b.coinbase = &Address{data.CoinBase}
	b.timestamp = data.TimeStamp
	return nil
}

// Block structure
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
func (block *Block) Serialize() ([]byte, error) {
	var data [][]byte
	serializer := &byteutils.JSONSerializer{}
	hir, err := block.header.Serialize()
	if err != nil {
		return nil, err
	}
	data = append(data, hir)
	tir, err := (&block.transactions).Serialize()
	if err != nil {
		return nil, err
	}
	data = append(data, tir)
	return serializer.Serialize(data)
}

// Deserialize a block
func (block *Block) Deserialize(blob []byte) error {
	var data [][]byte
	serializer := &byteutils.JSONSerializer{}
	if err := serializer.Deserialize(blob, &data); err != nil {
		return err
	}

	block.sealed = true
	block.header = &BlockHeader{}
	if err := block.header.Deserialize(data[0]); err != nil {
		return err
	}

	return block.transactions.Deserialize(data[1])
}

// NewBlock return new block.
func NewBlock(coinbase *Address, parentHash Hash, nonce uint64, stateTrie *trie.Trie, txPool *TransactionPool) *Block {
	block := &Block{
		header: &BlockHeader{
			parentHash: parentHash,
			coinbase:   coinbase,
			nonce:      nonce,
			timestamp:  time.Now().Unix(),
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

// SetTimestamp set timestamp
func (block *Block) SetTimestamp(timestamp int64) {
	if block.sealed {
		panic("Sealed block can't be changed.")
	}
	block.header.timestamp = timestamp
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

	stateTrie, err := parentBlock.stateTrie.Clone()
	if err != nil {
		log.WithFields(log.Fields{
			"func":        "linkedBlock.dfs",
			"err":         err,
			"parentBlock": parentBlock,
			"block":       block,
		}).Fatal("clone state trie from parent block fail.")
		panic("clone state trie from parent block fail.")
	}

	block.stateTrie = stateTrie
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

// Sealed return true if block seals. Otherwise return false.
func (block *Block) Sealed() bool {
	return block.sealed
}

// Seal seal block, calculate stateRoot and block hash.
func (block *Block) Seal() {
	if block.sealed {
		return
	}

	block.header.hash = HashBlock(block)
	block.header.stateRoot = HashBlockStateRoot(block)

	block.sealed = true
}

func (block *Block) String() string {
	return fmt.Sprintf("Block {height:%d; hash:%s; parentHash:%s; stateRoot:%s, nonce:%d, timestamp: %d}",
		block.height,
		byteutils.Hex(block.header.hash),
		byteutils.Hex(block.header.parentHash),
		byteutils.Hex(block.StateRoot()),
		block.header.nonce,
		block.header.timestamp,
	)
}

// Verify return block verify result, including Hash, Nonce and StateRoot.
func (block *Block) Verify(bc *BlockChain) error {
	if err := block.VerifyHash(bc); err != nil {
		return err
	}

	return block.VerifyStateRoot()
}

// VerifyHash return hash verify result.
func (block *Block) VerifyHash(bc *BlockChain) error {
	// verify nonce.
	if err := bc.ConsensusHandler().VerifyBlock(block); err != nil {
		return err
	}

	// verify hash.
	wantedHash := HashBlock(block)
	if wantedHash.Equals(block.Hash()) == false {
		log.WithFields(log.Fields{
			"func":       "Block.VerifyHash",
			"err":        ErrInvalidBlockHash,
			"block":      block,
			"wantedHash": wantedHash,
		}).Error("invalid block hash.")
		return ErrInvalidBlockHash
	}

	// verify transaction.
	for _, tx := range block.transactions {
		if err := tx.Verify(); err != nil {
			return err
		}
	}

	return nil
}

// VerifyStateRoot return hash verify result.
func (block *Block) VerifyStateRoot() error {
	sr := block.stateTrie.RootHash()
	wantedStateRoot := HashBlockStateRoot(block)
	log.Debugf("STATEROOT - VERIFY: %s from [%v] to [%v]", block.header.hash.Hex(), byteutils.Hex(sr), wantedStateRoot)

	if wantedStateRoot.Equals(block.StateRoot()) == false {
		log.WithFields(log.Fields{
			"func":            "Block.VerifyStateRoot",
			"err":             ErrInvalidBlockStateRoot,
			"block":           block,
			"wantedStateRoot": wantedStateRoot,
		}).Error("invalid block state root hash.")
		return ErrInvalidBlockStateRoot
	}

	return nil
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
	hasher.Write(byteutils.FromUint64(block.header.nonce))
	hasher.Write(block.header.coinbase.address)
	hasher.Write(byteutils.FromInt64(block.header.timestamp))

	for _, tx := range block.transactions {
		hasher.Write(tx.Hash())
	}

	return hasher.Sum(nil)
}

// HashBlockStateRoot return the hash of state trie of block.
func HashBlockStateRoot(block *Block) Hash {

	// 1st, reward coinbase.
	block.rewardCoinbase()

	// 2nd, execute transactions.
	block.executeTransactions()

	return block.stateTrie.RootHash()
}
