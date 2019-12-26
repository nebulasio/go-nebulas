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
	"bytes"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	lru "github.com/hashicorp/golang-lru"
	corepb "github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/crypto"
	"github.com/nebulasio/go-nebulas/crypto/hash"
	"github.com/nebulasio/go-nebulas/crypto/keystore/secp256k1/vrf/secp256k1VRF"
	"github.com/nebulasio/go-nebulas/net"

	"github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

// constants
const (
	NoSender = ""
)

// BlockPool a pool of all received blocks from network.
// Blocks will be sent to Consensus when it passes signature verification.
type BlockPool struct {
	size                          int
	receiveBlockMessageCh         chan net.Message
	receiveDownloadBlockMessageCh chan net.Message
	quitCh                        chan int

	bc    *BlockChain
	cache *lru.Cache

	ns net.Service
	mu sync.RWMutex
}

type linkedBlock struct {
	block      *Block
	chain      *BlockChain
	hash       byteutils.Hash
	parentHash byteutils.Hash

	parentBlock *linkedBlock
	childBlocks map[byteutils.HexHash]*linkedBlock
}

// NewBlockPool return new #BlockPool instance.
func NewBlockPool(size int) (*BlockPool, error) {
	bp := &BlockPool{
		size:                          size,
		receiveBlockMessageCh:         make(chan net.Message, size),
		receiveDownloadBlockMessageCh: make(chan net.Message, size),
		quitCh:                        make(chan int, 1),
	}
	var err error
	bp.cache, err = lru.NewWithEvict(size, func(key interface{}, value interface{}) {
		lb := value.(*linkedBlock)
		if lb != nil {
			lb.Dispose()
		}
	})

	if err != nil {
		return nil, err
	}
	return bp, nil
}

// RegisterInNetwork register message subscriber in network.
func (pool *BlockPool) RegisterInNetwork(ns net.Service) {
	ns.Register(net.NewSubscriber(pool, pool.receiveBlockMessageCh, true, MessageTypeNewBlock, net.MessageWeightNewBlock))
	ns.Register(net.NewSubscriber(pool, pool.receiveBlockMessageCh, false, MessageTypeBlockDownloadResponse, net.MessageWeightZero))
	ns.Register(net.NewSubscriber(pool, pool.receiveDownloadBlockMessageCh, false, MessageTypeParentBlockDownloadRequest, net.MessageWeightZero))
	pool.ns = ns
}

// Start start loop.
func (pool *BlockPool) Start() {
	logging.CLog().WithFields(logrus.Fields{
		"size": pool.size,
	}).Info("Starting BlockPool...")

	go pool.loop()
}

// Stop stop loop.
func (pool *BlockPool) Stop() {
	logging.CLog().WithFields(logrus.Fields{
		"size": pool.size,
	}).Info("Stopping BlockPool...")

	pool.quitCh <- 0
}

func (pool *BlockPool) handleReceivedBlock(msg net.Message) {
	if msg.MessageType() != MessageTypeNewBlock && msg.MessageType() != MessageTypeBlockDownloadResponse {
		logging.VLog().WithFields(logrus.Fields{
			"msgType": msg.MessageType(),
			"msg":     msg,
			"err":     "neither new block nor download block response msg",
		}).Debug("Received unregistered message.")
		return
	}

	block := new(Block)
	pbblock := new(corepb.Block)
	if err := proto.Unmarshal(msg.Data(), pbblock); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"msgType": msg.MessageType(),
			"msg":     msg,
			"err":     err,
		}).Debug("Failed to unmarshal data.")
		return
	}
	if err := block.FromProto(pbblock); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"msgType": msg.MessageType(),
			"msg":     msg,
			"err":     err,
		}).Debug("Failed to recover a block from proto data.")
		return
	}

	if msg.MessageType() == MessageTypeNewBlock &&
		pool.bc.ConsensusHandler().CheckTimeout(block) {
		return
	}

	if msg.MessageType() == MessageTypeNewBlock &&
		pool.bc.ConsensusHandler().CheckDoubleMint(block) {
		return
	}

	logging.VLog().WithFields(logrus.Fields{
		"block": block,
		"type":  msg.MessageType(),
	}).Debug("Received a new block.")

	pool.PushAndRelay(msg.MessageFrom(), block)
}

