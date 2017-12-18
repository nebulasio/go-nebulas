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
	"math"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	lru "github.com/hashicorp/golang-lru"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/net"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	metrics "github.com/rcrowley/go-metrics"
	log "github.com/sirupsen/logrus"
)

// Errors in block
var (
	duplicateBlockCounter = metrics.GetOrRegisterCounter("bp_duplicate", nil)
	invalidBlockCounter   = metrics.GetOrRegisterCounter("bp_invalid", nil)
)

// BlockPool a pool of all received blocks from network.
// Blocks will be sent to Consensus when it passes signature verification.
type BlockPool struct {
	receiveMessageCh chan net.Message
	receivedBlockCh  chan *Block
	quitCh           chan int

	bc    *BlockChain
	cache *lru.Cache
	slot  map[int64]*linkedBlock

	nm net.Manager
	mu sync.RWMutex
}

type linkedBlock struct {
	block      *Block
	pool       *BlockPool
	hash       byteutils.Hash
	parentHash byteutils.Hash

	parentBlock *linkedBlock
	childBlocks map[byteutils.HexHash]*linkedBlock
}

// NewBlockPool return new #BlockPool instance.
func NewBlockPool() *BlockPool {
	bp := &BlockPool{
		receiveMessageCh: make(chan net.Message, 128),
		receivedBlockCh:  make(chan *Block, 128),
		quitCh:           make(chan int, 1),
	}
	bp.cache, _ = lru.New(1024)
	bp.slot = make(map[int64]*linkedBlock)
	return bp
}

// ReceivedBlockCh return received block chan.
func (pool *BlockPool) ReceivedBlockCh() chan *Block {
	return pool.receivedBlockCh
}

// RegisterInNetwork register message subscriber in network.
func (pool *BlockPool) RegisterInNetwork(nm net.Manager) {
	nm.Register(net.NewSubscriber(pool, pool.receiveMessageCh, MessageTypeNewBlock))
	pool.nm = nm
}

// Start start loop.
func (pool *BlockPool) Start() {
	go pool.loop()
}

// Stop stop loop.
func (pool *BlockPool) Stop() {
	pool.quitCh <- 0
}

func (pool *BlockPool) loop() {
	log.WithFields(log.Fields{
		"func": "BlockPool.loop",
	}).Debug("running.")

	count := 0
	for {
		select {
		case <-pool.quitCh:
			log.WithFields(log.Fields{
				"func": "BlockPool.loop",
			}).Info("quit.")
			return
		case msg := <-pool.receiveMessageCh:
			count++
			log.WithFields(log.Fields{
				"func": "BlockPool.loop",
			}).Debugf("received message. Count=%d", count)

			if msg.MessageType() != MessageTypeNewBlock {
				log.WithFields(log.Fields{
					"func":        "BlockPool.loop",
					"messageType": msg.MessageType(),
					"message":     msg,
				}).Error("BlockPool.loop: received unregistered message, pls check code.")
				continue
			}

			block := new(Block)
			pbblock := new(corepb.Block)
			if err := proto.Unmarshal(msg.Data().([]byte), pbblock); err != nil {
				log.Error("BlockPool.loop:: unmarshal data occurs error, ", err)
				continue
			}
			if err := block.FromProto(pbblock); err != nil {
				log.Error("BlockPool.loop:: get block from proto occurs error: ", err)
				continue
			}
			diff := time.Now().Unix() - block.Timestamp()
			if int64(math.Abs(float64(diff))) > AcceptedNetWorkDelay {
				log.WithFields(log.Fields{
					"func":        "BlockPool.loop",
					"messageType": msg.MessageType(),
					"block":       block,
					"diff":        diff,
					"limit":       AcceptedNetWorkDelay,
					"err":         "timeout",
				}).Error("BlockPool.loop: invalid block, drop it.")
				continue
			}
			if err := pool.PushAndRelay(block); err != nil {
				log.WithFields(log.Fields{
					"func":        "BlockPool.loop",
					"messageType": msg.MessageType(),
					"block":       block,
					"err":         err,
				}).Warn("BlockPool.loop: invalid block, drop it.")
			}
		}
	}
}

func mockBlockFromNetwork(block *Block) (*Block, error) {
	pbBlock, err := block.ToProto()
	if err != nil {
		return nil, err
	}
	bytes, err := proto.Marshal(pbBlock)
	if err := proto.Unmarshal(bytes, pbBlock); err != nil {
		return nil, err
	}
	block = new(Block)
	block.FromProto(pbBlock)
	return block, nil
}

// Push block into block pool
func (pool *BlockPool) Push(block *Block) error {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	block, err := mockBlockFromNetwork(block)
	if err != nil {
		return nil
	}
	pushErr := pool.push(block)
	if pushErr != nil && pushErr != ErrDuplicatedBlock {
		return pushErr
	}
	return nil
}

// PushAndRelay push block into block pool and relay it.
func (pool *BlockPool) PushAndRelay(block *Block) error {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	block, err := mockBlockFromNetwork(block)
	if err != nil {
		return nil
	}
	if err := pool.push(block); err != nil {
		return err
	}
	pool.nm.Relay(MessageTypeNewBlock, block)
	return nil
}

