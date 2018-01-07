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
	"github.com/nebulasio/go-nebulas/net"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

// NetService service for nebulas p2p network
type NetService struct {
	node       *Node
	quitCh     chan bool
	dispatcher *net.Dispatcher
}

// NewNetService create netService
func NewNetService(n Neblet) (*NetService, error) {
	config := NewP2PConfig(n)
	node, err := NewNode(config)
	if err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to create node")
		return nil, err
	}

	ns := &NetService{node, make(chan bool, 1), net.NewDispatcher()}
	node.SetNetService(ns)

	return ns, nil
}

// Node return the peer node
func (ns *NetService) Node() *Node {
	return ns.node
}

// Start start p2p manager.
func (ns *NetService) Start() error {
	ns.dispatcher.Start()
	if err := ns.node.Start(); err != nil {
		ns.dispatcher.Stop()
		return err
	}
	return nil
}

// Stop stop p2p manager.
func (ns *NetService) Stop() {
	ns.dispatcher.Stop()
	ns.quitCh <- true
}

// Register register the subscribers.
func (ns *NetService) Register(subscribers ...*net.Subscriber) {
	ns.dispatcher.Register(subscribers...)
}

// Deregister Deregister the subscribers.
func (ns *NetService) Deregister(subscribers ...*net.Subscriber) {
	ns.dispatcher.Deregister(subscribers...)
}

// PutMessage put message to dispatcher.
func (ns *NetService) PutMessage(msg net.Message) {
	ns.dispatcher.PutMessage(msg)
}

// Broadcast message.
func (ns *NetService) Broadcast(name string, msg net.Serializable) {
	ns.node.broadcast(name, msg)
}

// Relay message.
func (ns *NetService) Relay(name string, msg net.Serializable) {
	ns.node.relay(name, msg)
}

// BroadcastNetworkID broadcast networkID when changed.
func (ns *NetService) BroadcastNetworkID(msg []byte) {
	ns.node.broadcastNetworkID(msg)
}

// BuildData returns net service request data
func (ns *NetService) BuildData(data []byte, msgName string) []byte {
	return ns.node.buildData(data, msgName)
}

// SendMsg send message to a peer.
func (ns *NetService) SendMsg(msgName string, msg []byte, target string) error {
	return ns.node.sendMsg(msgName, msg, target)
}