func (pool *BlockPool) handleParentDownloadRequest(msg net.Message) {
	if msg.MessageType() != MessageTypeParentBlockDownloadRequest {
		logging.VLog().WithFields(logrus.Fields{
			"messageType": msg.MessageType(),
			"message":     msg,
			"err":         "wrong msg type",
		}).Debug("Failed to received a download request.")
		return
	}

	pbDownloadBlock := new(corepb.DownloadBlock)
	if err := proto.Unmarshal(msg.Data(), pbDownloadBlock); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"msgType": msg.MessageType(),
			"msg":     msg,
			"err":     err,
		}).Debug("Failed to unmarshal data.")
		return
	}

	if byteutils.Equal(pbDownloadBlock.Hash, GenesisHash) {
		logging.VLog().WithFields(logrus.Fields{
			"download.hash": byteutils.Hex(pbDownloadBlock.Hash),
		}).Debug("Asked to download genesis's parent, ignore it.")
		return
	}

	block := pool.bc.GetBlock(pbDownloadBlock.Hash)
	if block == nil {
		logging.VLog().WithFields(logrus.Fields{
			"download.hash": byteutils.Hex(pbDownloadBlock.Hash),
		}).Debug("Failed to find the block asked for.")
		return
	}

	if !block.Signature().Equals(pbDownloadBlock.Sign) {
		logging.VLog().WithFields(logrus.Fields{
			"download.hash": byteutils.Hex(pbDownloadBlock.Hash),
			"download.sign": byteutils.Hex(pbDownloadBlock.Sign),
			"expect.sign":   block.Signature().Hex(),
		}).Debug("Failed to check the block's signature.")
		return
	}

	parent := pool.bc.GetBlock(block.header.parentHash)
	if parent == nil {
		logging.VLog().WithFields(logrus.Fields{
			"block": block,
		}).Debug("Failed to find the block's parent.")
		return
	}

	pbBlock, err := parent.ToProto()
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"parent": parent,
			"err":    err,
		}).Debug("Failed to convert the block's parent to proto data.")
		return
	}
	bytes, err := proto.Marshal(pbBlock)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"parent": parent,
			"err":    err,
		}).Debug("Failed to marshal the block's parent.")
		return
	}
	pool.ns.SendMsg(MessageTypeBlockDownloadResponse, bytes, msg.MessageFrom(), net.MessagePriorityNormal)

	logging.VLog().WithFields(logrus.Fields{
		"block":  block,
		"parent": parent,
	}).Debug("Responsed to the download request.")
}

func (pool *BlockPool) loop() {
	logging.CLog().Info("Started BlockPool.")
	timerChan := time.NewTicker(time.Second).C
	for {
		select {
		case <-timerChan:
			metricsCachedNewBlock.Update(int64(len(pool.receiveBlockMessageCh)))
			metricsCachedDownloadBlock.Update(int64(len(pool.receiveDownloadBlockMessageCh)))
			metricsLruPoolCacheBlock.Update(int64(pool.cache.Len()))
		case <-pool.quitCh:
			logging.CLog().Info("Stopped BlockPool.")
			return
		case msg := <-pool.receiveBlockMessageCh:
			go pool.handleReceivedBlock(msg)
		case msg := <-pool.receiveDownloadBlockMessageCh:
			go pool.handleParentDownloadRequest(msg)
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
	err = block.FromProto(pbBlock)
	return block, err
}

// Push block into block pool
func (pool *BlockPool) Push(block *Block) error {
	if block == nil {
		return ErrNilArgument
	}
	pool.mu.Lock()
	defer pool.mu.Unlock()
	block, err := mockBlockFromNetwork(block)
	if err != nil {
		return err
	}
	pushErr := pool.push(NoSender, block)
	if pushErr != nil && pushErr != ErrDuplicatedBlock {
		return pushErr
	}
	return nil
}

// PushAndRelay push block into block pool and relay it.
func (pool *BlockPool) PushAndRelay(sender string, block *Block) error {
	if block == nil {
		return ErrNilArgument
	}
	pool.mu.Lock()
	defer pool.mu.Unlock()

	block, err := mockBlockFromNetwork(block)
	if err != nil {
		return err
	}

	return pool.push(sender, block)
}

// PushAndBroadcast push block into block pool and broadcast it.
func (pool *BlockPool) PushAndBroadcast(block *Block) error {
	if block == nil {
		return ErrNilArgument
	}
	pool.mu.Lock()
	defer pool.mu.Unlock()

	block, err := mockBlockFromNetwork(block)
	if err != nil {
		return err
	}

	pool.ns.Broadcast(MessageTypeNewBlock, block, net.MessagePriorityHigh)

	return pool.push(NoSender, block)
}

func (pool *BlockPool) downloadParent(sender string, block *Block) error {
	downloadMsg := &corepb.DownloadBlock{
		Hash: block.Hash(),
		Sign: block.Signature(),
	}
	bytes, err := proto.Marshal(downloadMsg)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"block": block,
			"err":   err,
		}).Debug("Failed to send download request.")
		return err
	}

	pool.ns.SendMsg(MessageTypeParentBlockDownloadRequest, bytes, sender, net.MessagePriorityNormal)

	logging.VLog().WithFields(logrus.Fields{
		"target": sender,
		"block":  block,
		"tail":   pool.bc.TailBlock(),
		"gap":    block.Height() - pool.bc.TailBlock().Height(),
		"limit":  ChunkSize,
	}).Info("Send download request.")

	return nil
}

