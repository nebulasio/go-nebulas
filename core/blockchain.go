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
	"github.com/sirupsen/logrus"
)

// storage: key -> value
// scheme -> scheme version
// genesis hash -> genesis block
// blockchain_tail -> tail block hash
// block hash -> block
// height -> block hash

// BlockChain the BlockChain core type.
type BlockChain struct {
	chainID uint32

	genesis *corepb.Genesis

	genesisBlock *Block
	tailBlock    *Block

	bkPool *BlockPool
	txPool *TransactionPool

	consensusHandler Consensus
	syncService      SyncService

	cachedBlocks       *lru.Cache
	detachedTailBlocks *lru.Cache

	latestIrreversibleBlock *Block

	storage storage.Storage
	neb     Neblet

	eventEmitter *EventEmitter

	quitCh chan int
}

const (
	// TestNetID chain id for test net.
	TestNetID = 1

	// EagleNebula chain id for 1.x
	EagleNebula = 1 << 4

	// ChunkSize is the size of blocks in a chunk
	ChunkSize = 32

	// Tail Key in storage
	Tail = "blockchain_tail"

	// LIB (latest irreversible block) in storage
	LIB = "blockchain_lib"
)

// NewBlockChain create new #BlockChain instance.
func NewBlockChain(neb Neblet) (*BlockChain, error) {
	blockPool, err := NewBlockPool(1024)
	if err != nil {
		return nil, err
	}
	txPool, err := NewTransactionPool(40960)
	if err != nil {
		return nil, err
	}
	txPool.setEventEmitter(neb.EventEmitter())

	var bc = &BlockChain{
		chainID:      neb.Genesis().Meta.ChainId,
		genesis:      neb.Genesis(),
		bkPool:       blockPool,
		txPool:       txPool,
		storage:      neb.Storage(),
		neb:          neb,
		eventEmitter: neb.EventEmitter(),
		quitCh:       make(chan int, 1),
	}

	bc.cachedBlocks, _ = lru.NewWithEvict(4096, func(key interface{}, value interface{}) {
		block := value.(*Block)
		if block != nil {
			block.Dispose()
		}
	})

	bc.detachedTailBlocks, _ = lru.NewWithEvict(1024, func(key interface{}, value interface{}) {
		block := value.(*Block)
		if block != nil {
			block.Dispose()
		}
	})

	if err := bc.CheckChainConfig(neb); err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"meta.chainid":           neb.Genesis().Meta.ChainId,
			"consensus.dpos.dynasty": neb.Genesis().Consensus.Dpos.Dynasty,
			"token.distribution":     neb.Genesis().TokenDistribution,
		}).Error("Found unmatched genesis configuration.")
		return nil, err
	}

	logging.CLog().WithFields(logrus.Fields{
		"meta.chainid":           neb.Genesis().Meta.ChainId,
		"consensus.dpos.dynasty": neb.Genesis().Consensus.Dpos.Dynasty,
		"token.distribution":     neb.Genesis().TokenDistribution,
	}).Info("Genesis Configuration.")

	bc.genesisBlock, err = bc.loadGenesisFromStorage()
	if err != nil {
		return nil, err
	}

	bc.tailBlock, err = bc.loadTailFromStorage()
	if err != nil {
		return nil, err
	}
	logging.CLog().WithFields(logrus.Fields{
		"tail": bc.tailBlock,
	}).Info("Tail Block.")

	bc.latestIrreversibleBlock, err = bc.loadLIBFromStorage()
	if err != nil {
		return nil, err
	}
	logging.CLog().WithFields(logrus.Fields{
		"block": bc.latestIrreversibleBlock,
	}).Info("Latest Irreversible Block.")

	bc.bkPool.setBlockChain(bc)
	bc.txPool.setBlockChain(bc)

	return bc, nil
}

// Start start loop.
func (bc *BlockChain) Start() {
	logging.CLog().Info("Starting BlockChain...")

	go bc.loop()
}

// Stop stop loop.
func (bc *BlockChain) Stop() {
	logging.CLog().Info("Stopping BlockChain...")
	bc.quitCh <- 0
}

