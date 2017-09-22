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

package p2p

import (
	"github.com/nebulasio/go-nebulas/components/net"
	log "github.com/sirupsen/logrus"
)

// Manager is used to manage the p2p network
type Manager struct {
	config     *Config
	dispatcher *net.Dispatcher
	node       *Node
}

// NewManager create Manager instance.
func NewManager(config *Config) *Manager {
	if config == nil {
		config = DefautConfig()
	}
	n, err := NewNode(config)
	if err != nil {
		log.Error("NewP2pManager: node create fail...", err)
	}

	np := &Manager{
		config:     config,
		dispatcher: net.NewDispatcher(),
		node:       n,
	}
	np.RegisterBlockMsgService()
	return np
}

// Register register the subscribers.
func (np *Manager) Register(subscribers ...*net.Subscriber) {
	np.dispatcher.Register(subscribers...)
}

// Deregister Deregister the subscribers.
func (np *Manager) Deregister(subscribers ...*net.Subscriber) {
	np.dispatcher.Deregister(subscribers...)
}

// Start start p2p manager.
func (np *Manager) Start() {
	np.node.Start()
	np.dispatcher.Start()
}

// Stop stop p2p manager.
func (np *Manager) Stop() {
	np.dispatcher.Stop()
}

// PutMessage put message to dispatcher.
func (np *Manager) PutMessage(msg net.Message) {
	np.dispatcher.PutMessage(msg)
}

// BroadcastBlock broadcast block message
func (np *Manager) BroadcastBlock(block interface{}) {
	//TODO: broadcast block via underlying network lib to whole network.
	np.node.Broadcast(block)
}
