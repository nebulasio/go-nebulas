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
	log "github.com/sirupsen/logrus"
)

// const
const (
	DescendantCount = 3
)

// Manager is used to manage the sync service
type Manager struct {
	blockChain         *core.BlockChain
	consensus          consensus.Consensus
	ns                 *p2p.NetService
	quitCh             chan bool
	syncCh             chan bool
	receiveTailCh      chan net.Message
	receiveSyncReplyCh chan net.Message
	cacheList          map[string]*NetBlocks
	endSyncCh          chan bool
	curTail            *core.Block
	canHandleCh        chan bool
	errCount           int
	nonce              uint64
	goParentSyncCh     chan bool
}

// NewManager new sync manager
func NewManager(blockChain *core.BlockChain, consensus consensus.Consensus, ns *p2p.NetService) *Manager {
	m := &Manager{blockChain, consensus, ns, make(chan bool, 1), make(chan bool, 1), make(chan net.Message, 128), make(chan net.Message, 128), make(map[string]*NetBlocks), make(chan bool, 1), blockChain.TailBlock(), make(chan bool, 1), 0, 0, make(chan bool, 1)}
	m.RegisterSyncBlockInNetwork(ns)
	m.RegisterSyncReplyInNetwork(ns)
	return m
}

// RegisterSyncBlockInNetwork register message subscriber in network.
func (m *Manager) RegisterSyncBlockInNetwork(nm net.Manager) {
	nm.Register(net.NewSubscriber(m, m.receiveTailCh, net.MessageTypeSyncBlock))
}

// RegisterSyncReplyInNetwork register message subscriber in network.
func (m *Manager) RegisterSyncReplyInNetwork(nm net.Manager) {
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
	m.startMsgHandle()
	if len(m.ns.Node().Config().BootNodes) > 0 {
		m.startSync()
		m.curTail = m.blockChain.TailBlock()
	} else {
		log.Info("Sync.Start: i am a seed node.")
		m.ns.Node().SetSynchronized(true)
		m.consensus.SetCanMining(true)
		go m.loop()
		go m.downloader()
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
			if !m.ns.Node().GetSynchronized() {
				m.ns.Node().SetSynchronized(true)
				m.consensus.SetCanMining(true)
				m.downloader()
			}
		case <-m.syncCh:
			m.syncWithPeers(m.curTail)
		}
	}
}

func (m *Manager) downloader() {
	second := 10 * time.Second
	ticker := time.NewTicker(second)
	for {
		select {
		case <-ticker.C:
			block := m.blockChain.TailBlock()
			m.syncWithPeers(block)
		}
	}
}

func (m *Manager) syncWithPeers(block *core.Block) {
	// block := m.blockChain.TailBlock()
	m.nonce++
	//key, err := p2p.GenerateKey(m.ns.Addrs(), m.ns.Node().ID())
	//if err != nil {
	//	log.Error("GenerateKey occurs error, sync has been terminated.")
	//	return
	//}
	tail := NewNetBlock(m.ns.Node().ID(), m.nonce, block)
	log.WithFields(log.Fields{
		"tail":  tail,
		"block": tail.block,
	}).Info("syncWithPeers: got tail")
	err := m.ns.Sync(tail)

	switch err {
	case nil:
	case p2p.ErrNodeNotEnough:
		if !m.ns.Node().GetSynchronized() {
			log.Warn("syncWithPeers: sleep for 5 second...")
			time.Sleep(5 * time.Second)
			m.syncCh <- true
		}

	default:
		log.Error("syncWithPeers occurs error, sync has been terminated.")
	}
	go (func() {
		timeout := 30 * time.Second

		select {
		case <-m.canHandleCh:
			m.syncWithBlockList(m.cacheList)
		case <-m.goParentSyncCh:
			m.goSyncParentWithPeers()
		case <-time.After(timeout):
			m.goSyncParentWithPeers()
		}
	})()

}

func (m *Manager) goSyncParentWithPeers() {
	if !m.ns.Node().GetSynchronized() && !core.CheckGenesisBlock(m.curTail) {
		m.curTail = m.blockChain.GetBlock(m.curTail.ParentHash())
		m.syncWithPeers(m.curTail)
	}
}