func (bc *BlockChain) loop() {
	logging.CLog().Info("Started BlockChain.")
	timerChan := time.NewTicker(5 * time.Second).C
	for {
		select {
		case <-bc.quitCh:
			logging.CLog().Info("Stopped BlockChain.")
			return
		case <-timerChan:
			bc.updateLatestIrreversibleBlock(bc.tailBlock)
		}
	}
}

// CheckChainConfig check if the genesis and config is valid
func (bc *BlockChain) CheckChainConfig(neb Neblet) error {
	if neb.Config().Chain.ChainId != neb.Genesis().Meta.ChainId {
		return ErrInvalidConfigChainID
	}

	if genesis, _ := DumpGenesis(bc.storage); genesis != nil {
		if neb.Genesis().Meta.ChainId != genesis.Meta.ChainId {
			return ErrGenesisConfNotMatch
		}

		if len(neb.Genesis().Consensus.Dpos.Dynasty) != len(genesis.Consensus.Dpos.Dynasty) {
			return ErrGenesisConfNotMatch
		}

		if len(neb.Genesis().TokenDistribution) != len(genesis.TokenDistribution) {
			return ErrGenesisConfNotMatch
		}

		// check dpos equal
		for _, confDposAddr := range neb.Genesis().Consensus.Dpos.Dynasty {
			contains := false
			for _, dposAddr := range genesis.Consensus.Dpos.Dynasty {
				if dposAddr == confDposAddr {
					contains = true
					break
				}
			}
			if !contains {
				return ErrGenesisConfNotMatch
			}

		}

		// check distribution equal
		for _, confDistribution := range neb.Genesis().TokenDistribution {
			contains := false
			for _, distribution := range genesis.TokenDistribution {
				if distribution.Address == confDistribution.Address &&
					distribution.Value == confDistribution.Value {
					contains = true
					break
				}
			}
			if !contains {
				return ErrGenesisConfNotMatch
			}
		}
	}

	return nil
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

func (bc *BlockChain) revertBlocks(from *Block, to *Block) error {
	reverted := to
	var revertTimes int64
	for revertTimes = 0; !reverted.Hash().Equals(from.Hash()); {
		if reverted.Hash().Equals(bc.latestIrreversibleBlock.Hash()) {
			return ErrCannotRevertLIB
		}
		reverted.ReturnTransactions()
		logging.VLog().WithFields(logrus.Fields{
			"block": reverted,
		}).Warn("A block is reverted.")
		revertTimes++

		reverted = bc.GetBlock(reverted.header.parentHash)
		if reverted == nil {
			return ErrMissingParentBlock
		}
	}
	// record count of reverted blocks
	if revertTimes > 0 {
		metricsBlockRevertTimesGauge.Update(revertTimes)
		metricsBlockRevertMeter.Mark(1)
	}
	return nil
}

func (bc *BlockChain) buildIndexByBlockHeight(from *Block, to *Block) error {
	for !to.Hash().Equals(from.Hash()) {
		err := bc.storage.Put(byteutils.FromUint64(to.height), to.Hash())
		if err != nil {
			return err
		}
		to = bc.GetBlock(to.header.parentHash)
		if to == nil {
			return ErrMissingParentBlock
		}
	}
	return nil
}

// SetTailBlock set tail block.
func (bc *BlockChain) SetTailBlock(newTail *Block) error {
	// startAt := time.Now().Unix()

	oldTail := bc.tailBlock
	ancestor, err := bc.FindCommonAncestorWithTail(newTail)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"target": newTail,
			"tail":   oldTail,
		}).Debug("Failed to find common ancestor with tail")
		return err
	}
	// foundAt := time.Now().Unix()

	if err := bc.revertBlocks(ancestor, oldTail); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"from":  ancestor,
			"to":    oldTail,
			"range": "(from, to]",
		}).Debug("Failed to revert blocks.")
		return err
	}
	// revertedAt := time.Now().Unix()

	// build index by block height
	if err := bc.buildIndexByBlockHeight(ancestor, newTail); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"from":  ancestor,
			"to":    newTail,
			"range": "(from, to]",
		}).Debug("Failed to build index by block height.")
		return err
	}
	// builtAt := time.Now().Unix()

	// record new tail
	if err := bc.storeTailToStorage(newTail); err != nil {
		return err
	}
	bc.tailBlock = newTail
	// storedAt := time.Now().Unix()

	metricsBlockHeightGauge.Update(int64(newTail.Height()))
	metricsBlocktailHashGauge.Update(int64(byteutils.HashBytes(newTail.Hash())))

	// endAt := time.Now().Unix()
	/* 	logging.VLog().WithFields(logrus.Fields{
		"time.foundAncestor":    foundAt - startAt,
		"time.revertBlocks":     revertedAt - foundAt,
		"time.buildHeightIndex": builtAt - revertedAt,
		"time.storeNewTail":     storedAt - builtAt,
		"time.all":              endAt - startAt,
		"tail.old":              oldTail,
		"tail.new":              newTail,
	}).Info("Succeed to set tail block.") */

	return nil
}

