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
	"strconv"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"
	lru "github.com/hashicorp/golang-lru"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/nebulasio/go-nebulas/util/logging"
	metrics "github.com/rcrowley/go-metrics"
	"github.com/sirupsen/logrus"
)

// BlockChain the BlockChain core type.
type BlockChain struct {
	chainID uint32

	genesis *corepb.Genesis

	genesisBlock *Block
	tailBlock    *Block

	bkPool           *BlockPool
	txPool           *TransactionPool
	consensusHandler Consensus

	cachedBlocks       *lru.Cache
	detachedTailBlocks *lru.Cache

	storage storage.Storage
	neb     Neblet

	eventEmitter *EventEmitter
}

const (
	// TestNetID chain id for test net.
	TestNetID = 1

	// EagleNebula chain id for 1.x
	EagleNebula = 1 << 4

	// Tail Key in storage
	Tail = "blockchain_tail"
)

var (
	blockHeightGauge      = metrics.GetOrRegisterGauge("neb.block.height", nil)
	blocktailHashGauge    = metrics.GetOrRegisterGauge("neb.block.tailhash", nil)
	blockRevertTimesGauge = metrics.GetOrRegisterGauge("neb.block.revertcount", nil)
	blockRevertMeter      = metrics.GetOrRegisterMeter("neb.block.revert", nil)
	blockOnchainTimer     = metrics.GetOrRegisterTimer("neb.block.onchain", nil)
	txOnchainTimer        = metrics.GetOrRegisterTimer("neb.tx.onchain", nil)
)

// NewBlockChain create new #BlockChain instance.
func NewBlockChain(neb Neblet) (*BlockChain, error) {
	blockPool, err := NewBlockPool(4096)
	if err != nil {
		return nil, err
	}
	txPool, err := NewTransactionPool(4096)
	if err != nil {
		return nil, err
	}

	var bc = &BlockChain{
		chainID:      neb.Genesis().Meta.ChainId,
		genesis:      neb.Genesis(),
		bkPool:       blockPool,
		txPool:       txPool,
		storage:      neb.Storage(),
		neb:          neb,
		eventEmitter: neb.EventEmitter(),
	}

	bc.cachedBlocks, _ = lru.New(1024)
	bc.detachedTailBlocks, _ = lru.New(64)

	bc.genesisBlock, err = bc.loadGenesisFromStorage()
	if err != nil {
		return nil, err
	}
	genesisConf, err := DumpGenesis(bc.storage)
	if err != nil {
		return nil, err
	}
	logging.CLog().WithFields(logrus.Fields{
		"meta.chainid":           genesisConf.Meta.ChainId,
		"consensus.dpos.dynasty": genesisConf.Consensus.Dpos.Dynasty,
		"token.distribution":     genesisConf.TokenDistribution,
	}).Info("Genesis Configuration.")

	bc.tailBlock, err = bc.loadTailFromStorage()
	if err != nil {
		return nil, err
	}
	logging.CLog().WithFields(logrus.Fields{
		"block": bc.tailBlock,
	}).Info("Tail Block.")

	bc.bkPool.setBlockChain(bc)
	bc.txPool.setBlockChain(bc)

	return bc, nil
}

// ChainID return the chainID.
func (bc *BlockChain) ChainID() uint32 {
	return bc.chainID
}

// Storage return the storage.
func (bc *BlockChain) Storage() storage.Storage {
	return bc.storage
}

// Neb return the neblet.
func (bc *BlockChain) Neb() Neblet {
	return bc.neb
}

// GenesisBlock return the genesis block.
func (bc *BlockChain) GenesisBlock() *Block {
	return bc.genesisBlock
}

// TailBlock return the tail block.
func (bc *BlockChain) TailBlock() *Block {
	return bc.tailBlock
}

// EventEmitter return the eventEmitter.
func (bc *BlockChain) EventEmitter() *EventEmitter {
	return bc.eventEmitter
}

