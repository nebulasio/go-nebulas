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
	"context"
	"errors"
	"fmt"
	"net"
	"sync"

	lru "github.com/hashicorp/golang-lru"
	"github.com/libp2p/go-libp2p-kbucket"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-peerstore"
	swarm "github.com/libp2p/go-libp2p-swarm"
	"github.com/libp2p/go-libp2p/p2p/host/basic"
	multiaddr "github.com/multiformats/go-multiaddr"
	"github.com/nebulasio/go-nebulas/common/pdeque"
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
	ErrLoadKeypairFromFile = errors.New("failed to get Keypair from file")
	ErrNodeIsRunning       = errors.New("node is already running")
	ErrConnectToSeed       = errors.New("failed to say hello to seed")
)

// Node the node can be used as both the client and the server
type Node struct {
	quitCh         chan bool
	netService     *NetService
	host           *basichost.BasicHost
	id             peer.ID
	peerstore      peerstore.Peerstore
	streamCache    *pdeque.PriorityDeque
	stream         *sync.Map
	routeTable     *kbucket.RoutingTable
	context        context.Context
	config         *Config
	running        bool
	synchronizing  bool
	relayness      *lru.Cache
	bootIds        []string
	networkIDCache *lru.Cache
	network        *swarm.Network
}

// NewNode return new Node according to the config.
func NewNode(config *Config) (*Node, error) {
	// check Listen port.
	if err := checkPortAvailable(config.Listen); err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"err":    err,
			"listen": config.Listen,
		}).Error("Listen port is not available.")
		return nil, err
	}

	node := &Node{
		quitCh:        make(chan bool, 10),
		config:        config,
		context:       context.Background(),
		stream:        new(sync.Map),
		streamCache:   pdeque.NewPriorityDeque(streamEliminationAlgorithm),
		peerstore:     peerstore.NewPeerstore(),
		synchronizing: false,
		running:       false,
	}
	node.relayness, _ = lru.New(config.RelayCacheSize)
	node.networkIDCache, _ = lru.New(config.StreamStoreSize)

	initP2PNetworkKey(config, node)
	initP2PRouteTable(config, node)
	initP2PSwarmNetwork(config, node)

	return node, nil
}

// Start host & route table discovery
func (node *Node) Start() error {
	if err := node.startHost(); err != nil {
		return err
	}

	go node.discovery(node.context)
	go node.manageStreamStore()

	logging.CLog().WithFields(logrus.Fields{
		"id":                node.ID(),
		"listening address": node.host.Addrs(),
	}).Info("Succeed to start node.")

	return nil
}

func (node *Node) startHost() error {
	// add nat manager
	options := &basichost.HostOpts{}
	options.NATManager = basichost.NewNATManager(node.network)
	host, err := basichost.NewHost(node.context, node.network, options)
	if err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"err":            err,
			"listen address": node.config.Listen,
		}).Error("Failed to start node.")
		return err
	}
	host.SetStreamHandler(ProtocolID, node.messageHandler)

	node.host = host
	return nil
}

// Config return node config.
func (node *Node) Config() *Config {
	return node.config
}

// SetNetService set netService
func (node *Node) SetNetService(ns *NetService) {
	node.netService = ns
}

// ID return node ID.
func (node *Node) ID() string {
	return node.id.Pretty()
}

// PeerStore return node peerstore
func (node *Node) PeerStore() peerstore.Peerstore {
	return node.peerstore
}

// IsSynchronizing return node synchronizing
func (node *Node) IsSynchronizing() bool {
	return node.synchronizing
}

// SetSynchronizing set node synchronizing.
func (node *Node) SetSynchronizing(synchronizing bool) {
	node.synchronizing = synchronizing
}

// GetStream return node stream.
func (node *Node) GetStream() *sync.Map {
	return node.stream
}

func initP2PNetworkKey(config *Config, node *Node) error {
	// init p2p network key.
	networkKey, err := LoadNetworkKeyFromFileOrCreateNew(config.PrivateKeyPath)
	if err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"err":        err,
			"NetworkKey": config.PrivateKeyPath,
		}).Error("Failed to load network private key from file.")
		return err
	}

	node.id, err = peer.IDFromPublicKey(networkKey.GetPublic())
	if err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"err":        err,
			"NetworkKey": config.PrivateKeyPath,
		}).Error("Failed to generate ID from network key file.")
		return err
	}

	node.peerstore.AddPubKey(node.id, networkKey.GetPublic())
	node.peerstore.AddPrivKey(node.id, networkKey)

	return nil
}

func initP2PRouteTable(config *Config, node *Node) error {
	// init p2p route table.
	node.routeTable = kbucket.NewRoutingTable(
		node.config.Bucketsize,
		kbucket.ConvertPeerID(node.id),
		node.config.Latency,
		node.peerstore,
	)
	node.routeTable.Update(node.id)

	return nil
}

func initP2PSwarmNetwork(config *Config, node *Node) error {
	// init p2p multiaddr and swarm network.
	multiaddrs := make([]multiaddr.Multiaddr, len(config.Listen))
	for idx, v := range node.config.Listen {
		tcpAddr, err := net.ResolveTCPAddr("tcp", v)
		if err != nil {
			logging.CLog().WithFields(logrus.Fields{
				"err":            err,
				"listen address": v,
			}).Error("Invalid listen address.")
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
				"err":            err,
				"listen address": v,
			}).Error("Invalid listen address.")
			return err
		}

		multiaddrs[idx] = addr
	}

	network, err := swarm.NewNetwork(
		node.context,
		multiaddrs,
		node.id,
		node.peerstore,
		nil, // TODO: @robin integrate metrics.Reporter.
	)
	if err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"err":            err,
			"listen address": config.Listen,
			"node.id":        node.id.Pretty(),
		}).Error("Failed to create swarm network.")
		return err
	}
	node.network = network
	return nil
}
