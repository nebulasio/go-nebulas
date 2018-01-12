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

package sync

import (
	"math"
	"time"

	"github.com/nebulasio/go-nebulas/util/byteutils"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/common/trie"
	"github.com/nebulasio/go-nebulas/consensus"
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/net"
	"github.com/nebulasio/go-nebulas/net/p2p"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/sync/pb"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

// const
const (
	DescendantCount = 10
)

var (
	batch       = uint64(0)
	msgErrCount = 0
)

// Manager is used to manage the sync service
type Manager struct {
	blockChain             *core.BlockChain
	consensus              consensus.Consensus
	ns                     p2p.Manager
	quitCh                 chan bool
	syncCh                 chan bool
	receiveTailCh          chan net.Message
	receiveSyncReplyCh     chan net.Message
	cacheList              map[string]*NetBlocks
	endSyncCh              chan bool
	curTail                *core.Block
	canSyncWithBlockListCh chan bool
	goParentSyncCh         chan bool
}

// NewManager new sync manager
func NewManager(blockChain *core.BlockChain, consensus consensus.Consensus, ns p2p.Manager) *Manager {
	m := &Manager{
		blockChain,
		consensus,
		ns,
		make(chan bool, 1),
		make(chan bool, 1),
		make(chan net.Message, 128),
		make(chan net.Message, 128),
		make(map[string]*NetBlocks),
		make(chan bool, 1),
		blockChain.TailBlock(),
		make(chan bool, 1),
		make(chan bool, 1),
	}
	m.RegisterSyncBlockInNetwork(ns)
	m.RegisterSyncReplyInNetwork(ns)
	return m
}

// RegisterSyncBlockInNetwork register message subscriber in network.
func (m *Manager) RegisterSyncBlockInNetwork(nm p2p.Manager) {
	nm.Register(net.NewSubscriber(m, m.receiveTailCh, net.SyncBlock))
}

// RegisterSyncReplyInNetwork register message subscriber in network.
func (m *Manager) RegisterSyncReplyInNetwork(nm p2p.Manager) {
	nm.Register(net.NewSubscriber(m, m.receiveSyncReplyCh, net.SyncReply))
}

// Start start sync service
/*
1. send my tail to remote peers and then find the common ancestor
2. the remote peers will return the common ancestor and 10 blocks after the common ancestor if exist
3. compare the common ancestors, if over n+1 are the same, suppose the ancestor is the right ancestor
4. find overlapping blocks in 10 blocks who has the same ancestors
5. give the overlapping blocks to block pool one by one, if return false, go to next sync
6. if all remote peers return the number of blocks less than 10, end sync
*/
func (m *Manager) Start() {
	// if the node is syncing, return.
	if m.ns.Node().IsSynchronizing() {
		return
	}

	logging.CLog().Info("Starting Sync...")

	m.startMsgHandle()
	if len(m.ns.Node().Config().BootNodes) > 0 {
		m.ns.Node().SetSynchronizing(true)
		go m.startSync()
		m.curTail = m.blockChain.TailBlock()
	} else {
		logging.VLog().Info("Sync.Start: i am a seed node.")
		m.consensus.ContinueMining()
		go m.loop()
	}
}

func (m *Manager) startSync() {
	go m.loop()
	m.syncWithPeers(m.curTail)
}

func (m *Manager) loop() {

	for {
		select {
		case <-m.quitCh:
			return
		case <-m.endSyncCh:
			if m.ns.Node().IsSynchronizing() {
				m.ns.Node().SetSynchronizing(false)
			}
			m.consensus.ContinueMining()
			logging.VLog().Info("sync finish.")
		case <-m.syncCh:
			if m.curTail == nil {
				logging.VLog().Error("the current tail is nil.")
				m.curTail = m.blockChain.TailBlock()
			}
			logging.VLog().Info("sync continue")
			// sync continue
			m.syncWithPeers(m.curTail)
		}
	}
}

func (m *Manager) syncWithPeers(block *core.Block) {
	batch++
	tail := NewNetBlock(m.ns.Node().ID(), batch, block)
	logging.VLog().WithFields(logrus.Fields{
		"tail":  tail,
		"block": tail.block,
		"batch": batch,
	}).Info("sync with current tail")
	err := m.ns.Node().Sync(tail)

	switch err {
	case nil:
	case net.ErrPeersIsNotEnough:
		if m.ns.Node().IsSynchronizing() {
			logging.VLog().Info("sync target not enough, sleep for 30 second...")
			time.Sleep(30 * time.Second)
			m.syncCh <- true
			return
		}
	default:
		logging.VLog().Error("error occurs, sync has been terminated")
		panic("error occurs, sync has been terminated")
	}
	go (func() {
		timeout := 30 * time.Second

		select {
		case <-m.canSyncWithBlockListCh:
			m.syncWithBlockList(m.cacheList)
		case <-m.goParentSyncCh:
			logging.VLog().Info("sync with parent")
			m.goSyncParentWithPeers()
		case <-time.After(timeout):
			logging.VLog().Info("sync time out")
			m.syncWithPeers(m.curTail)
		}
	})()

}

func (m *Manager) goSyncParentWithPeers() {
	if m.ns.Node().IsSynchronizing() && !core.CheckGenesisBlock(m.curTail) {
		m.curTail = m.blockChain.GetBlock(m.curTail.ParentHash())
		m.syncWithPeers(m.curTail)
	} else {
		m.endSyncCh <- true
	}
}

// StartMsgHandle start sync message handle loop
func (m *Manager) startMsgHandle() {
	go (func() {
		for {
			select {
			case msg := <-m.receiveTailCh:
				if m.ns.Node().IsSynchronizing() {
					logging.VLog().Warn("Failed to reply sync message when synchronizing")
					continue
				}
				// 1.find the common ancestors
				// 2.find 10 blocks after ancestors if exist
				tail := new(NetBlock)
				pbblock := new(corepb.NetBlock)
				if err := proto.Unmarshal(msg.Data().([]byte), pbblock); err != nil {
					logging.VLog().WithFields(logrus.Fields{
						"err": err,
					}).Error("Failed to reply sync message")
					continue
				}
				if err := tail.FromProto(pbblock); err != nil {
					logging.VLog().WithFields(logrus.Fields{
						"err": err,
					}).Error("Failed to reply sync message")
					continue
				}

				key := m.ns.Node().ID()

				err := m.blockChain.CheckBlockOnCanonicalChain(tail.block)
				var emptyblocks []*core.Block
				if err != nil {
					logging.VLog().WithFields(logrus.Fields{
						"err": err,
					}).Error("Failed to reply sync message, failed to find common ancestor")
					netblocks := NewNetBlocks(key, tail.batch, emptyblocks)
					pb, _ := netblocks.ToProto()
					data, err := proto.Marshal(pb)
					if err != nil {
						continue
					}
					m.ns.SendMsg(net.SyncReply, data, tail.from, net.MessagePriorityHigh)
					continue
				}
				subsequentBlocks, err := m.blockChain.FetchDescendantInCanonicalChain(DescendantCount, tail.block)
				if err != nil {
					logging.VLog().WithFields(logrus.Fields{
						"err": err,
					}).Error("Failed to reply sync message, failed to fetch descendant in canonical chain")
					netblocks := NewNetBlocks(key, tail.batch, emptyblocks)
					pb, _ := netblocks.ToProto()
					data, err := proto.Marshal(pb)
					if err != nil {
						continue
					}
					m.ns.SendMsg(net.SyncReply, data, tail.from, net.MessagePriorityHigh)
					continue
				}
				subsequentBlocks = append(subsequentBlocks, tail.block)
				blocks := NewNetBlocks(key, tail.batch, subsequentBlocks)
				logging.VLog().WithFields(logrus.Fields{
					"from":   blocks.from,
					"batch":  blocks.batch,
					"blocks": blocks.blocks,
				}).Info("Send sync block response message")

				pb, _ := blocks.ToProto()
				data, err := proto.Marshal(pb)
				if err != nil {
					continue
				}
				m.ns.SendMsg(net.SyncReply, data, tail.from, net.MessagePriorityHigh)

			case msg := <-m.receiveSyncReplyCh:
				// 1. compare the common ancestors, if over n+1 are the same, suppose the ancestor is the right ancestor
				// 2. find overlapping blocks in 10 blocks who has the same ancestors
				// 3. give the overlapping blocks to block pool one by one, if return false, go to next sync.
				// 4. if all remote peers return the number of blocks less than 10, end sync
				data := new(NetBlocks)
				pbblocks := new(corepb.NetBlocks)
				if err := proto.Unmarshal(msg.Data().([]byte), pbblocks); err != nil {
					logging.VLog().WithFields(logrus.Fields{
						"err": err,
					}).Error("Failed to receive sync reply message")
					continue
				}
				if err := data.FromProto(pbblocks); err != nil {
					logging.VLog().WithFields(logrus.Fields{
						"err": err,
					}).Error("Failed to receive sync reply message")
					continue
				}

				blocks := data.Blocks()

				if data.batch < batch {
					logging.VLog().WithFields(logrus.Fields{
						"from":       data.from,
						"blocks":     data.Blocks(),
						"data.batch": data.batch,
						"batch":      batch,
					}).Info("batch is error")
					continue
				}

				limit := m.syncLimit()
				if len(blocks) == 0 {
					msgErrCount++
					logging.VLog().WithFields(logrus.Fields{
						"from":       data.from,
						"blocks":     blocks,
						"data.batch": data.batch,
						"batch":      batch,
					}).Info("Received sync reply message is wrong")

					if msgErrCount >= limit/2 {
						// go to next sync
						msgErrCount = 0
						m.goParentSyncCh <- true
					}
				}

				logging.VLog().WithFields(logrus.Fields{
					"from":       data.from,
					"blocks":     blocks,
					"data.batch": data.batch,
					"batch":      batch,
				}).Infof("Received sync reply message, %d/%d", len(m.cacheList), limit)

				if len(blocks) > 0 && len(m.cacheList) < limit {
					m.checkSyncLimitHandler(data)
				} else {
					continue
				}

			}
		}
	})()
}

func (m *Manager) syncLimit() int {
	return int(math.Sqrt(float64(m.ns.Node().PeersCount())))
}

func (m *Manager) checkSyncLimitHandler(data *NetBlocks) {
	m.cacheList[data.from] = data
	if len(m.cacheList) >= m.syncLimit() {
		m.canSyncWithBlockListCh <- true
	}

}

func (m *Manager) syncWithBlockList(list map[string]*NetBlocks) {
	// the map key is remote peer addrs
	// find the map key who have the common ancestor
	addrsArray := m.findBlocksWithCommonAncestor()

	// do sync blocks who have the common ancestor
	m.doSyncBlocksWithCommonAncestor(addrsArray)

}

func (m *Manager) doSyncBlocksWithCommonAncestor(addrsArray []string) {
	if len(addrsArray) == 0 {
		logging.VLog().Warn("Failed to find common ancestor")
		m.clearCacheList()
		m.syncCh <- true
		return
	}
	root := m.cacheList[addrsArray[0]].blocks
	var tail *core.Block
	for i := 0; i < len(root)-1; i++ {
		count := 1
		for j := 1; j < len(addrsArray); j++ {
			temp := m.cacheList[addrsArray[j]].blocks
			if len(temp)-1 > i {
				if root[i].Hash().String() == temp[i].Hash().String() {
					count++
				}
			}
		}
		// suppose root[i] is a legal block
		if count >= len(addrsArray) {
			if err := m.blockChain.BlockPool().Push(root[i]); err != nil {
				m.clearCacheList()
				logging.VLog().WithFields(logrus.Fields{
					"err": err,
				}).Error("Failed to push a block to pool")
				m.syncCh <- true
				return
			}
			tail = root[i]
		}
	}

	syncContinue := false
	for i := 0; i < len(addrsArray); i++ {
		if len(m.cacheList[addrsArray[i]].blocks) > DescendantCount {
			logging.VLog().Info("go to next synchronization")
			syncContinue = true
		}
	}

	if syncContinue {
		m.clearCacheList()
		m.curTail = tail
		m.syncCh <- true
	} else { // sync finish
		m.clearCacheList()
		m.endSyncCh <- true
	}
}

func (m *Manager) clearCacheList() {
	for k := range m.cacheList {
		delete(m.cacheList, k)
	}
}

// find blocks who have the common ancestor
func (m *Manager) findBlocksWithCommonAncestor() []string {
	tempList := make(map[string]int)
	ancestorList := make(map[string]string)
	for key, value := range m.cacheList {
		ancestor := value.blocks[len(value.blocks)-1]
		ancestorList[key] = ancestor.Hash().String()
		if _, ok := tempList[ancestor.Hash().String()]; ok {
			tempList[ancestor.Hash().String()] = tempList[ancestor.Hash().String()] + 1
		} else {
			tempList[ancestor.Hash().String()] = 1
		}
	}

	limitLen := m.syncLimit()/2 + 1
	var addrsArray []string

	for key, value := range tempList {
		if value >= limitLen {
			// Make sure the common ancestor is correct
			for addrs, hash := range ancestorList {
				if key == hash {
					addrsArray = append(addrsArray, addrs)
					if len(addrsArray) == limitLen {
						break
					}
				}
			}
			break
		}
	}
	return addrsArray
}

func (m *Manager) generateChunkMeta(syncpoint *core.Block) (*syncpb.ChunksMeta, error) {
	if err := m.blockChain.CheckBlockOnCanonicalChain(syncpoint); err != nil {
		return nil, err
	}
	tail := m.blockChain.TailBlock()
	if tail.Timestamp()-syncpoint.Timestamp() < core.DynastyInterval {
		logging.VLog().WithFields(logrus.Fields{
			"err": ErrTooSmallGapToSync,
		}).Warn("Failed to generate sync blocks meta info")
		return nil, ErrTooSmallGapToSync
	}

	chunks := [][]byte{}
	stor, err := storage.NewMemoryStorage()
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to create memory storage")
		return nil, err
	}
	chunksTrie, err := trie.NewBatchTrie(nil, stor)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to create merkle tree")
		return nil, err
	}

	curBlock := syncpoint
	target := int64(tail.Timestamp()/core.DynastyInterval) * core.DynastyInterval
	for curBlock.Timestamp() < target {
		hash := curBlock.Hash()
		height := curBlock.Height() + 1
		chunks = append(chunks, hash)
		chunksTrie.Put(hash, hash)
		curBlock = m.blockChain.GetBlockByHeight(height)
		if curBlock == nil {
			logging.VLog().WithFields(logrus.Fields{
				"err":    err,
				"height": height,
			}).Error("Failed to find the block on canonical chain.")
			return nil, ErrCannotFindBlockByHeight
		}
	}

	logging.VLog().WithFields(logrus.Fields{
		"syncpoint": syncpoint,
		"root":      chunksTrie.RootHash(),
		"chunks":    chunks,
	}).Debug("Succeed to generate chunks meta info.")
	return &syncpb.ChunksMeta{ChunksRoot: chunks, Root: chunksTrie.RootHash()}, nil
}

