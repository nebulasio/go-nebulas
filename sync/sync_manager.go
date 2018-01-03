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
	"time"

	pb "github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/consensus"
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/net"
	"github.com/nebulasio/go-nebulas/net/p2p"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

// const
const (
	DescendantCount = 3
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
	nm.Register(net.NewSubscriber(m, m.receiveTailCh, net.MessageTypeSyncBlock))
}

// RegisterSyncReplyInNetwork register message subscriber in network.
func (m *Manager) RegisterSyncReplyInNetwork(nm p2p.Manager) {
	nm.Register(net.NewSubscriber(m, m.receiveSyncReplyCh, net.MessageTypeSyncReply))
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
	if m.ns.Node().GetSynchronizing() {
		return
	}
	m.startMsgHandle()
	if len(m.ns.Node().Config().BootNodes) > 0 {
		m.ns.Node().SetSynchronizing(true)
		m.startSync()
		m.curTail = m.blockChain.TailBlock()
	} else {
		logging.VLog().Info("Sync.Start: i am a seed node.")
		m.consensus.SetCanMining(true)
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
			if m.ns.Node().GetSynchronizing() {
				m.ns.Node().SetSynchronizing(false)
			}
			m.consensus.SetCanMining(true)
			logging.VLog().Info("sync finish.")
		case <-m.syncCh:
			if m.curTail == nil {
				logging.VLog().Error("sync occurs error, the current tail is nil.")
				m.curTail = m.blockChain.TailBlock()
			}
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
	}).Info("syncWithPeers: got tail")
	err := m.ns.Sync(tail)

	switch err {
	case nil:
	case p2p.ErrNodeNotEnough:
		if m.ns.Node().GetSynchronizing() {
			logging.VLog().Info("syncWithPeers: sleep for 5 second...")
			time.Sleep(5 * time.Second)
			m.syncCh <- true
		}

	default:
		logging.VLog().Error("syncWithPeers occurs error, sync has been terminated.")
	}
	go (func() {
		timeout := 30 * time.Second

		select {
		case <-m.canSyncWithBlockListCh:
			m.syncWithBlockList(m.cacheList)
		case <-m.goParentSyncCh:
			m.goSyncParentWithPeers()
		case <-time.After(timeout):
			m.goSyncParentWithPeers()
		}
	})()

}

func (m *Manager) goSyncParentWithPeers() {
	if m.ns.Node().GetSynchronizing() && !core.CheckGenesisBlock(m.curTail) {
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
				if m.ns.Node().GetSynchronizing() {
					logging.VLog().Warn("node can not reply sync message when it is synchronizing")
					continue
				}
				// 1.find the common ancestors
				// 2.find 10 blocks after ancestors if exist
				tail := new(NetBlock)
				pbblock := new(corepb.NetBlock)
				if err := pb.Unmarshal(msg.Data().([]byte), pbblock); err != nil {
					logging.VLog().Error("StartMsgHandle.receiveTailCh: unmarshal data occurs error, ", err)
					continue
				}
				if err := tail.FromProto(pbblock); err != nil {
					logging.VLog().Error("StartMsgHandle.receiveTailCh: get block from proto occurs error: ", err)
					continue
				}

				key := m.ns.Node().ID()

				ancestor, err := m.blockChain.FindCommonAncestorWithTail(tail.block)
				var emptyblocks []*core.Block
				if err != nil {
					logging.VLog().Error("StartMsgHandle.receiveTailCh: find common ancestor with tail occurs error, ", err)
					netblocks := NewNetBlocks(key, tail.batch, emptyblocks)
					m.ns.SendSyncReply(tail.from, netblocks)
					continue
				}
				subsequentBlocks, err := m.blockChain.FetchDescendantInCanonicalChain(DescendantCount, ancestor)
				if err != nil {
					logging.VLog().Error("StartMsgHandle.receiveTailCh: FetchDescendantInCanonicalChain occurs error, ", err)
					netblocks := NewNetBlocks(key, tail.batch, emptyblocks)
					m.ns.SendSyncReply(tail.from, netblocks)
					continue
				}
				subsequentBlocks = append(subsequentBlocks, ancestor)
				blocks := NewNetBlocks(key, tail.batch, subsequentBlocks)
				logging.VLog().WithFields(logrus.Fields{
					"from":   blocks.from,
					"batch":  blocks.batch,
					"blocks": blocks.blocks,
				}).Info("StartMsgHandle.receiveTailCh: receive receiveTailCh message.")
				m.ns.SendSyncReply(tail.from, blocks)

			case msg := <-m.receiveSyncReplyCh:
				// 1. compare the common ancestors, if over n+1 are the same, suppose the ancestor is the right ancestor
				// 2. find overlapping blocks in 10 blocks who has the same ancestors
				// 3. give the overlapping blocks to block pool one by one, if return false, go to next sync.
				// 4. if all remote peers return the number of blocks less than 10, end sync
				data := new(NetBlocks)
				pbblocks := new(corepb.NetBlocks)
				if err := pb.Unmarshal(msg.Data().([]byte), pbblocks); err != nil {
					logging.VLog().Error("StartMsgHandle.receiveSyncReplyCh: unmarshal data occurs error, ", err)
					continue
				}
				if err := data.FromProto(pbblocks); err != nil {
					logging.VLog().Error("StartMsgHandle.receiveSyncReplyCh: get blocks from proto occurs error: ", err)
					continue
				}
				if data.batch < batch {
					continue
				}
				blocks := data.Blocks()

				if len(blocks) == 0 {
					msgErrCount++
					if msgErrCount >= p2p.LimitToSync/2 {
						// go to next sync
						msgErrCount = 0
						m.goParentSyncCh <- true
					}
				}

				logging.VLog().WithFields(logrus.Fields{
					"from":   data.from,
					"blocks": blocks,
				}).Info("StartMsgHandle.receiveSyncReplyCh: receive receiveSyncReplyCh message.")

				if len(blocks) > 0 && len(m.cacheList) < p2p.LimitToSync {
					m.checkSyncLimitHandler(data)
				} else {
					continue
				}

			}
		}
	})()
}

func (m *Manager) checkSyncLimitHandler(data *NetBlocks) {
	m.cacheList[data.from] = data
	if len(m.cacheList) >= p2p.LimitToSync {
		// m.syncWithBlockList(m.cacheList)
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
		logging.VLog().Warn("doSyncBlocksWithCommonAncestor: no common ancestor have been found")
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
				for k := range m.cacheList {
					delete(m.cacheList, k)
				}
				logging.VLog().Error("doSyncBlocksWithCommonAncestor: push a block to pool occrus error, ", err)
				m.syncCh <- true
				return
			}
			tail = root[i]
		}
	}

	syncContinue := false
	for i := 0; i < len(addrsArray); i++ {
		if len(m.cacheList[addrsArray[i]].blocks) > DescendantCount {
			logging.VLog().Info("StartMsgHandle: more Descendant need to synchronize, go to next synchronization")
			syncContinue = true
		}
	}

	if syncContinue {
		for k := range m.cacheList {
			delete(m.cacheList, k)
		}
		m.curTail = tail
		m.syncCh <- true
	} else { // sync finish
		for k := range m.cacheList {
			delete(m.cacheList, k)
		}
		m.endSyncCh <- true
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

	limitLen := p2p.LimitToSync/2 + 1
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
