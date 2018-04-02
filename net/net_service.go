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

package net

import (
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

// NebService service for nebulas p2p network
type NebService struct {
	node       *Node
	dispatcher *Dispatcher
}

// NewNebService create netService
func NewNebService(n Neblet) (*NebService, error) {
	if networkConf := n.Config().GetNetwork(); networkConf == nil {
		logging.CLog().Fatal("Failed to find network config in config file")
		return nil, ErrConfigLackNetWork
	}
	node, err := NewNode(NewP2PConfig(n))
	if err != nil {
		return nil, err
	}

	ns := &NebService{
		node:       node,
		dispatcher: NewDispatcher(),
	}
	node.SetNebService(ns)

	return ns, nil
}

// Node return the peer node
func (ns *NebService) Node() *Node {
	return ns.node
}

// Start start p2p manager.
func (ns *NebService) Start() error {
	logging.CLog().Info("Starting NebService...")

	// start dispatcher.
	ns.dispatcher.Start()

	// start node.
	if err := ns.node.Start(); err != nil {
		ns.dispatcher.Stop()
		logging.CLog().WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to start NebService.")
		return err
	}

	logging.CLog().Info("Started NebService.")
	return nil
}

// Stop stop p2p manager.
func (ns *NebService) Stop() {
	logging.CLog().Info("Stopping NebService...")

	ns.node.Stop()
	ns.dispatcher.Stop()
}

// Register register the subscribers.
func (ns *NebService) Register(subscribers ...*Subscriber) {
	ns.dispatcher.Register(subscribers...)
}

// Deregister Deregister the subscribers.
func (ns *NebService) Deregister(subscribers ...*Subscriber) {
	ns.dispatcher.Deregister(subscribers...)
}

// PutMessage put message to dispatcher.
func (ns *NebService) PutMessage(msg Message) {
	ns.dispatcher.PutMessage(msg)
}

// Broadcast message.
func (ns *NebService) Broadcast(name string, msg Serializable, priority int) {
	ns.node.BroadcastMessage(name, msg, priority)
}

// Relay message.
func (ns *NebService) Relay(name string, msg Serializable, priority int) {
	ns.node.RelayMessage(name, msg, priority)
}

// BroadcastNetworkID broadcast networkID when changed.
func (ns *NebService) BroadcastNetworkID(msg []byte) {
	// TODO: @robin networkID.
}

// BuildRawMessageData return the raw NebMessage content data.
func (ns *NebService) BuildRawMessageData(data []byte, msgName string) []byte {
	message, err := NewNebMessage(ns.node.config.ChainID, DefaultReserved, 0, msgName, data)
	if err != nil {
		return nil
	}

	return message.Content()
}

// SendMsg send message to a peer.
func (ns *NebService) SendMsg(msgName string, msg []byte, target string, priority int) error {
	return ns.node.SendMessageToPeer(msgName, msg, priority, target)
}

// SendMessageToPeers send message to peers.
func (ns *NebService) SendMessageToPeers(messageName string, data []byte, priority int, filter PeerFilterAlgorithm) []string {
	return ns.node.streamManager.SendMessageToPeers(messageName, data, priority, filter)
}

// SendMessageToPeer send message to a peer.
func (ns *NebService) SendMessageToPeer(messageName string, data []byte, priority int, peerID string) error {
	return ns.node.SendMessageToPeer(messageName, data, priority, peerID)
}

// ClosePeer close the stream to a peer.
func (ns *NebService) ClosePeer(peerID string, reason error) {
	ns.node.streamManager.CloseStream(peerID, reason)
}
