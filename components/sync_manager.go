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

package components

import (
	"github.com/nebulasio/go-nebulas/components/net"
	"github.com/nebulasio/go-nebulas/components/net/p2p"
	"github.com/nebulasio/go-nebulas/core"
)

// SyncManager is used to manage the sync service
type SyncManager struct {
	blockChain         *core.BlockChain
	ns                 *p2p.NetService
	quitCh             chan bool
	syncCh             chan bool
	receiveTailCh      chan net.Message
	receiveSyncReplyCh chan net.Message
	quitPreSync        chan bool
}

// NewSyncManager new sync manager
func NewSyncManager(blockChain *core.BlockChain, ns *p2p.NetService) *SyncManager {
	sync := &SyncManager{blockChain, ns, make(chan bool), make(chan bool), make(chan net.Message, 128), make(chan net.Message, 128), make(chan bool)}
	sync.RegisterSyncBlockInNetwork(ns)
	sync.RegisterSyncReplyInNetwork(ns)
	return sync
}

// RegisterSyncBlockInNetwork register message subscriber in network.
func (sync *SyncManager) RegisterSyncBlockInNetwork(nm net.Manager) {
	nm.Register(net.NewSubscriber(sync, sync.receiveTailCh, net.MessageTypeSyncBlock))
}

// RegisterSyncReplyInNetwork register message subscriber in network.
func (sync *SyncManager) RegisterSyncReplyInNetwork(nm net.Manager) {
	nm.Register(net.NewSubscriber(sync, sync.receiveSyncReplyCh, net.MessageTypeSyncReply))
}

// Start start sync service
func (sync *SyncManager) Start() {
	sync.StartMsgHandle()
	sync.StartSync()
}

// StartSync start sync loop
func (sync *SyncManager) StartSync() {
	go (func() {
		tail := sync.blockChain.TailBlock()
		sync.PreSync(tail)
	Loop:
		for {
			select {
			case <-sync.quitCh:
				break Loop
			case <-sync.syncCh:
				go sync.sync()
				// case <-sync.receiveMessageCh:

			}
		}
	})()
}

// StartMsgHandle start sync message handle loop
func (sync *SyncManager) StartMsgHandle() {
	go (func() {
		for {
			select {
			case <-sync.receiveTailCh:
				// 1.find the common ancestors
				// 2.find 10 blocks after ancestors if exist

			case <-sync.receiveSyncReplyCh:
			}
		}
	})()
}

// PreSync be ready before start sync
func (sync *SyncManager) PreSync(tail *core.Block) {
	go (func() {
		ns := sync.ns
		for {
			select {
			case <-sync.quitPreSync:
				return
			default:
				ns.PreSync(tail)
			}
		}
	})()

}

func (sync *SyncManager) sync() {

}