// StartMsgHandle start sync message handle loop
func (m *Manager) startMsgHandle() {
	go (func() {
		for {
			select {
			case msg := <-m.receiveTailCh:
				if !m.ns.Node().GetSynchronized() {
					log.Warn("node can not reply sync message when it is synchronizing")
					continue
				}
				// 1.find the common ancestors
				// 2.find 10 blocks after ancestors if exist
				tail := new(NetBlock)
				pbblock := new(corepb.NetBlock)
				if err := pb.Unmarshal(msg.Data().([]byte), pbblock); err != nil {
					log.Error("StartMsgHandle.receiveTailCh: unmarshal data occurs error, ", err)
					continue
				}
				if err := tail.FromProto(pbblock); err != nil {
					log.Error("StartMsgHandle.receiveTailCh: get block from proto occurs error: ", err)
					continue
				}

				//key, err := p2p.GenerateKey(m.ns.Addrs(), m.ns.Node().ID())
				//if err != nil {
				//	log.Warn("StartMsgHandle.receiveTailCh: GenerateKey occurs error, ", err)
				//	continue
				//}
				key := m.ns.Node().ID()

				ancestor, err := m.blockChain.FindCommonAncestorWithTail(tail.block)
				var emptyblocks []*core.Block
				if err != nil {
					log.Warn("StartMsgHandle.receiveTailCh: find common ancestor with tail occurs error, ", err)
					netblocks := NewNetBlocks(key, tail.nonce, emptyblocks)
					m.ns.SendSyncReply(tail.from, netblocks)
					continue
				}
				subsequentBlocks, err := m.blockChain.FetchDescendantInCanonicalChain(DescendantCount, ancestor)
				if err != nil {
					log.Warn("StartMsgHandle.receiveTailCh: FetchDescendantInCanonicalChain occurs error, ", err)
					netblocks := NewNetBlocks(key, tail.nonce, emptyblocks)
					m.ns.SendSyncReply(tail.from, netblocks)
					continue
				}
				subsequentBlocks = append(subsequentBlocks, ancestor)
				blocks := NewNetBlocks(key, tail.nonce, subsequentBlocks)
				log.WithFields(log.Fields{
					"from":   blocks.from,
					"nonce":  blocks.nonce,
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
					log.Error("StartMsgHandle.receiveSyncReplyCh: unmarshal data occurs error, ", err)
					continue
				}
				if err := data.FromProto(pbblocks); err != nil {
					log.Error("StartMsgHandle.receiveSyncReplyCh: get blocks from proto occurs error: ", err)
					continue
				}
				if data.nonce < m.nonce {
					continue
				}
				blocks := data.Blocks()

				if len(blocks) == 0 {
					m.errCount++
					if m.errCount >= p2p.LimitToSync/2 {
						// go to next sync
						m.goParentSyncCh <- true
					}
				}

				log.WithFields(log.Fields{
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
		m.canHandleCh <- true
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
		log.Warn("doSyncBlocksWithCommonAncestor: no common ancestor have been found")
		m.syncCh <- true
		return
	}
	root := m.cacheList[addrsArray[0]].blocks
	var tail *core.Block
	for i := 0; i < len(root)-1; i++ {
		count := 1
		for j := 1; j < len(addrsArray); j++ {
			temp := m.cacheList[addrsArray[j]].blocks
			if root[i].Hash().String() == temp[i].Hash().String() {
				count++
			}
		}
		// suppose root[i] is a legal block
		if count >= len(addrsArray) {
			if err := m.blockChain.BlockPool().Push(root[i]); err != nil {
				for k := range m.cacheList {
					delete(m.cacheList, k)
				}
				log.Warn("doSyncBlocksWithCommonAncestor: push a block to pool occrus error, ", err)
				m.syncCh <- true
				return
			}
			tail = root[i]
		}
	}

	syncContinue := false
	for i := 0; i < len(addrsArray); i++ {
		if len(m.cacheList[addrsArray[i]].blocks) > DescendantCount {
			log.Info("StartMsgHandle: more Descendant need to synchronize, go to next synchronization")
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