func (bc *BlockChain) updateLatestIrreversibleBlock(tail *Block) {
	// startAt := time.Now().Unix()

	lib := bc.latestIrreversibleBlock
	cur := tail
	miners := make(map[string]bool)
	dynasty := int64(0)
	for !cur.Hash().Equals(lib.Hash()) {
		curDynasty := cur.header.timestamp / DynastyInterval
		if curDynasty != dynasty {
			miners = make(map[string]bool)
			dynasty = curDynasty
		}
		if int(cur.height)-int(lib.height) < ConsensusSize-len(miners) {
			logging.VLog().WithFields(logrus.Fields{
				"tail": tail,
				"lib":  lib,
				"cur":  cur,
				// "time":             time.Now().Unix() - startAt,
				"err":              "supported miners is not enough",
				"miners.limit":     ConsensusSize,
				"miners.supported": len(miners),
			}).Debug("Failed to update latest irreversible block.")
			return
		}
		miners[cur.miner.String()] = true
		if len(miners) >= ConsensusSize {
			if err := bc.storeLIBToStorage(cur); err != nil {
				logging.VLog().WithFields(logrus.Fields{
					"tail": tail,
					"lib":  cur,
				}).Debug("Failed to store latest irreversible block.")
				return
			}
			logging.VLog().WithFields(logrus.Fields{
				"lib.new": cur,
				"lib.old": bc.latestIrreversibleBlock,
				"tail":    tail,
				// "time":             time.Now().Unix() - startAt,
				"miners.limit":     ConsensusSize,
				"miners.supported": len(miners),
			}).Info("Succeed to update latest irreversible block.")
			bc.latestIrreversibleBlock = cur
			return
		}

		tmp := cur
		cur = bc.GetBlock(cur.header.parentHash)
		if cur == nil || CheckGenesisBlock(cur) {
			logging.VLog().WithFields(logrus.Fields{
				"tail": tail,
				"cur":  tmp,
			}).Debug("Failed to find latest irreversible block.")
			return
		}
	}

	logging.VLog().WithFields(logrus.Fields{
		"cur":  cur,
		"lib":  bc.latestIrreversibleBlock,
		"tail": tail,
		// "time":             time.Now().Unix() - startAt,
		"err":              "supported miners is not enough",
		"miners.limit":     ConsensusSize,
		"miners.supported": len(miners),
	}).Warn("Failed to update latest irreversible block.")

	e := &Event{
		Topic: TopicLibBlock,
		Data:  bc.latestIrreversibleBlock.String(),
	}
	bc.eventEmitter.Trigger(e)
}

// LatestIrreversibleBlock return the latest irreversible block
func (bc *BlockChain) LatestIrreversibleBlock() *Block {
	return bc.latestIrreversibleBlock
}

// GetBlockOnCanonicalChainByHeight return block in given height
func (bc *BlockChain) GetBlockOnCanonicalChainByHeight(height uint64) *Block {
	blockHash, err := bc.storage.Get(byteutils.FromUint64(height))
	if err != nil {
		return nil
	}
	return bc.GetBlock(blockHash)
}

