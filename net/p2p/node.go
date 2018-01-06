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
	"github.com/libp2p/go-libp2p-swarm"
	"github.com/libp2p/go-libp2p/p2p/host/basic"
	"github.com/multiformats/go-multiaddr"
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
	ErrPortInUse           = errors.New("the port is already in use")
	ErrLoadKeypairFromFile = errors.New("failed to get Keypair from file")
	ErrNodeIsRunning       = errors.New("node is already running")
	ErrConnectToSeed       = errors.New("failed to say hello to seed")
)

// Node the node can be used as both the client and the server
type Node struct {
	ns        *NetService
	host      *basichost.BasicHost
	id        peer.ID
	peerstore peerstore.Peerstore
	// key: peer.ID: ip
	streamCache   *pdeque.PriorityDeque
	stream        *sync.Map
	routeTable    *kbucket.RoutingTable
	context       context.Context
	version       uint8
	config        *Config
	running       bool
	synchronizing bool
	syncList      []string
	// key: datachecksum value: peer.ID
	relayness      *lru.Cache
	bootIds        []string
	networkIDCache *lru.Cache
}

// NewNode start a local node and join the node to network
func NewNode(config *Config) (*Node, error) {

	node := &Node{}
	node.config = config
	node.context = context.Background()

	err := node.init()
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Error("node init failed.")
		return nil, err
	}
	logging.CLog().WithFields(logrus.Fields{
		"node.listen": node.config.Listen,
	}).Info("node init success")
	return node, nil
}

func (node *Node) init() error {

	ctx := node.context

	if err := node.checkPort(); err != nil {
		return err
	}
	if err := node.generatePeerStore(); err != nil {
		return err
	}

	node.routeTable = kbucket.NewRoutingTable(
		node.config.Bucketsize,
		kbucket.ConvertPeerID(node.id),
		node.config.Latency,
		node.peerstore,
	)

	node.routeTable.Update(node.id)

	node.stream = new(sync.Map)
	node.streamCache = pdeque.NewPriorityDeque(streamEliminationAlgorithm)
	node.version = node.config.Version

	var multiaddrs []multiaddr.Multiaddr
	for _, v := range node.config.Listen {
		tcpAddr, err := net.ResolveTCPAddr("tcp", v)
		if err != nil {
			return err
		}

		address, err := multiaddr.NewMultiaddr(
			fmt.Sprintf(
				"/ip4/%s/tcp/%d",
				tcpAddr.IP,
				tcpAddr.Port,
			),
		)
		multiaddrs = append(multiaddrs, address)
	}

	network, err := swarm.NewNetwork(
		ctx,
		multiaddrs,
		node.id,
		node.peerstore,
		nil,
	)
	node.relayness, err = lru.New(node.config.RelayCacheSize)
	node.networkIDCache, err = lru.New(node.config.StreamStoreSize)

	options := &basichost.HostOpts{}
	// add nat manager
	options.NATManager = basichost.NewNATManager(network)
	node.host, err = basichost.NewHost(node.context, network, options)
	return err
}

// Config return node config.
func (node *Node) Config() *Config {
	return node.config
}

// SetNs set netService
func (node *Node) SetNs(ns *NetService) {
	node.ns = ns
}

// ID return node ID.
func (node *Node) ID() string {
	return node.id.Pretty()
}

// PeerStore return node peerstore
func (node *Node) PeerStore() peerstore.Peerstore {
	return node.peerstore
}

// GetSynchronizing return node synchronizing
func (node *Node) GetSynchronizing() bool {
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

func (node *Node) checkPort() error {
	for _, v := range node.config.Listen {
		conn, err := net.Dial("tcp", v)
		if err == nil {
			conn.Close()
			return ErrPortInUse
		}
	}

	return nil
}

func (node *Node) generatePeerStore() error {

	path := node.Config().PrivateKeyPath
	priv, pub, err := getKeypairFromFile(path)
	if err != nil {
		priv, pub, err = GenerateEd25519Key()
		if err != nil {
			return err
		}
	}

	// Obtain Peer ID from public key
	node.id, err = peer.IDFromPublicKey(pub)
	if err != nil {
		return err
	}

	ps := peerstore.NewPeerstore()
	ps.AddPrivKey(node.id, priv)
	ps.AddPubKey(node.id, pub)
	node.peerstore = ps

	return nil
}

func (node *Node) setStreamHandler() *Node {
	// register streamHandler to start loop to handle stream origined from remote node.
	node.host.SetStreamHandler(ProtocolID, node.messageHandler)
	return node
}

func (node *Node) start() error {

	logging.CLog().WithFields(logrus.Fields{
		"id":    node.ID(),
		"addrs": node.host.Addrs(),
	}).Info("node start")

	if node.running {
		return ErrNodeIsRunning
	}
	node.running = true

	node.setStreamHandler()

	var success bool
	var wg sync.WaitGroup
	for _, bootNode := range node.config.BootNodes {
		wg.Add(1)
		go func(bootNode multiaddr.Multiaddr) {
			defer wg.Done()
			err := node.sayHelloToSeed(bootNode)
			if err != nil {
				logging.VLog().WithFields(logrus.Fields{
					"id":  bootNode,
					"err": err,
				}).Error("failed to say hello to seed")
			} else {
				logging.VLog().WithFields(logrus.Fields{
					"id": bootNode,
				}).Info("success to say hello to seed")
				success = true
			}

		}(bootNode)
	}
	wg.Wait()

	if success || len(node.Config().BootNodes) == 0 {
		go node.discovery(node.context)
		go node.manageStreamStore()
		logging.CLog().WithFields(logrus.Fields{
			"listen": node.config.Listen,
		}).Info("node start and join to p2p network success")
	} else {
		return ErrConnectToSeed
	}
	return nil
}