// PushAndBroadcast push block into block pool and broadcast it.
func (pool *BlockPool) PushAndBroadcast(block *Block) error {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	block, err := mockBlockFromNetwork(block)
	if err != nil {
		return nil
	}
	if err := pool.push(block); err != nil {
		return err
	}
	pool.nm.Broadcast(MessageTypeNewBlock, block)
	return nil
}

func (pool *BlockPool) push(block *Block) error {
	// verify non-dup block
	if pool.cache.Contains(block.Hash().Hex()) ||
		pool.bc.GetBlock(block.Hash()) != nil {
		duplicateBlockCounter.Inc(1)
		return ErrDuplicatedBlock
	}

	// verify block hash & txs
	if err := block.verifyHash(pool.bc.chainID); err != nil {
		invalidBlockCounter.Inc(1)
		return err
	}

	// verify consensus.
	if err := pool.bc.ConsensusHandler().FastVerifyBlock(block); err != nil {
		invalidBlockCounter.Inc(1)
		return err
	}

	bc := pool.bc
	cache := pool.cache

	var plb *linkedBlock
	lb := newLinkedBlock(block, pool)

	if _, exist := pool.slot[lb.block.Timestamp()]; exist {
		invalidBlockCounter.Inc(1)
		return ErrDoubleBlockMinted
	}
	pool.slot[lb.block.Timestamp()] = lb
	cache.Add(lb.hash.Hex(), lb)

	// find child block in pool.
	for _, k := range cache.Keys() {
		v, _ := cache.Get(k)
		c := v.(*linkedBlock)
		if c.parentHash.Equals(lb.hash) {
			// found child block and continue.
			c.LinkParent(lb)
		}
	}

	// find parent block in cache.
	v, _ := cache.Get(lb.parentHash.Hex())
	if v != nil {
		// found in cache.
		plb = v.(*linkedBlock)
		lb.LinkParent(plb)

		return nil
	}

	// find parent in Chain.
	var parentBlock *Block
	if parentBlock = bc.GetBlock(lb.parentHash); parentBlock == nil {
		// still not found, wait to parent block from network.
		return nil
	}

	// found in BlockChain, then we can verify the state root, and tell the Consensus all the tails.
	// performance depth-first search to verify state root, and get all tails.
	allBlocks, tailBlocks := lb.travelToLinkAndReturnAllValidBlocks(parentBlock)
	err := bc.PutVerifiedNewBlocks(allBlocks, tailBlocks)
	if err != nil {
		return err
	}

	// remove allBlocks from cache.
	for _, v := range allBlocks {
		cache.Remove(v.Hash().Hex())
		pool.bc.storeBlockToStorage(v)
	}

	// notify consensus to handle new block.
	pool.receivedBlockCh <- block

	return nil
}

func (pool *BlockPool) setBlockChain(bc *BlockChain) {
	pool.bc = bc
}

func newLinkedBlock(block *Block, pool *BlockPool) *linkedBlock {
	return &linkedBlock{
		block:       block,
		pool:        pool,
		hash:        block.Hash(),
		parentHash:  block.ParentHash(),
		parentBlock: nil,
		childBlocks: make(map[byteutils.HexHash]*linkedBlock),
	}
}

func (lb *linkedBlock) LinkParent(parentBlock *linkedBlock) {
	lb.parentBlock = parentBlock
	parentBlock.childBlocks[lb.hash.Hex()] = lb
}

func (lb *linkedBlock) travelToLinkAndReturnAllValidBlocks(parentBlock *Block) ([]*Block, []*Block) {
	if lb.block.LinkParentBlock(parentBlock) == false {
		log.WithFields(log.Fields{
			"func":        "linkedBlock.dfs",
			"parentBlock": parentBlock,
			"block":       lb.block,
		}).Fatal("link parent block fail.")
		return nil, nil
	}

	if err := lb.pool.bc.ConsensusHandler().VerifyBlock(lb.block, parentBlock); err != nil {
		log.WithFields(log.Fields{
			"func":  "BlockPool.Verify",
			"block": lb.block,
			"err":   err,
		}).Error("BlockPool.Verify: consensus verify block.")
		return nil, nil
	}

	if err := lb.block.Verify(parentBlock.header.chainID); err != nil {
		log.WithFields(log.Fields{
			"func":  "BlockPool.Verify",
			"block": lb.block,
			"err":   err,
		}).Error("BlockPool.Verify: verify block failed.")
		return nil, nil
	}

	log.WithFields(log.Fields{
		"block": lb.block,
	}).Info("Block Verified.")

	allBlocks := []*Block{lb.block}
	tailBlocks := []*Block{}

	if len(lb.childBlocks) == 0 {
		tailBlocks = append(tailBlocks, lb.block)
	}

	for _, clb := range lb.childBlocks {
		a, b := clb.travelToLinkAndReturnAllValidBlocks(lb.block)
		if a != nil && b != nil {
			allBlocks = append(allBlocks, a...)
			tailBlocks = append(tailBlocks, b...)
		}
	}

	return allBlocks, tailBlocks
}