// SetTailBlock set tail block.
func (bc *BlockChain) SetTailBlock(newTail *Block) error {
	oldTail := bc.tailBlock
	bc.tailBlock = newTail
	bc.storeTailToStorage(bc.tailBlock)
	// giveBack txs in reverted blocks to tx pool
	ancestor, err := bc.FindCommonAncestorWithTail(oldTail)
	if err != nil {
		return err
	}
	if ancestor.Hash().Equals(oldTail.Hash()) {
		// oldTail and newTail is on same chain, no reverted blocks
		// when tail change, add metrics
		blockHeightGauge.Update(int64(newTail.Height()))
		ancestorKDegree, err := bc.getAncestorHash(6)
		if err == nil {
			hashStr := byteutils.Hex(ancestorKDegree)
			hash, err := hashToInt64(hashStr)
			if err == nil {
				blocktailHashGauge.Update(hash)
			}
		}
		return nil
	}
	reverted := oldTail
	var revertTimes int64
	for revertTimes = 0; !reverted.Hash().Equals(ancestor.Hash()); {
		revertTimes++
		reverted.ReturnTransactions()
		reverted = bc.GetBlock(reverted.header.parentHash)
		if reverted == nil {
			return ErrMissingParentBlock
		}
	}
	if revertTimes > 0 {
		blockRevertTimesGauge.Update(revertTimes)
		blockRevertMeter.Mark(1)
	}
	return nil
}

func hashToInt64(hash string) (int64, error) {
	rs := []rune(hash)
	h := string(rs[len(hash)-4 : len(hash)])
	var s int64
	var err error
	if s, err = strconv.ParseInt(h, 16, 32); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"hash": hash,
			"err":  err,
		}).Error("Failed to parseInt")
		return 0, err
	}
	return s, nil
}

