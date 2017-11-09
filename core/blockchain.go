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
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/hashicorp/golang-lru"
	"github.com/nebulasio/go-nebulas/common/trie"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/core/state"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util/byteutils"
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

	storage storage.Storage
}

const (
	// TestNetID chain id for test net.
	TestNetID = 1

	// EagleNebula chain id for 1.x
	EagleNebula = 1 << 4

	// Tail Key in storage
	Tail = "blockchain_tail"
)

// NewBlockChain create new #BlockChain instance.
func NewBlockChain(chainID uint32, storage storage.Storage) (*BlockChain, error) {
	var bc = &BlockChain{
		chainID: chainID,
		bkPool:  NewBlockPool(),
		txPool:  NewTransactionPool(4096),
		storage: storage,
	}

	bc.cachedBlocks, _ = lru.New(1024)
	bc.detachedTailBlocks, _ = lru.New(64)

	tailBlock, err := bc.loadTailFromStorage()
	if err != nil {
		bc.genesisBlock = NewGenesisBlock(chainID, storage)
		if err := bc.storeBlockToStorage(bc.genesisBlock); err != nil {
			return nil, err
		}
		tailBlock = bc.genesisBlock
	}
	bc.tailBlock = tailBlock

	bc.bkPool.setBlockChain(bc)
	bc.txPool.setBlockChain(bc)
	return bc, nil
}

// ChainID return the chainID.
func (bc *BlockChain) ChainID() uint32 {
	return bc.chainID
}

// Storage return the storage
func (bc *BlockChain) Storage() storage.Storage {
	return bc.storage
}

// TailBlock return the tail block.
func (bc *BlockChain) TailBlock() *Block {
	return bc.tailBlock
}

// SetTailBlock set tail block.
func (bc *BlockChain) SetTailBlock(newTail *Block) {
	oldTail := bc.tailBlock
	bc.detachedTailBlocks.Remove(newTail.Hash().Hex())
	bc.tailBlock = newTail
	bc.storeTailToStorage(bc.tailBlock)
	// giveBack txs in reverted blocks to tx pool
	ancestor, _ := bc.FindCommonAncestorWithTail(oldTail)
	if ancestor.Hash().Equals(oldTail.Hash()) {
		// oldTail and newTail is on same chain, no reverted blocks
		return
	}
	reverted := oldTail
	for !reverted.Hash().Equals(ancestor.Hash()) {
		for _, v := range reverted.transactions {
			// give back reverted txs to tx pool
			bc.txPool.Push(v)
		}
		reverted = bc.GetBlock(reverted.header.parentHash)
		if reverted == nil {
			panic("find a block on chain, we cannot find its parent block")
		}
	}
}

// FindCommonAncestorWithTail return the block's common ancestor with current tail
func (bc *BlockChain) FindCommonAncestorWithTail(block *Block) (*Block, error) {
	target := bc.GetBlock(block.Hash())
	if target == nil {
		target = bc.GetBlock(block.ParentHash())
	}
	if target == nil {
		return nil, errors.New("cannot location the block in blockchain")
	}
	tail := bc.TailBlock()
	for tail.Height() > target.Height() {
		tail = bc.GetBlock(tail.header.parentHash)
	}
	for tail.Height() < target.Height() {
		target = bc.GetBlock(target.header.parentHash)
	}
	for !tail.Hash().Equals(target.Hash()) {
		tail = bc.GetBlock(tail.header.parentHash)
		target = bc.GetBlock(target.header.parentHash)
		if tail == nil || target == nil {
			panic("find a block on chain, we cannot find its parent block")
		}
	}
	return target, nil
}

// FetchDescendantInCanonicalChain return the subsequent blocks of the block
// lookup the block's descendant from tail to genesis
// if the block is not in canonical chain, return err
func (bc *BlockChain) FetchDescendantInCanonicalChain(n int, block *Block) ([]*Block, error) {
	curIdx := -1
	queue := make([]*Block, n)
	// get tail in canonical chain
	curBlock := bc.tailBlock
	for curBlock != nil && !curBlock.Hash().Equals(block.Hash()) {
		if curBlock.Hash().Equals(bc.genesisBlock.Hash()) {
			return nil, errors.New("cannot find the block in canonical chain")
		}
		curIdx = (curIdx + 1) % n
		queue[curIdx] = curBlock
		curBlock = bc.GetBlock(curBlock.header.parentHash)
	}
	var res []*Block
	for i := 0; curIdx >= 0 && i < n; i++ {
		if queue[curIdx] != nil {
			res = append(res, queue[curIdx])
		}
		curIdx = (curIdx + n - 1) % n
	}
	return res, nil
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
	return NewBlock(bc.chainID, coinbase, parentBlock, bc.txPool, bc.storage)
}

