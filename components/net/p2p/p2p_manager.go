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
	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/components/net"
)

// Manager is used to manage the p2p network
type Manager struct {
	config *Config
	// dispatcher *net.Dispatcher
	// node       *Node
	netService *NetService
}

// NewManager create Manager instance.
func NewManager(config *Config) *Manager {
	// if config == nil {
	// 	config = DefautConfig()
	// }
	// n, err := NewNode(config)
	// if err != nil {
	// 	log.Error("NewP2pManager: node create fail...", err)
	// }
	netService := NewNetService(config)

	np := &Manager{
		config: config,
		// dispatcher: net.NewDispatcher(),
		// node:       n,
		netService: netService,
	}
	return np
}

// Register register the subscribers.
func (np *Manager) Register(subscribers ...*net.Subscriber) {
	np.netService.Register(subscribers...)
}

// Deregister Deregister the subscribers.
func (np *Manager) Deregister(subscribers ...*net.Subscriber) {
	np.netService.Deregister(subscribers...)
}

// Start start p2p manager.
func (np *Manager) Start() {
	np.netService.Start()
	// np.dispatcher.Start()
}

// Stop stop p2p manager.
func (np *Manager) Stop() {
	np.netService.Stop()
}

// PutMessage put message to dispatcher.
func (np *Manager) PutMessage(msg net.Message) {
	np.netService.PutMessage(msg)
}

// Broadcast message
func (np *Manager) Broadcast(name string, msg proto.Message) {
	//TODO: broadcast block via underlying network lib to whole network.
	np.netService.Broadcast(name, msg)
}

// Relay message
func (np *Manager) Relay(name string, msg proto.Message) {
	//TODO: broadcast block via underlying network lib to whole network.
	np.Broadcast(name, msg)
}
