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
	"context"

	"errors"
	"fmt"
	"net"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

const letterBytes = "0123456789ABCDEF0123456789ABCDE10123456789ABCDEF0123456789ABCDEF"

const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// Error types
var (
	ErrPeerIsNotConnected = errors.New("peer is not connected")
)

// Node the node can be used as both the client and the server
type Node struct {
	synchronizing bool
	quitCh        chan bool
	netService    *NebService
	config        *Config
	context       context.Context
	id            peer.ID
	networkKey    crypto.PrivKey
	//network       inet.Network
	multiaddrs    []multiaddr.Multiaddr
	host          host.Host
	streamManager *StreamManager
	routeTable    *RouteTable
}

// NewNode return new Node according to the config.
func NewNode(config *Config) (*Node, error) {
	// check Listen port.
	if err := checkPortAvailable(config.Listen); err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"err":    err,
			"listen": config.Listen,
		}).Error("Failed to check port.")
		return nil, err
	}

	node := &Node{
		quitCh:        make(chan bool, 10),
		config:        config,
		context:       context.Background(),
		streamManager: NewStreamManager(config),
		synchronizing: false,
	}

	initP2PNetworkKey(config, node)
	initP2PRouteTable(config, node)

	if err := initP2PSwarmNetwork(config, node); err != nil {
		return nil, err
	}

	return node, nil
}

// Start host & route table discovery
func (node *Node) Start() error {
	logging.CLog().Info("Starting NebService Node...")

	node.streamManager.Start()

	if err := node.startHost(); err != nil {
		return err
	}

	node.routeTable.Start()

	logging.CLog().WithFields(logrus.Fields{
		"id":                node.ID(),
		"listening address": node.host.Addrs(),
	}).Info("Started NebService Node.")

	return nil
}

// Stop stop a node.
func (node *Node) Stop() {
	logging.CLog().WithFields(logrus.Fields{
		"id":                node.ID(),
		"listening address": node.host.Addrs(),
	}).Info("Stopping NebService Node...")

	node.routeTable.Stop()
	node.stopHost()
	node.streamManager.Stop()
}

func (node *Node) startHost() error {
	// add nat manager
	//options := &basichost.HostOpts{}
	//options.NATManager = basichost.NewNATManager
	//host, err := basichost.NewHost(node.context, node.swarm, options)
	opts := []libp2p.Option{
		libp2p.ListenAddrs(node.multiaddrs...),
		libp2p.Identity(node.networkKey),
		libp2p.Peerstore(node.routeTable.peerStore),
		libp2p.NATPortMap(),
	}
	host, err := libp2p.New(node.context, opts...)
	if err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"err":            err,
			"listen address": node.config.Listen,
		}).Error("Failed to start node.")
		return err
	}

	host.SetStreamHandler(NebProtocolID, node.onStreamConnected)
	//node.network = host.Network()
	node.host = host

	return nil
}

func (node *Node) stopHost() {
	//node.network.Close()

	if node.host == nil {
		return
	}

	node.host.Close()
}

// Config return node config.
func (node *Node) Config() *Config {
	return node.config
}

// SetNebService set netService
func (node *Node) SetNebService(ns *NebService) {
	node.netService = ns
}

// ID return node ID.
func (node *Node) ID() string {
	return node.id.Pretty()
}

// IsSynchronizing return node synchronizing
func (node *Node) IsSynchronizing() bool {
	return node.synchronizing
}

// SetSynchronizing set node synchronizing.
func (node *Node) SetSynchronizing(synchronizing bool) {
	node.synchronizing = synchronizing
}

// PeersCount return stream count.
func (node *Node) PeersCount() int32 {
	return node.streamManager.Count()
}

// RouteTable return route table.
func (node *Node) RouteTable() *RouteTable {
	return node.routeTable
}

func initP2PNetworkKey(config *Config, node *Node) {
	// init p2p network key.
	networkKey, err := LoadNetworkKeyFromFileOrCreateNew(config.PrivateKeyPath)
	if err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"err":        err,
			"NetworkKey": config.PrivateKeyPath,
		}).Warn("Failed to load network private key from file.")
	}

	peer.AdvancedEnableInlining = false

	node.networkKey = networkKey
	node.id, err = peer.IDFromPublicKey(networkKey.GetPublic())
	if err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"err":        err,
			"NetworkKey": config.PrivateKeyPath,
		}).Warn("Failed to generate ID from network key file.")
	}
}

func initP2PRouteTable(config *Config, node *Node) error {
	// init p2p route table.
	node.routeTable = NewRouteTable(config, node)
	return nil
}

func initP2PSwarmNetwork(config *Config, node *Node) error {
	// init p2p multiaddr and swarm network.
	multiaddrs := make([]multiaddr.Multiaddr, len(config.Listen))
	for idx, v := range node.config.Listen {
		tcpAddr, err := net.ResolveTCPAddr("tcp", v)
		if err != nil {
			logging.CLog().WithFields(logrus.Fields{
				"err":    err,
				"listen": v,
			}).Error("Failed to bind node socket.")
			return err
		}

		addr, err := multiaddr.NewMultiaddr(
			fmt.Sprintf(
				"/ip4/%s/tcp/%d",
				tcpAddr.IP,
				tcpAddr.Port,
			),
		)
		if err != nil {
			logging.CLog().WithFields(logrus.Fields{
				"err":    err,
				"listen": v,
			}).Error("Failed to bind node socket.")
			return err
		}

		multiaddrs[idx] = addr
	}
	node.multiaddrs = multiaddrs

	//swarm := swarm.NewSwarm(
	//	node.context,
	//	node.id,
	//	node.routeTable.peerStore,
	//	nil, // TODO: @robin integrate metrics.Reporter.
	//)
	//if err := swarm.Listen(multiaddrs...); err != nil {
	//	logging.CLog().WithFields(logrus.Fields{
	//		"err":    err,
	//		"addr": multiaddrs,
	//	}).Error("Failed to listen addr.")
	//	swarm.Close()
	//	return err
	//}

	//node.network = swarm
	return nil
}

func (node *Node) onStreamConnected(s network.Stream) {
	node.streamManager.Add(s, node)
}

// SendMessageToPeer send message to a peer.
func (node *Node) SendMessageToPeer(messageName string, data []byte, priority int, peerID string) error {
	stream := node.streamManager.FindByPeerID(peerID)
	if stream == nil {
		logging.VLog().WithFields(logrus.Fields{
			"pid": peerID,
			"err": ErrPeerIsNotConnected,
		}).Debug("Failed to locate peer's stream")
		return ErrPeerIsNotConnected
	}

	return stream.SendMessage(messageName, data, priority)
}

// BroadcastMessage broadcast message.
func (node *Node) BroadcastMessage(messageName string, data Serializable, priority int) {
	// node can not broadcast or relay message if it is in synchronizing.
	if node.synchronizing {
		return
	}

	node.streamManager.BroadcastMessage(messageName, data, priority)
}

// RelayMessage relay message.
func (node *Node) RelayMessage(messageName string, data Serializable, priority int) {
	// node can not broadcast or relay message if it is in synchronizing.
	if node.synchronizing {
		return
	}

	node.streamManager.RelayMessage(messageName, data, priority)
}