// GetBlockOnCanonicalChainByHash check if a block is on canonical chain
func (bc *BlockChain) GetBlockOnCanonicalChainByHash(blockHash byteutils.Hash) *Block {
	blockByHash := bc.GetBlock(blockHash)
	if blockByHash == nil {
		logging.VLog().WithFields(logrus.Fields{
			"hash": blockHash.Hex(),
			"tail": bc.tailBlock,
			"err":  "cannot find block with the given hash in local storage",
		}).Debug("Failed to check a block on canonical chain.")
		return nil
	}
	blockByHeight := bc.GetBlockOnCanonicalChainByHeight(blockByHash.height)
	if blockByHeight == nil {
		logging.VLog().WithFields(logrus.Fields{
			"height": blockByHeight.Height(),
			"tail":   bc.tailBlock,
			"err":    "cannot find block with the given height in local storage",
		}).Debug("Failed to check a block on canonical chain.")
		return nil
	}
	if !blockByHeight.Hash().Equals(blockByHash.Hash()) {
		logging.VLog().WithFields(logrus.Fields{
			"blockByHash":   blockByHash,
			"blockByHeight": blockByHeight,
			"tail":          bc.tailBlock,
			"err":           "block with the given hash isn't on canonical chain",
		}).Debug("Failed to check a block on canonical chain.")
		return nil
	}
	return blockByHeight
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
func (bc *BlockChain) FetchDescendantInCanonicalChain(n int, block *Block) ([]*Block, error) {
	// get tail in canonical chain
	curHeight := block.height + 1
	tailHeight := bc.tailBlock.height
	index := uint64(0)
	res := []*Block{}
	for curHeight+index <= tailHeight && index < uint64(n) {
		block := bc.GetBlockOnCanonicalChainByHeight(curHeight + index)
		if block == nil {
			logging.VLog().WithFields(logrus.Fields{
				"err":    ErrCannotFindBlockAtGivenHeight,
				"height": strconv.Itoa(int(curHeight + index)),
			}).Debug("Failed to fetch descendant.")
			return nil, ErrCannotFindBlockAtGivenHeight
		}
		res = append(res, block)
		index++
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

// SetSyncService set sync service.
func (bc *BlockChain) SetSyncService(syncService SyncService) {
	bc.syncService = syncService
}

// StartActiveSync start active sync task
func (bc *BlockChain) StartActiveSync() bool {
	if bc.syncService.StartActiveSync() {
		bc.consensusHandler.SuspendMining()
		go func() {
			bc.syncService.WaitingForFinish()
			bc.consensusHandler.ResumeMining()
		}()
		return true
	}
	return false
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
		bc.cachedBlocks.Add(v.Hash().Hex(), v)
		if err := bc.storeBlockToStorage(v); err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"block": v,
				"err":   err,
			}).Error("Failed to store the verified block.")
			return err
		}

		logging.VLog().WithFields(logrus.Fields{
			"block": v,
		}).Info("Accepted the new block on chain")

		metricsBlockOnchainTimer.Update(time.Duration(time.Now().Unix() - v.Timestamp()))
		for _, tx := range v.transactions {
			metricsTxOnchainTimer.Update(time.Duration(time.Now().Unix() - tx.Timestamp()))
		}
	}
	for _, v := range tailBlocks {
		bc.detachedTailBlocks.Add(v.Hash().Hex(), v)
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
	gas, _, err := tx.LocalExecution(bc.tailBlock)
	return gas, err
}

// Call returns the transaction call result
func (bc *BlockChain) Call(tx *Transaction) (string, error) {
	_, result, err := tx.LocalExecution(bc.tailBlock)
	return result, err
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

func (bc *BlockChain) storeTailToStorage(block *Block) error {
	return bc.storage.Put([]byte(Tail), block.Hash())
}

func (bc *BlockChain) storeLIBToStorage(block *Block) error {
	return bc.storage.Put([]byte(LIB), block.Hash())
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

		if err := bc.storeTailToStorage(genesis); err != nil {
			return nil, err
		}

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
		heightKey := byteutils.FromUint64(genesis.height)
		if err := bc.storage.Put(heightKey, genesis.Hash()); err != nil {
			return nil, err
		}
	}
	return genesis, nil
}

func (bc *BlockChain) loadLIBFromStorage() (*Block, error) {
	hash, err := bc.storage.Get([]byte(LIB))
	if err != nil && err != storage.ErrKeyNotFound {
		return nil, err
	}

	if err == storage.ErrKeyNotFound {
		return bc.genesisBlock, nil
	}

	return LoadBlockFromStorage(hash, bc.storage, bc.txPool, bc.eventEmitter)
}
