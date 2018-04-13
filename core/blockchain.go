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
	"strings"
	"time"

	"github.com/nebulasio/go-nebulas/core/state"

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

	// latest irreversible block
	lib *Block

	storage storage.Storage

	eventEmitter *EventEmitter

	nvm NVM

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
	if neb == nil || neb.Config() == nil || neb.Config().Chain == nil {
		return nil, ErrNilArgument
	}

	var gasPrice, gasLimit *util.Uint128
	var err error
	if 0 == len(neb.Config().Chain.GasPrice) {
		gasPrice = util.NewUint128()
	} else {
		gasPrice, err = util.NewUint128FromString(neb.Config().Chain.GasPrice)
		if err != nil {
			return nil, err
		}
	}

	if 0 == len(neb.Config().Chain.GasLimit) {
		gasLimit = util.NewUint128()
	} else {
		gasLimit, err = util.NewUint128FromString(neb.Config().Chain.GasLimit)
		if err != nil {
			return nil, err
		}
	}

	blockPool, err := NewBlockPool(128)
	if err != nil {
		return nil, err
	}
	blockPool.RegisterInNetwork(neb.NetService())

	txPool, err := NewTransactionPool(327680)
	if err != nil {
		return nil, err
	}
	txPool.setEventEmitter(neb.EventEmitter())
	if err := txPool.SetGasConfig(gasPrice, gasLimit); err != nil {
		return nil, err
	}
	txPool.RegisterInNetwork(neb.NetService())

	var bc = &BlockChain{
		chainID:      neb.Config().Chain.ChainId,
		genesis:      neb.Genesis(),
		bkPool:       blockPool,
		txPool:       txPool,
		storage:      neb.Storage(),
		eventEmitter: neb.EventEmitter(),
		nvm:          neb.Nvm(),
		quitCh:       make(chan int, 1),
	}

	bc.cachedBlocks, err = lru.New(128)
	if err != nil {
		return nil, err
	}

	bc.detachedTailBlocks, err = lru.New(128)
	if err != nil {
		return nil, err
	}

	bc.bkPool.setBlockChain(bc)
	bc.txPool.setBlockChain(bc)

	return bc, nil
}

// Setup the blockchain
func (bc *BlockChain) Setup(neb Neblet) error {
	bc.consensusHandler = neb.Consensus()

	if err := bc.CheckGenesisConfig(neb); err != nil {
		return err
	}

	var err error
	bc.genesisBlock, err = bc.LoadGenesisFromStorage()
	if err != nil {
		return err
	}

	bc.tailBlock, err = bc.LoadTailFromStorage()
	if err != nil {
		return err
	}
	logging.CLog().WithFields(logrus.Fields{
		"tail": bc.tailBlock,
	}).Info("Tail Block.")

	bc.lib, err = bc.LoadLIBFromStorage()
	if err != nil {
		return err
	}
	logging.CLog().WithFields(logrus.Fields{
		"block": bc.lib,
	}).Info("Latest Irreversible Block.")

	return nil
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
	timerChan := time.NewTicker(15 * time.Second).C
	for {
		select {
		case <-bc.quitCh:
			logging.CLog().Info("Stopped BlockChain.")
			return
		case <-timerChan:
			bc.ConsensusHandler().UpdateLIB()
			metricsLruCacheBlock.Update(int64(bc.cachedBlocks.Len()))
			metricsLruTailBlock.Update(int64(bc.detachedTailBlocks.Len()))
		}
	}
}