func (pool *BlockPool) push(sender string, block *Block) error {
	// verify non-dup block
	if pool.cache.Contains(block.Hash().Hex()) ||
		pool.bc.GetBlock(block.Hash()) != nil {
		metricsDuplicatedBlock.Inc(1)
		logging.VLog().WithFields(logrus.Fields{
			"block": block,
		}).Debug("Found duplicated block.")
		return ErrDuplicatedBlock
	}

	// verify block integrity
	if err := block.VerifyIntegrity(pool.bc.chainID, pool.bc.ConsensusHandler()); err != nil {
		metricsInvalidBlock.Inc(1)
		logging.VLog().WithFields(logrus.Fields{
			"block": block,
			"err":   err,
		}).Debug("Failed to check block integrity.")
		if err == ErrInvalidBlockProposer {
			tail := pool.bc.TailBlock()
			blockSerial := pool.bc.ConsensusHandler().Serial(block.Timestamp())
			tailSerial := pool.bc.ConsensusHandler().Serial(tail.Timestamp())
			if blockSerial > tailSerial {
				return pool.startChainSync(int(block.height-tail.height), sender, block)
			}
		}
		return err
	}

	bc := pool.bc
	cache := pool.cache

	var plb *linkedBlock
	lb := newLinkedBlock(block, pool.bc)
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
	gap := 0
	v, _ := cache.Get(lb.parentHash.Hex())
	if v != nil {
		// found in cache.
		plb = v.(*linkedBlock)
		lb.LinkParent(plb)
		lb = plb
		gap++

		for lb.parentBlock != nil {
			lb = lb.parentBlock
			gap++
		}

		logging.VLog().WithFields(logrus.Fields{
			"block": plb.block,
		}).Warn("Found unlinked ancestor.")

		if sender == NoSender {
			return ErrMissingParentBlock
		}
	}

	// find parent in Chain.
	var parentBlock *Block
	if parentBlock = bc.GetBlock(lb.parentHash); parentBlock == nil {
		return pool.startChainSync(gap, sender, lb.block)
	}

	if sender != NoSender {
		pool.ns.Relay(MessageTypeNewBlock, block, net.MessagePriorityHigh)
	}

	// found in BlockChain, then we can verify the state root, and tell the Consensus all the tails.
	// performance depth-first search to verify state root, and get all tails.
	allBlocks, tailBlocks, err := lb.travelToLinkAndReturnAllValidBlocks(parentBlock)
	if err != nil {
		cache.Remove(lb.hash.Hex())
		return err
	}

	if err := bc.putVerifiedNewBlocks(parentBlock, allBlocks, tailBlocks); err != nil {
		cache.Remove(lb.hash.Hex())
		return err
	}
	// remove allBlocks from cache.
	for _, v := range allBlocks {
		cache.Remove(v.Hash().Hex())
	}

	// notify consensus to handle new block.
	return pool.bc.ConsensusHandler().ForkChoice()
}

func (pool *BlockPool) startChainSync(gap int, sender string, block *Block) error {
	// still not found, wait to parent block from network.
	if sender == NoSender {
		return ErrMissingParentBlock
	}

	// do sync if there are so many empty slots.
	if gap > ChunkSize {
		if pool.bc.StartActiveSync() {
			logging.CLog().WithFields(logrus.Fields{
				"tail":    pool.bc.tailBlock,
				"block":   block,
				"offline": gap,
				"limit":   ChunkSize,
			}).Warn("Offline too long, pend mining and restart sync from others.")
		}
		return ErrInvalidBlockCannotFindParentInLocalAndTrySync
	}
	if !pool.bc.IsActiveSyncing() {
		if err := pool.downloadParent(sender, block); err != nil {
			return err
		}
	}
	return ErrInvalidBlockCannotFindParentInLocalAndTryDownload
}