func (m *Manager) generateChunk(chunkRoot [][]byte) (*syncpb.Chunk, error) {
	blocks := []*corepb.Block{}
	for k, v := range chunkRoot {
		block := m.blockChain.GetBlock(v)
		if block == nil {
			logging.VLog().WithFields(logrus.Fields{
				"index": k,
				"hash":  byteutils.Hex(v),
				"err":   ErrCannotFindBlockByHash,
			}).Error("Failed to find the block on canonical chain.")
			return nil, ErrCannotFindBlockByHash
		}
		pbBlock, err := block.ToProto()
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"block": block,
				"err":   err,
			}).Error("Failed to serialize block.")
			return nil, err
		}
		blocks = append(blocks, pbBlock.(*corepb.Block))
	}

	logging.VLog().WithFields(logrus.Fields{
		"chunk": blocks,
	}).Debug("Succeed to generate chunk.")
	return &syncpb.Chunk{Blocks: blocks}, nil
}

func (m *Manager) processChunk(chunk *syncpb.Chunk) error {
	for k, v := range chunk.Blocks {
		block := new(core.Block)
		if err := block.FromProto(v); err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"index": k,
				"hash":  v.Header.Hash,
				"err":   err,
			}).Error("Failed to recover a block from proto data.")
			return err
		}
		if err := m.blockChain.BlockPool().Push(block); err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"index": k,
				"hash":  v.Header.Hash,
				"err":   err,
			}).Error("Failed to recover a block from proto data.")
			return err
		}
	}

	logging.VLog().WithFields(logrus.Fields{
		"chunk": chunk,
	}).Debug("Succeed to process chunk.")
	return nil
}
