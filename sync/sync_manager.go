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
	"github.com/nebulasio/go-nebulas/components/net"
	"github.com/nebulasio/go-nebulas/components/net/p2p"
	"github.com/nebulasio/go-nebulas/consensus"
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/core/pb"
	log "github.com/sirupsen/logrus"
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
	quitPreSync        chan bool
	cacheList          map[string]*SyncBlocks
	cacheListChangeCh  chan bool
}

// NewManager new sync manager
func NewManager(blockChain *core.BlockChain, consensus consensus.Consensus, ns *p2p.NetService) *Manager {
	m := &Manager{blockChain, consensus, ns, make(chan bool), make(chan bool), make(chan net.Message, 128), make(chan net.Message, 128), make(chan bool), make(map[string]*SyncBlocks), make(chan bool)}
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
func (m *Manager) Start() {
	m.StartMsgHandle()
	if len(m.ns.Node().Config().BootNodes) > 0 {
		m.StartSync()
	}
	// TODO(@leon): start mining after consensus
	// m.consensus.Start()
}

// StartSync start sync loop
func (m *Manager) StartSync() {
	go (func() {
		m.syncWithTail()
	Loop:
		for {
			select {
			case <-m.quitCh:
				break Loop
			case <-m.syncCh:
				m.syncWithTail()
			}
		}
	})()
}

func (m *Manager) syncWithTail() {
	tail := m.blockChain.TailBlock()
	log.WithFields(log.Fields{
		"tail": tail,
	}).Info("syncWithTail: got tail")
	err := m.ns.Sync(tail)
	switch err {
	case nil:
	case p2p.ErrNodeNotEnough:
		time.Sleep(5 * time.Second)
		m.syncCh <- true
	}
}

// StartMsgHandle start sync message handle loop
func (m *Manager) StartMsgHandle() {
	go (func() {
		for {
			select {
			case msg := <-m.receiveTailCh:
				// 1.find the common ancestors
				// 2.find 10 blocks after ancestors if exist
				tail := new(core.Block)
				pbblock := new(corepb.Block)
				if err := pb.Unmarshal(msg.Data().([]byte), pbblock); err != nil {
					log.Error("StartMsgHandle.receiveTailCh: unmarshal data occurs error, ", err)
					continue
				}
				if err := tail.FromProto(pbblock); err != nil {
					log.Error("StartMsgHandle.receiveTailCh: get block from proto occurs error: ", err)
					continue
				}
				// tail := sync.blockChain.TailBlock()
				ancestor, err := m.blockChain.FindCommonAncestorWithTail(tail)
				if err != nil {
					log.Warn("StartMsgHandle.receiveTailCh: find common ancestor with tail occurs error, ", err)
					continue
				}
				subsequentBlocks, err := m.blockChain.FetchDescendantInCanonicalChain(10, ancestor)
				if err != nil {
					log.Warn("StartMsgHandle.receiveTailCh: FetchDescendantInCanonicalChain occurs error, ", err)
					continue
				}
				subsequentBlocks = append(subsequentBlocks, ancestor)
				addrs := m.ns.Addrs()
				blocks := NewSyncBlocks(addrs, subsequentBlocks)
				log.WithFields(log.Fields{
					"blocks": blocks,
				}).Info("StartMsgHandle.receiveTailCh: receive receiveTailCh message.")
				m.ns.SendSyncReply(blocks)

			case msg := <-m.receiveSyncReplyCh:
				// 1. compare the common ancestors, if over n+1 are the same, suppose the ancestor is the right ancestor
				// 2. find overlapping blocks in 10 blocks who has the same ancestors
				// 3. give the overlapping blocks to block pool one by one, if return false, go to next sync.
				// 4. if all remote peers return the number of blocks less than 10, end sync
				data := new(SyncBlocks)
				pbblocks := new(corepb.SyncBlocks)
				if err := pb.Unmarshal(msg.Data().([]byte), pbblocks); err != nil {
					log.Error("StartMsgHandle.receiveSyncReplyCh: unmarshal data occurs error, ", err)
					continue
				}
				if err := data.FromProto(pbblocks); err != nil {
					log.Error("StartMsgHandle.receiveSyncReplyCh: get blocks from proto occurs error: ", err)
					continue
				}
				blocks := data.Blocks()
				if len(blocks) > 0 && len(m.cacheList) < p2p.LimitToSync {
					m.cacheList[data.addrs] = data
				} else {
					continue
				}

			case <-m.cacheListChangeCh:
				if len(m.cacheList) >= p2p.LimitToSync {
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
					addrsArray := make([]string, limitLen)

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
					root := m.cacheList[addrsArray[0]].blocks
					for i := 0; i < len(root)-1; i++ {
						count := 0
						for j := 1; j < len(addrsArray); j++ {
							temp := m.cacheList[addrsArray[j]].blocks
							if root[i].Hash().String() == temp[i].Hash().String() {
								count++
							}
						}
						// suppose root[i] is a legal block
						if count > limitLen {
							if err := m.blockChain.BlockPool().Push(root[i]); err != nil {
								for k := range m.cacheList {
									delete(m.cacheList, k)
								}

								m.syncCh <- true
								return
							}
						}
					}

					syncContinue := false
					for i := 0; i < len(addrsArray); i++ {
						if len(m.cacheList[addrsArray[0]].blocks) > 10 {
							syncContinue = true
						}
					}

					if syncContinue {
						for k := range m.cacheList {
							delete(m.cacheList, k)
						}
						m.syncCh <- true
					} else { // sync finish
						m.quitCh <- true
					}

				}
			}
		}
	})()
}
