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

// constants
const (
	NoSender = ""
)

// Errors in block
var (
	duplicateBlockCounter = metrics.GetOrRegisterCounter("bp_duplicate", nil)
	invalidBlockCounter   = metrics.GetOrRegisterCounter("bp_invalid", nil)
)

// BlockPool a pool of all received blocks from network.
// Blocks will be sent to Consensus when it passes signature verification.
type BlockPool struct {
	receiveBlockMessageCh         chan net.Message
	receiveDownloadBlockMessageCh chan net.Message
	receivedLinkedBlockCh         chan *Block
	quitCh                        chan int

	bc    *BlockChain
	cache *lru.Cache
	slot  *lru.Cache

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
func NewBlockPool(size int) (*BlockPool, error) {
	bp := &BlockPool{
		receiveBlockMessageCh:         make(chan net.Message, 128),
		receiveDownloadBlockMessageCh: make(chan net.Message, 128),
		receivedLinkedBlockCh:         make(chan *Block, 128),
		quitCh:                        make(chan int, 1),
	}
	var err error
	bp.cache, err = lru.New(size)
	if err != nil {
		return nil, err
	}
	bp.slot, _ = lru.New(size)
	if err != nil {
		return nil, err
	}
	return bp, nil
}

// ReceivedLinkedBlockCh return received block chan.
func (pool *BlockPool) ReceivedLinkedBlockCh() chan *Block {
	return pool.receivedLinkedBlockCh
}

// RegisterInNetwork register message subscriber in network.
func (pool *BlockPool) RegisterInNetwork(nm net.Manager) {
	nm.Register(net.NewSubscriber(pool, pool.receiveBlockMessageCh, MessageTypeNewBlock))
	nm.Register(net.NewSubscriber(pool, pool.receiveBlockMessageCh, MessageTypeDownloadedBlockReply))
	nm.Register(net.NewSubscriber(pool, pool.receiveDownloadBlockMessageCh, MessageTypeDownloadedBlock))
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

func (pool *BlockPool) handleBlock(msg net.Message) {
	if msg.MessageType() != MessageTypeNewBlock && msg.MessageType() != MessageTypeDownloadedBlockReply {
		log.WithFields(log.Fields{
			"func":        "BlockPool.loop",
			"messageType": msg.MessageType(),
			"message":     msg,
		}).Error("BlockPool.loop: received unregistered message, pls check code.")
		return
	}

	block := new(Block)
	pbblock := new(corepb.Block)
	if err := proto.Unmarshal(msg.Data().([]byte), pbblock); err != nil {
		log.Error("BlockPool.loop:: unmarshal data occurs error, ", err)
		return
	}
	if err := block.FromProto(pbblock); err != nil {
		log.Error("BlockPool.loop:: get block from proto occurs error: ", err)
		return
	}
	diff := time.Now().Unix() - block.Timestamp()
	if msg.MessageType() == MessageTypeNewBlock && int64(math.Abs(float64(diff))) > AcceptedNetWorkDelay {
		log.WithFields(log.Fields{
			"func":        "BlockPool.loop",
			"messageType": msg.MessageType(),
			"block":       block,
			"diff":        diff,
			"limit":       AcceptedNetWorkDelay,
			"err":         "timeout",
		}).Error("BlockPool.loop: invalid block, drop it.")
		return
	}
	if err := pool.PushAndRelay(msg.MessageFrom(), block); err != nil {
		log.WithFields(log.Fields{
			"func":        "BlockPool.loop",
			"messageType": msg.MessageType(),
			"block":       block,
			"err":         err,
		}).Warn("BlockPool.loop: invalid block, drop it.")
	}
}

func (pool *BlockPool) handleDownloadedBlock(msg net.Message) {
	if msg.MessageType() != MessageTypeDownloadedBlock {
		log.WithFields(log.Fields{
			"func":        "BlockPool.handleDownloadedBlock",
			"messageType": msg.MessageType(),
			"message":     msg,
		}).Error("BlockPool.loop: received unregistered message, pls check code.")
		return
	}

	pbDownloadBlock := new(corepb.DownloadBlock)
	if err := proto.Unmarshal(msg.Data().([]byte), pbDownloadBlock); err != nil {
		log.Error("BlockPool.loop: unmarshal data occurs error, ", err)
		return
	}

	if byteutils.Equal(pbDownloadBlock.Hash, GenesisHash) {
		log.WithFields(log.Fields{
			"func":        "BlockPool.handleDownloadedBlock",
			"messageType": msg.MessageType(),
			"message":     msg,
		}).Warn("BlockPool.loop: received genesis block, ignore it.")
		return
	}

	block := pool.bc.GetBlock(pbDownloadBlock.Hash)
	if block == nil {
		log.WithFields(log.Fields{
			"func":        "BlockPool.handleDownloadedBlock",
			"messageType": msg.MessageType(),
			"wantedHash":  byteutils.Hex(pbDownloadBlock.Hash),
		}).Error("BlockPool.loop: received download request, but cannot find the block.")
		return
	}
	if !block.Signature().Equals(pbDownloadBlock.Sign) {
		log.WithFields(log.Fields{
			"func":         "BlockPool.handleDownloadedBlock",
			"messageType":  msg.MessageType(),
			"wantedSign":   byteutils.Hex(pbDownloadBlock.Sign),
			"expectedSign": block.Signature().Hex(),
		}).Error("BlockPool.loop: received download request, but with wrong signature.")
		return
	}
	parent, err := block.ParentBlock()
	if err != nil {
		log.WithFields(log.Fields{
			"func":        "BlockPool.handleDownloadedBlock",
			"messageType": msg.MessageType(),
			"block":       block,
		}).Error("BlockPool.loop: received download request, but cannot find the block's parent.")
		return
	}
	pbBlock, err := parent.ToProto()
	if err != nil {
		log.Error("BlockPool.loop: convert block to proto occurs error: ", err)
		return
	}
	bytes, err := proto.Marshal(pbBlock)
	if err != nil {
		log.Error("BlockPool.loop: convert block proto to bytes occurs error: ", err)
		return
	}
	pool.nm.SendMsg(MessageTypeDownloadedBlockReply, bytes, msg.MessageFrom())
	log.WithFields(log.Fields{
		"func":        "BlockPool.handleDownloadedBlock",
		"messageType": msg.MessageType(),
		"block":       block,
	}).Info("BlockPool.loop: received download request, send back the block's parent.")
}

func (pool *BlockPool) loop() {
	log.WithFields(log.Fields{
		"func": "BlockPool.loop",
	}).Debug("running.")

	for {
		select {
		case <-pool.quitCh:
			log.WithFields(log.Fields{
				"func": "BlockPool.loop",
			}).Info("quit.")
			return
		case msg := <-pool.receiveBlockMessageCh:
			pool.handleBlock(msg)
		case msg := <-pool.receiveDownloadBlockMessageCh:
			pool.handleDownloadedBlock(msg)
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
	pushErr := pool.push(NoSender, block)
	if pushErr != nil && pushErr != ErrDuplicatedBlock {
		return pushErr
	}
	return nil
}

// PushAndRelay push block into block pool and relay it.
func (pool *BlockPool) PushAndRelay(sender string, block *Block) error {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	block, err := mockBlockFromNetwork(block)
	if err != nil {
		return nil
	}
	if err := pool.push(sender, block); err != nil {
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
	if err := pool.push(NoSender, block); err != nil {
		return err
	}
	pool.nm.Broadcast(MessageTypeNewBlock, block)
	return nil
}

func (pool *BlockPool) push(sender string, block *Block) error {
	// verify non-dup block
	if pool.cache.Contains(block.Hash().Hex()) ||
		pool.bc.GetBlock(block.Hash()) != nil {
		duplicateBlockCounter.Inc(1)
		return ErrDuplicatedBlock
	}

	// verify block integrity
	if err := block.VerifyIntegrity(pool.bc.chainID, pool.bc.ConsensusHandler()); err != nil {
		invalidBlockCounter.Inc(1)
		return err
	}

	bc := pool.bc
	cache := pool.cache

	var plb *linkedBlock
	lb := newLinkedBlock(block, pool)

	if exist := pool.slot.Contains(lb.block.Timestamp()); exist {
		invalidBlockCounter.Inc(1)
		return ErrDoubleBlockMinted
	}
	pool.slot.Add(lb.block.Timestamp(), lb.block.Hash())
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
		if sender == NoSender {
			log.WithFields(log.Fields{
				"func": "BlockPool.loop",
				"err":  "cannot find the block's parent",
			}).Error("BlockPool.loop: receive block from local.")
			return nil
		}
		// do sync if there are so many empty slots.
		if lb.block.Timestamp()-bc.TailBlock().Timestamp() > BlockInterval*DynastySize {
			bc.Neb().StartSync()
			return nil
		}
		downloadMsg := &corepb.DownloadBlock{
			Hash: lb.block.Hash(),
			Sign: lb.block.Signature(),
		}
		bytes, err := proto.Marshal(downloadMsg)
		if err != nil {
			return err
		}
		pool.nm.SendMsg(MessageTypeDownloadedBlock, bytes, sender)
		log.WithFields(log.Fields{
			"func":   "BlockPool.loop",
			"target": sender,
			"hash":   lb.block.Hash().Hex(),
			"sign":   lb.block.Signature().Hex(),
		}).Info("BlockPool.loop: send download request.")
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
	pool.receivedLinkedBlockCh <- block

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
	if err := lb.block.LinkParentBlock(parentBlock); err != nil {
		log.WithFields(log.Fields{
			"func":        "BlockPool.LinkParentBlock",
			"parentBlock": parentBlock,
			"block":       lb.block,
			"err":         err,
		}).Error("link parent block fail.")
		return nil, nil
	}

	if err := lb.block.VerifyExecution(parentBlock, lb.pool.bc.ConsensusHandler()); err != nil {
		log.WithFields(log.Fields{
			"func":  "BlockPool.VerifyExecution",
			"block": lb.block,
			"err":   err,
		}).Warn("block execution fail.")
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