func vrfProof(block *Block, ancestorHash, parentSeed []byte) error {
	signature, err := crypto.NewSignature(block.Alg())
	if err != nil {
		return err
	}
	pub, err := signature.RecoverPublic(block.Hash(), block.Signature())
	if err != nil {
		return err
	}
	pubdata, err := pub.Encoded()
	if err != nil {
		return err
	}

	verifier, err := secp256k1VRF.NewVRFVerifierFromRawKey(pubdata)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to new VRF verifier.")
		return err
	}

	data := hash.Sha3256(ancestorHash, parentSeed)
	index, err := verifier.ProofToHash(data, block.header.random.VrfProof)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to calculate VRF proof.")
		return err
	}

	if !bytes.Equal(index[:], block.header.random.VrfSeed) {
		logging.VLog().WithFields(logrus.Fields{
			"block": block,
		}).Error("VRF proof failed.")
		return ErrVRFProofFailed
	}
	return nil
}

func (pool *BlockPool) setBlockChain(bc *BlockChain) {
	pool.bc = bc
}

func newLinkedBlock(block *Block, chain *BlockChain) *linkedBlock {
	return &linkedBlock{
		block:       block,
		chain:       chain,
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
	if err := lb.block.LinkParentBlock(lb.chain, parentBlock); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"parent": parentBlock,
			"block":  lb.block,
			"err":    err,
		}).Error("Failed to link the block with its parent.")
		return nil, nil, err
	}

	// verify vrf
	if RandomAvailableAtHeight(lb.block.height) {
		// prepare vrf inputs
		var ancestorHash, parentSeed []byte

		nob := lb.chain.ConsensusHandler().NumberOfBlocksInDynasty()
		i := uint64(0)
		tmp := lb
		for i < nob*2 && tmp != nil {
			i++
			tmp = tmp.parentBlock
		}
		if tmp == nil {
			if lb.block.height > nob*2 {
				b := lb.chain.GetBlockOnCanonicalChainByHeight(lb.block.height - nob*2)
				if b == nil {
					logging.VLog().WithFields(logrus.Fields{
						"blockHeight":          lb.block.height,
						"targetHeight":         lb.block.height - nob,
						"numOfBlocksInDynasty": nob,
					}).Error("Block not found, unexpected error.")
					metricsUnexpectedBehavior.Update(1)
					return nil, nil, ErrNotBlockInCanonicalChain
				}
				ancestorHash = b.Hash()
			} else {
				ancestorHash = lb.chain.GenesisBlock().Hash()
			}
		} else {
			ancestorHash = tmp.block.Hash()
		}

		if RandomAvailableAtHeight(parentBlock.height) {
			if !parentBlock.HasRandomSeed() {
				logging.VLog().WithFields(logrus.Fields{
					"parent": parentBlock,
				}).Error("Parent block has no random seed, unexpected error.")
				metricsUnexpectedBehavior.Update(1)
				return nil, nil, ErrInvalidBlockRandom
			}
			parentSeed = parentBlock.header.random.VrfSeed
		} else {
			parentSeed = lb.chain.GenesisBlock().Hash()
		}

		if err := vrfProof(lb.block, ancestorHash, parentSeed); err != nil {
			logging.CLog().WithFields(logrus.Fields{
				"err":      err,
				"lb.block": lb.block,
			}).Error("VRF proof failed.")
			return nil, nil, ErrVRFProofFailed
		}
	}

	if err := lb.block.VerifyExecution(); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"block": lb.block,
			"err":   err,
		}).Error("Failed to execute block.")
		return nil, nil, err
	}

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

// Dispose dispose linkedBlock
func (lb *linkedBlock) Dispose() {
	// clear pointer
	lb.block = nil
	lb.chain = nil
	// cut the relationship with children
	for _, v := range lb.childBlocks {
		v.parentBlock = nil
	}
	lb.childBlocks = nil
	// cut the relationship whit parent
	if lb.parentBlock != nil {
		delete(lb.parentBlock.childBlocks, lb.hash.Hex())
		lb.parentBlock = nil
	}
}