// FindCommonAncestorWithTail return the block's common ancestor with current tail
func (bc *BlockChain) FindCommonAncestorWithTail(block *Block) (*Block, error) {
	target := bc.GetBlock(block.Hash())
	if target == nil {
		target = bc.GetBlock(block.ParentHash())
	}
	if target == nil {
		return nil, ErrMissingParentBlock
	}
	tail := bc.TailBlock()
	for tail.Height() > target.Height() {
		tail = bc.GetBlock(tail.header.parentHash)
		if tail == nil {
			return nil, ErrMissingParentBlock
		}
	}
	for tail.Height() < target.Height() {
		target = bc.GetBlock(target.header.parentHash)
		if target == nil {
			return nil, ErrMissingParentBlock
		}
	}
	for !tail.Hash().Equals(target.Hash()) {
		tail = bc.GetBlock(tail.header.parentHash)
		target = bc.GetBlock(target.header.parentHash)
		if tail == nil || target == nil {
			return nil, ErrMissingParentBlock
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
		if CheckGenesisBlock(curBlock) {
			return nil, ErrNotBlockInCanonicalChain
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
func (bc *BlockChain) NewBlock(coinbase *Address) (*Block, error) {
	return bc.NewBlockFromParent(coinbase, bc.tailBlock)
}

// NewBlockFromParent create new block from parent block and return it.
func (bc *BlockChain) NewBlockFromParent(coinbase *Address, parentBlock *Block) (*Block, error) {
	return NewBlock(bc.chainID, coinbase, parentBlock)
}

// PutVerifiedNewBlocks put verified new blocks and tails.
func (bc *BlockChain) putVerifiedNewBlocks(parent *Block, allBlocks, tailBlocks []*Block) error {
	for _, v := range allBlocks {
		bc.cachedBlocks.ContainsOrAdd(v.Hash().Hex(), v)
		if err := bc.storeBlockToStorage(v); err != nil {
			return err
		}

		logging.CLog().WithFields(logrus.Fields{
			"block": v,
		}).Info("Accepted the new block on chain")

		blockOnchainTimer.Update(time.Duration(time.Now().Unix() - v.Timestamp()))
		for _, tx := range v.transactions {
			txOnchainTimer.Update(time.Duration(time.Now().Unix() - tx.Timestamp()))
		}
	}
	for _, v := range tailBlocks {
		bc.detachedTailBlocks.ContainsOrAdd(v.Hash().Hex(), v)
	}

	bc.detachedTailBlocks.Remove(parent.Hash().Hex())

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
		block, err := LoadBlockFromStorage(hash, bc.storage, bc.txPool, bc.eventEmitter)
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

// GasPrice returns the lowest transaction gas price.
func (bc *BlockChain) GasPrice() *util.Uint128 {
	gasPrice := TransactionMaxGasPrice
	tailBlock := bc.tailBlock
	for {
		// if the block is genesis, stop find the parent block
		if CheckGenesisBlock(tailBlock) {
			break
		}

		if len(tailBlock.transactions) > 0 {
			break
		}
		tailBlock = bc.GetBlock(tailBlock.ParentHash())
	}

	if len(tailBlock.transactions) > 0 {
		for _, tx := range tailBlock.transactions {
			if tx.gasPrice.Cmp(gasPrice.Int) < 0 {
				gasPrice = tx.gasPrice
			}
		}
	} else {
		// if no transactions have been submited, use the default gasPrice
		gasPrice = TransactionGasPrice
	}

	return gasPrice
}

// EstimateGas returns the transaction gas cost
func (bc *BlockChain) EstimateGas(tx *Transaction) (*util.Uint128, error) {
	// update gas to max for estimate
	tx.gasLimit = TransactionMaxGas

	bc.tailBlock.accState.BeginBatch()
	fromAcc := bc.tailBlock.accState.GetOrCreateUserAccount(tx.from.address)
	fromAcc.AddBalance(tx.MinBalanceRequired())
	fromAcc.AddBalance(tx.value)
	defer bc.tailBlock.accState.RollBack()
	return tx.VerifyExecution(bc.tailBlock)
}

func (bc *BlockChain) getAncestorHash(number int) (byteutils.Hash, error) {
	block := bc.tailBlock
	for i := 0; i < number; i++ {
		if !CheckGenesisBlock(block) {
			block = bc.GetBlock(block.ParentHash())
			if block == nil {
				return nil, ErrMissingParentBlock
			}
		}
	}
	return block.Hash(), nil
}

// Dump dump full chain.
func (bc *BlockChain) Dump(count int) string {
	rl := []string{}
	block := bc.tailBlock
	rl = append(rl, block.String())
	for i := 1; i < count; i++ {
		if !CheckGenesisBlock(block) {
			block = bc.GetBlock(block.ParentHash())
			rl = append(rl, block.String())
		}
	}

	rls := "[" + strings.Join(rl, ",") + "]"
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

func (bc *BlockChain) storeTailToStorage(block *Block) {
	bc.storage.Put([]byte(Tail), block.Hash())
}

func (bc *BlockChain) loadTailFromStorage() (*Block, error) {
	hash, err := bc.storage.Get([]byte(Tail))
	if err != nil && err != storage.ErrKeyNotFound {
		return nil, err
	}

	if err == storage.ErrKeyNotFound {
		genesis, err := bc.loadGenesisFromStorage()
		if err != nil {
			return nil, err
		}

		bc.storeTailToStorage(genesis)

		return genesis, nil
	}

	return LoadBlockFromStorage(hash, bc.storage, bc.txPool, bc.eventEmitter)
}

func (bc *BlockChain) loadGenesisFromStorage() (*Block, error) {
	genesis, err := LoadBlockFromStorage(GenesisHash, bc.storage, bc.txPool, bc.eventEmitter)
	if err != nil {
		genesis, err = NewGenesisBlock(bc.genesis, bc)
		if err != nil {
			return nil, err
		}
		if err := bc.storeBlockToStorage(genesis); err != nil {
			return nil, err
		}

	}
	return genesis, nil
}