// CheckGenesisConfig check if the genesis and config is valid
func (bc *BlockChain) CheckGenesisConfig(neb Neblet) error {
	genesis, err := DumpGenesis(bc)
	//db.genesis has and config lack
	if neb.Genesis() == nil && err == nil {
		neb.SetGenesis(genesis)
		if neb.Config().Chain.ChainId != neb.Genesis().Meta.ChainId {
			return ErrInvalidConfigChainID
		}
	} else if neb.Genesis() == nil && err != nil {
		logging.CLog().Fatal("Failed to find genesis config in config file")
	} else if neb.Genesis() != nil && err != nil {
		//first start
		if neb.Config().Chain.ChainId != neb.Genesis().Meta.ChainId {
			return ErrInvalidConfigChainID
		}
	} else {
		if neb.Config().Chain.ChainId != neb.Genesis().Meta.ChainId {
			return ErrInvalidConfigChainID
		}

		return CheckGenesisConfByDB(genesis, neb.Genesis())
	}

	logging.CLog().WithFields(logrus.Fields{
		"meta.chainid":           neb.Genesis().Meta.ChainId,
		"consensus.dpos.dynasty": neb.Genesis().Consensus.Dpos.Dynasty,
		"token.distribution":     neb.Genesis().TokenDistribution,
	}).Info("Genesis Configuration.")
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

// GenesisBlock return the genesis block.
func (bc *BlockChain) GenesisBlock() *Block {
	return bc.genesisBlock
}

// TailBlock return the tail block.
func (bc *BlockChain) TailBlock() *Block {
	return bc.tailBlock
}

// LIB return the latest irrversible block
func (bc *BlockChain) LIB() *Block {
	return bc.lib
}

// SetLIB update the latest irrversible block
func (bc *BlockChain) SetLIB(lib *Block) {
	bc.lib = lib
}

// EventEmitter return the eventEmitter.
func (bc *BlockChain) EventEmitter() *EventEmitter {
	return bc.eventEmitter
}

func (bc *BlockChain) triggerRevertBlockEvent(blocks []string) {
	for i := len(blocks) - 1; i >= 0; i-- {
		bc.eventEmitter.Trigger(&state.Event{
			Topic: TopicRevertBlock,
			Data:  blocks[i],
		})
	}
}

func (bc *BlockChain) revertBlocks(from *Block, to *Block) error {
	reverted := to
	var revertTimes int64
	blocks := []string{}
	for revertTimes = 0; !reverted.Hash().Equals(from.Hash()); {
		if reverted.Hash().Equals(bc.lib.Hash()) {
			return ErrCannotRevertLIB
		}

		reverted.ReturnTransactions()
		logging.VLog().WithFields(logrus.Fields{
			"block": reverted,
		}).Warn("A block is reverted.")
		revertTimes++
		blocks = append(blocks, reverted.String())

		reverted = bc.GetBlock(reverted.header.parentHash)
		if reverted == nil {
			return ErrMissingParentBlock
		}
	}
	go bc.triggerRevertBlockEvent(blocks)
	// record count of reverted blocks
	if revertTimes > 0 {
		metricsBlockRevertTimesGauge.Update(revertTimes)
		metricsBlockRevertMeter.Mark(1)
	}
	return nil
}

func (bc *BlockChain) dropTxsInBlockFromTxPool(block *Block) {
	for _, tx := range block.transactions {
		bc.txPool.Del(tx)
	}
}

func (bc *BlockChain) triggerNewTailEvent(blocks []*Block) {
	for i := len(blocks) - 1; i >= 0; i-- {
		block := blocks[i]
		bc.eventEmitter.Trigger(&state.Event{
			Topic: TopicNewTailBlock,
			Data:  block.String(),
		})

		for _, v := range block.transactions {
			events, err := block.FetchEvents(v.hash)
			if err == nil {
				for _, e := range events {
					bc.eventEmitter.Trigger(e)
				}
			}
		}
	}
}

func (bc *BlockChain) buildIndexByBlockHeight(from *Block, to *Block) error {
	blocks := []*Block{}
	for !to.Hash().Equals(from.Hash()) {
		err := bc.storage.Put(byteutils.FromUint64(to.height), to.Hash())
		if err != nil {
			return err
		}
		blocks = append(blocks, to)
		go bc.dropTxsInBlockFromTxPool(to)
		to = bc.GetBlock(to.header.parentHash)
		if to == nil {
			return ErrMissingParentBlock
		}
	}
	go bc.triggerNewTailEvent(blocks)
	return nil
}

// SetTailBlock set tail block.
func (bc *BlockChain) SetTailBlock(newTail *Block) error {
	if newTail == nil {
		return ErrNilArgument
	}
	oldTail := bc.tailBlock
	ancestor, err := bc.FindCommonAncestorWithTail(newTail)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"target": newTail,
			"tail":   oldTail,
		}).Debug("Failed to find common ancestor with tail")
		return err
	}

	if err := bc.revertBlocks(ancestor, oldTail); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"from":  ancestor,
			"to":    oldTail,
			"range": "(from, to]",
		}).Debug("Failed to revert blocks.")
		return err
	}

	// build index by block height
	if err := bc.buildIndexByBlockHeight(ancestor, newTail); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"from":  ancestor,
			"to":    newTail,
			"range": "(from, to]",
		}).Debug("Failed to build index by block height.")
		return err
	}

	// record new tail
	if err := bc.StoreTailHashToStorage(newTail); err != nil { // Refine: rename, delete ToStorage
		return err
	}
	bc.tailBlock = newTail

	logging.CLog().WithFields(logrus.Fields{
		"tail": newTail,
	}).Info("Succeed to update new tail.")

	metricsBlockHeightGauge.Update(int64(newTail.Height()))
	metricsBlocktailHashGauge.Update(int64(byteutils.HashBytes(newTail.Hash())))

	return nil
}