// PutVerifiedNewBlocks put verified new blocks and tails.
func (bc *BlockChain) PutVerifiedNewBlocks(allBlocks, tailBlocks []*Block) error {
	for _, v := range allBlocks {
		bc.cachedBlocks.ContainsOrAdd(v.Hash().Hex(), v)
		if err := bc.storeBlockToStorage(v); err != nil {
			return err
		}
	}
	for _, v := range tailBlocks {
		bc.detachedTailBlocks.ContainsOrAdd(v.Hash().Hex(), v)
	}
	return nil
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
func (bc *BlockChain) GetBlock(hash byteutils.Hash) *Block {
	// TODO: get block from local storage.
	v, _ := bc.cachedBlocks.Get(hash.Hex())
	if v == nil {
		block, err := bc.loadBlockFromStorage(hash)
		if err != nil {
			return nil
		}
		return block
	}

	block := v.(*Block)
	return block
}

// GetTransaction return transaction of given hash from local storage.
func (bc *BlockChain) GetTransaction(hash byteutils.Hash) *Transaction {
	// TODO: get transaction err handle.
	tx, err := bc.tailBlock.GetTransaction(hash)
	if err != nil {
		return nil
	}
	return tx
}

// Dump dump full chain.
func (bc *BlockChain) Dump(count int) string {
	rl := make([]string, 1)
	if count > 0 {
		i := 0
		for block := bc.tailBlock; !CheckGenesisBlock(block); block = bc.GetBlock(block.ParentHash()) {
			if i < count {
				i++
				rl = append(rl,
					fmt.Sprintf(
						"{%d, hash: %s, parent: %s, stateRoot: %s, coinbase: %s}",
						block.height,
						block.Hash().Hex(),
						block.ParentHash().Hex(),
						block.StateRoot().Hex(),
						block.header.coinbase.address.Hex(),
					))
			} else {
				break
			}
		}
	} else {
		for block := bc.tailBlock; !CheckGenesisBlock(block); block = bc.GetBlock(block.ParentHash()) {
			rl = append(rl,
				fmt.Sprintf(
					"{%d, hash: %s, parent: %s, stateRoot: %s, coinbase: %s}",
					block.height,
					block.Hash().Hex(),
					block.ParentHash().Hex(),
					block.StateRoot().Hex(),
					block.header.coinbase.address.Hex(),
				))
		}
	}

	rls := strings.Join(rl, " ")
	return rls
}

func (bc *BlockChain) storeBlockToStorage(block *Block) error {
	pbBlock, err := block.ToProto()
	if err != nil {
		return err
	}
	value, err := proto.Marshal(pbBlock)
	if err != nil {
		return err
	}
	err = bc.storage.Put(block.Hash(), value)
	if err != nil {
		return err
	}
	return nil
}

func (bc *BlockChain) loadBlockFromStorage(hash byteutils.Hash) (*Block, error) {
	value, err := bc.storage.Get(hash)
	if err != nil {
		return nil, err
	}
	pbBlock := new(corepb.Block)
	block := new(Block)
	if err = proto.Unmarshal(value, pbBlock); err != nil {
		return nil, err
	}
	if err = block.FromProto(pbBlock); err != nil {
		return nil, err
	}
	block.accState, err = state.NewAccountState(block.StateRoot(), bc.storage)
	if err != nil {
		return nil, err
	}
	block.txsTrie, err = trie.NewBatchTrie(block.TxsRoot(), bc.storage)
	if err != nil {
		return nil, err
	}
	block.storage = bc.storage
	return block, nil
}

func (bc *BlockChain) storeTailToStorage(block *Block) {
	bc.storage.Put([]byte(Tail), block.Hash())
}

func (bc *BlockChain) loadTailFromStorage() (*Block, error) {
	hash, err := bc.storage.Get([]byte(Tail))
	if err != nil {
		return nil, err
	}
	return bc.loadBlockFromStorage(hash)
}
