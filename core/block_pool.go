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
	"strconv"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	lru "github.com/hashicorp/golang-lru"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/net"
	"github.com/nebulasio/go-nebulas/net/p2p"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/nebulasio/go-nebulas/util/logging"
	metrics "github.com/rcrowley/go-metrics"
	"github.com/sirupsen/logrus"
)

// constants
const (
	NoSender = ""
)

// Errors in block
var (
	duplicatedBlockCounter = metrics.GetOrRegisterCounter("neb.block.duplicated", nil)
	invalidBlockCounter    = metrics.GetOrRegisterCounter("neb.block.invalid", nil)
	BlockExecutedTimer     = metrics.GetOrRegisterTimer("neb.block.executed", nil)
	TxExecutedTimer        = metrics.GetOrRegisterTimer("neb.tx.executed", nil)
)

// BlockPool a pool of all received blocks from network.
// Blocks will be sent to Consensus when it passes signature verification.
type BlockPool struct {
	size                          int
	receiveBlockMessageCh         chan net.Message
	receiveDownloadBlockMessageCh chan net.Message
	receivedLinkedBlockCh         chan *Block
	quitCh                        chan int

	bc    *BlockChain
	cache *lru.Cache
	slot  *lru.Cache

	nm p2p.Manager
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
		size: size,
		receiveBlockMessageCh:         make(chan net.Message, size),
		receiveDownloadBlockMessageCh: make(chan net.Message, size),
		receivedLinkedBlockCh:         make(chan *Block, size),
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
func (pool *BlockPool) RegisterInNetwork(nm p2p.Manager) {
	nm.Register(net.NewSubscriber(pool, pool.receiveBlockMessageCh, MessageTypeNewBlock))
	nm.Register(net.NewSubscriber(pool, pool.receiveBlockMessageCh, MessageTypeDownloadedBlockReply))
	nm.Register(net.NewSubscriber(pool, pool.receiveDownloadBlockMessageCh, MessageTypeDownloadedBlock))
	pool.nm = nm
}

// Start start loop.
func (pool *BlockPool) Start() {
	logging.CLog().WithFields(logrus.Fields{
		"size": pool.size,
	}).Info("Start BlockPool.")

	go pool.loop()
}

// Stop stop loop.
func (pool *BlockPool) Stop() {
	logging.CLog().WithFields(logrus.Fields{
		"size": pool.size,
	}).Info("Stop BlockPool.")

	pool.quitCh <- 0
}

func (pool *BlockPool) handleBlock(msg net.Message) {
	if msg.MessageType() != MessageTypeNewBlock && msg.MessageType() != MessageTypeDownloadedBlockReply {
		logging.VLog().WithFields(logrus.Fields{
			"msgType": msg.MessageType(),
			"msg":     msg,
			"err":     "neither new block nor download block response msg",
		}).Warn("Received unregistered message.")
		return
	}

	block := new(Block)
	pbblock := new(corepb.Block)
	if err := proto.Unmarshal(msg.Data().([]byte), pbblock); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"msgType": msg.MessageType(),
			"msg":     msg,
			"err":     err,
		}).Error("Failed to unmarshal data.")
		return
	}
	if err := block.FromProto(pbblock); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"msgType": msg.MessageType(),
			"msg":     msg,
			"err":     err,
		}).Error("Failed to recover a block from proto data.")
		return
	}

	diff := time.Now().Unix() - block.Timestamp()
	if msg.MessageType() == MessageTypeNewBlock && int64(math.Abs(float64(diff))) > AcceptedNetWorkDelay {
		logging.VLog().WithFields(logrus.Fields{
			"block": block,
			"diff":  diff,
			"limit": AcceptedNetWorkDelay,
			"err":   "timeout",
		}).Warn("Failed to accept a timeout block.")
		return
	}

	logging.VLog().WithFields(logrus.Fields{
		"block": block,
		"type":  msg.MessageType(),
	}).Info("Received a new block.")

	if err := pool.PushAndRelay(msg.MessageFrom(), block); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"block": block,
			"err":   err,
		}).Error("Failed to push a block into block pool.")
	}
}

func (pool *BlockPool) handleDownloadedBlock(msg net.Message) {
	if msg.MessageType() != MessageTypeDownloadedBlock {
		logging.VLog().WithFields(logrus.Fields{
			"messageType": msg.MessageType(),
			"message":     msg,
			"err":         "not download block request msg",
		}).Warn("Received unregistered message.")
		return
	}

	pbDownloadBlock := new(corepb.DownloadBlock)
	if err := proto.Unmarshal(msg.Data().([]byte), pbDownloadBlock); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"msgType": msg.MessageType(),
			"msg":     msg,
			"err":     err,
		}).Error("Failed to unmarshal data.")
		return
	}

	if byteutils.Equal(pbDownloadBlock.Hash, GenesisHash) {
		logging.VLog().WithFields(logrus.Fields{
			"download.hash": byteutils.Hex(pbDownloadBlock.Hash),
		}).Warn("Asked to download genesis's parent, ignore it.")
		return
	}

	block := pool.bc.GetBlock(pbDownloadBlock.Hash)
	if block == nil {
		logging.VLog().WithFields(logrus.Fields{
			"download.hash": byteutils.Hex(pbDownloadBlock.Hash),
		}).Error("Failed to find the block asked for.")
		return
	}

	if !block.Signature().Equals(pbDownloadBlock.Sign) {
		logging.VLog().WithFields(logrus.Fields{
			"download.hash": byteutils.Hex(pbDownloadBlock.Hash),
			"download.sign": byteutils.Hex(pbDownloadBlock.Sign),
			"expect.sign":   block.Signature().Hex(),
		}).Error("Failed to check the block's signature.")
		return
	}

	parent, err := block.ParentBlock()
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"block": block,
		}).Error("Failed to find the block's parent.")
		return
	}

	pbBlock, err := parent.ToProto()
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"parent": parent,
			"err":    err,
		}).Error("Failed to convert the block's parent to proto data.")
		return
	}
	bytes, err := proto.Marshal(pbBlock)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"parent": parent,
			"err":    err,
		}).Error("Failed to marshal the block's parent.")
		return
	}
	pool.nm.SendMsg(MessageTypeDownloadedBlockReply, bytes, msg.MessageFrom())

	logging.VLog().WithFields(logrus.Fields{
		"block":  block,
		"parent": parent,
	}).Info("Responsed to the download request.")
}