// GetBlockOnCanonicalChainByHeight return block in given height
func (bc *BlockChain) GetBlockOnCanonicalChainByHeight(height uint64) *Block {

	if height > bc.tailBlock.height {
		return nil
	}

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
	if block == nil {
		return nil, ErrNilArgument
	}
	target := bc.GetBlock(block.Hash())
	if target == nil {
		target = bc.GetBlock(block.ParentHash())
	}
	if target == nil {
		return nil, ErrMissingParentBlock
	}

	tail := bc.TailBlock()
	if tail.Height() > target.Height() {
		tail = bc.GetBlockOnCanonicalChainByHeight(target.Height())
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
	if coinbase == nil {
		return nil, ErrInvalidArgument
	}
	return bc.NewBlockFromParent(coinbase, bc.tailBlock)
}

// NewBlockFromParent create new block from parent block and return it.
func (bc *BlockChain) NewBlockFromParent(coinbase *Address, parentBlock *Block) (*Block, error) {
	if parentBlock == nil || coinbase == nil {
		return nil, ErrNilArgument
	}
	return NewBlock(bc.chainID, coinbase, parentBlock)
}

// PutVerifiedNewBlocks put verified new blocks and tails.
func (bc *BlockChain) putVerifiedNewBlocks(parent *Block, allBlocks, tailBlocks []*Block) error {
	for _, v := range allBlocks {
		bc.cachedBlocks.Add(v.Hash().Hex(), v)
		if err := bc.StoreBlockToStorage(v); err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"block": v,
				"err":   err,
			}).Debug("Failed to store the verified block.")
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
	v, _ := bc.cachedBlocks.Get(hash.Hex())
	if v == nil {
		block, err := LoadBlockFromStorage(hash, bc)
		if err != nil {
			return nil
		}
		return block
	}

	block := v.(*Block)
	return block
}

// GetTransaction return transaction of given hash from local storage.
func (bc *BlockChain) GetTransaction(hash byteutils.Hash) (*Transaction, error) {
	worldState, err := bc.TailBlock().WorldState().Clone()
	if err != nil {
		return nil, err
	}
	tx, err := GetTransaction(hash, worldState)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// GasPrice returns the lowest transaction gas price.
func (bc *BlockChain) GasPrice() *util.Uint128 {
	gasPrice := TransactionMaxGasPrice
	tailBlock := bc.TailBlock()
	// search latest block who has transactions, try 128 times at most
	for i := 0; i < 128; i++ {
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
			if tx.gasPrice.Cmp(gasPrice) < 0 {
				gasPrice = tx.gasPrice
			}
		}
	} else {
		// if no transactions have been submitted, use the default gasPrice
		gasPrice = TransactionGasPrice
	}

	return gasPrice
}

// SimulateResult the result of simulating transaction execution
type SimulateResult struct {
	GasUsed *util.Uint128
	Msg     string
	Err     error
}

// SimulateTransactionExecution execute transaction in sandbox and rollback all changes, used to EstimateGas and Call api.
func (bc *BlockChain) SimulateTransactionExecution(tx *Transaction) (*SimulateResult, error) {
	if tx == nil {
		return nil, ErrInvalidArgument
	}

	// create block.
	block, err := bc.NewBlock(GenesisCoinbase)
	if err != nil {
		return nil, err
	}
	defer block.RollBack()

	// simulate execution.
	return tx.simulateExecution(block)
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

// StoreBlockToStorage store block
func (bc *BlockChain) StoreBlockToStorage(block *Block) error {
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

// StoreTailHashToStorage store tail block hash
func (bc *BlockChain) StoreTailHashToStorage(block *Block) error { // ToRefine, update func to StoreTailHashToStorage
	return bc.storage.Put([]byte(Tail), block.Hash())
}

// StoreLIBHashToStorage store LIB block hash
func (bc *BlockChain) StoreLIBHashToStorage(block *Block) error {
	return bc.storage.Put([]byte(LIB), block.Hash())
}

// LoadTailFromStorage load tail block
func (bc *BlockChain) LoadTailFromStorage() (*Block, error) {
	hash, err := bc.storage.Get([]byte(Tail))
	if err != nil && err != storage.ErrKeyNotFound {
		return nil, err
	}
	if err == storage.ErrKeyNotFound {
		genesis, err := bc.LoadGenesisFromStorage()
		if err != nil {
			return nil, err
		}

		if err := bc.StoreTailHashToStorage(genesis); err != nil {
			return nil, err
		}

		return genesis, nil
	}

	return LoadBlockFromStorage(hash, bc)
}

// LoadGenesisFromStorage load genesis
func (bc *BlockChain) LoadGenesisFromStorage() (*Block, error) { // ToRefine, remove or ?
	genesis, err := LoadBlockFromStorage(GenesisHash, bc)
	if err != nil && err != storage.ErrKeyNotFound {
		return nil, err
	}
	if err == storage.ErrKeyNotFound {
		genesis, err = NewGenesisBlock(bc.genesis, bc)
		if err != nil {
			return nil, err
		}
		if err := bc.StoreBlockToStorage(genesis); err != nil {
			return nil, err
		}
		heightKey := byteutils.FromUint64(genesis.height)
		if err := bc.storage.Put(heightKey, genesis.Hash()); err != nil {
			return nil, err
		}
	}
	return genesis, nil
}

// LoadLIBFromStorage load LIB
func (bc *BlockChain) LoadLIBFromStorage() (*Block, error) {
	hash, err := bc.storage.Get([]byte(LIB))
	if err != nil && err != storage.ErrKeyNotFound {
		return nil, err
	}

	if err == storage.ErrKeyNotFound {
		if err := bc.StoreLIBHashToStorage(bc.genesisBlock); err != nil {
			return nil, err
		}
		return bc.genesisBlock, nil
	}

	return LoadBlockFromStorage(hash, bc)
}