func (pool *BlockPool) loop() {
	logging.CLog().Info("Launched BlockPool.")
	for {
		select {
		case <-pool.quitCh:
			logging.CLog().Info("Shutdowned BlockPool.")
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

func (pool *BlockPool) download(sender string, block *Block) error {
	downloadMsg := &corepb.DownloadBlock{
		Hash: block.Hash(),
		Sign: block.Signature(),
	}
	bytes, err := proto.Marshal(downloadMsg)
	if err != nil {
		return err
	}

	pool.nm.SendMsg(MessageTypeDownloadedBlock, bytes, sender)

	logging.VLog().WithFields(logrus.Fields{
		"target": sender,
		"block":  block,
	}).Info("Send download request.")

	return nil
}

func (pool *BlockPool) push(sender string, block *Block) error {
	logging.VLog().WithFields(logrus.Fields{
		"block": block,
	}).Info("Try to push a new block.")

	// verify non-dup block
	if pool.cache.Contains(block.Hash().Hex()) ||
		pool.bc.GetBlock(block.Hash()) != nil {
		duplicatedBlockCounter.Inc(1)
		return ErrDuplicatedBlock
	}

	// verify block integrity
	if err := block.VerifyIntegrity(pool.bc.chainID, pool.bc.ConsensusHandler()); err != nil {
		invalidBlockCounter.Inc(1)
		return err
	}

	logging.VLog().WithFields(logrus.Fields{
		"block": block,
	}).Info("Block Integrity Verified.")

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

		for plb.parentBlock != nil {
			logging.VLog().WithFields(logrus.Fields{
				"block": plb.block,
			}).Warn("Found unlinked ancestor.")
			plb = plb.parentBlock
		}

		if err := pool.download(sender, plb.block); err != nil {
			return err
		}

		return nil
	}

	// find parent in Chain.
	var parentBlock *Block
	if parentBlock = bc.GetBlock(lb.parentHash); parentBlock == nil {
		// still not found, wait to parent block from network.
		if sender == NoSender {
			return ErrMissingParentBlock
		}
		// do sync if there are so many empty slots.
		if lb.block.Timestamp()-bc.TailBlock().Timestamp() > BlockInterval*DynastySize {

			logging.CLog().WithFields(logrus.Fields{
				"tail":    bc.tailBlock,
				"offline": strconv.Itoa(int(lb.block.Timestamp()-bc.TailBlock().Timestamp())) + "s",
				"limit":   strconv.Itoa(int(BlockInterval*DynastySize)) + "s",
			}).Warn("offline too long, restart sync from others.")

			bc.Neb().StartSync()
			return nil
		}
		if err := pool.download(sender, lb.block); err != nil {
			return err
		}
		return ErrInvalidBlockCannotFindParentInLocal
	}

	// found in BlockChain, then we can verify the state root, and tell the Consensus all the tails.
	// performance depth-first search to verify state root, and get all tails.
	allBlocks, tailBlocks, err := lb.travelToLinkAndReturnAllValidBlocks(parentBlock)
	if err != nil {
		return err
	}

	if err := bc.putVerifiedNewBlocks(parentBlock, allBlocks, tailBlocks); err != nil {
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

func (lb *linkedBlock) travelToLinkAndReturnAllValidBlocks(parentBlock *Block) ([]*Block, []*Block, error) {
	if err := lb.block.LinkParentBlock(parentBlock); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"parent": parentBlock,
			"block":  lb.block,
			"err":    err,
		}).Error("Failed to link the block with its parent.")
		return nil, nil, err
	}

	if err := lb.block.VerifyExecution(parentBlock, lb.pool.bc.ConsensusHandler()); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"block": lb.block,
			"err":   err,
		}).Error("Failed to execute block.")
		return nil, nil, err
	}

	logging.VLog().WithFields(logrus.Fields{
		"block": lb.block,
	}).Info("Block Verified.")

	allBlocks := []*Block{lb.block}
	tailBlocks := []*Block{}

	if len(lb.childBlocks) == 0 {
		tailBlocks = append(tailBlocks, lb.block)
	}

	for _, clb := range lb.childBlocks {
		a, b, err := clb.travelToLinkAndReturnAllValidBlocks(lb.block)
		if err == nil {
			allBlocks = append(allBlocks, a...)
			tailBlocks = append(tailBlocks, b...)
		}
	}

	return allBlocks, tailBlocks, nil
}
