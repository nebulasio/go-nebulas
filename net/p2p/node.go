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
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	mrand "math/rand"
	"net"
	"strings"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-kbucket"
	libnet "github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-peerstore"
	"github.com/libp2p/go-libp2p-swarm"
	"github.com/libp2p/go-libp2p/p2p/host/basic"
	"github.com/multiformats/go-multiaddr"
	"github.com/nebulasio/go-nebulas/common/pdeque"
	log "github.com/sirupsen/logrus"
)

const letterBytes = "0123456789ABCDEF0123456789ABCDE10123456789ABCDEF0123456789ABCDEF"

const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// Node the node can be used as both the client and the server
type Node struct {
	host      *basichost.BasicHost
	id        peer.ID
	peerstore peerstore.Peerstore
	// key: peer.ID: ip
	streamCache  *pdeque.PriorityDeque
	stream       map[string]*StreamStore
	routeTable   *kbucket.RoutingTable
	context      context.Context
	chainID      uint32
	version      uint8
	config       *Config
	running      bool
	synchronized bool
	syncList     []string
	// key: datachecksum value: peer.ID
	relayness     *lru.Cache
	relaynessLock *sync.Mutex
}

// StreamStore is for stream cache
type StreamStore struct {
	key       string
	conn      int
	stream    libnet.Stream
	timestamp int64
}

func less(a interface{}, b interface{}) bool {
	sa := a.(*StreamStore)
	sb := b.(*StreamStore)
	return sa.timestamp < sb.timestamp
}

// NewStreamStore return a new streamStore
func NewStreamStore(key string, conn int, stream libnet.Stream) *StreamStore {
	return &StreamStore{key, conn, stream, time.Now().Unix()}
}

// NewNode start a local node and join the node to network
func NewNode(config *Config) (*Node, error) {

	node := &Node{}
	node.config = config
	node.context = context.Background()

	err := node.init()
	if err != nil {
		log.Error("start node fail, can not init node", err)
		return nil, err
	}
	log.WithFields(log.Fields{
		"node.id":   node.ID(),
		"node.port": node.config.Port,
	}).Debug("node init success")
	return node, nil
}

// Config return node config.
func (node *Node) Config() *Config {
	return node.config
}

// ID return node ID.
func (node *Node) ID() string {
	return node.id.Pretty()
}

// PeerStore return node peerstore
func (node *Node) PeerStore() peerstore.Peerstore {
	return node.peerstore
}

// SetSynchronized set node synchronized.
func (node *Node) SetSynchronized(synchronized bool) {
	node.synchronized = synchronized
}

// GetSynchronized return node synchronized status.
func (node *Node) GetSynchronized() bool {
	return node.synchronized
}

// GetStream return node stream.
func (node *Node) GetStream() map[string]*StreamStore {
	return node.stream
}

func (node *Node) checkPort() error {
	conn, err := net.Dial("tcp",
		fmt.Sprintf(
			"%s:%d",
			node.config.IP,
			node.config.Port,
		),
	)
	if err == nil {
		conn.Close()
		return errors.New("The port already in use")
	}
	return nil
}

func (node *Node) generatePeerStore() error {
	var randseedstr string
	if len(node.Config().BootNodes) == 0 {
		// seednode
		randseedstr = letterBytes
	} else {
		randseedstr = randSeed(64)
	}
	randseed, err := hex.DecodeString(randseedstr)
	priv, pub, err := crypto.GenerateEd25519Key(
		bytes.NewReader(randseed),
	)
	if err != nil {
		return err
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
	node.stream = make(map[string]*StreamStore)
	node.streamCache = pdeque.NewPriorityDeque(less)
	node.chainID = node.config.ChainID
	node.version = node.config.Version
	node.synchronized = false
	node.relaynessLock = &sync.Mutex{}
	address, err := multiaddr.NewMultiaddr(
		fmt.Sprintf(
			"/ip4/%s/tcp/%d",
			node.config.IP,
			node.config.Port,
		),
	)
	network, err := swarm.NewNetwork(
		ctx,
		[]multiaddr.Multiaddr{address},
		node.id,
		node.peerstore,
		nil,
	)
	node.relayness, err = lru.New(node.config.RelayCacheSize)

	options := &basichost.HostOpts{}
	// add nat manager
	options.NATManager = basichost.NewNATManager(network)
	node.host, err = basichost.NewHost(node.context, network, options)
	return err
}

func randSeed(n int) string {
	var src = mrand.NewSource(time.Now().UnixNano())
	b := make([]byte, n)
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = mrand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

// SayHello Say hello to trustedNode
func (netService *NetService) SayHello(bootNode multiaddr.Multiaddr) error {
	node := netService.node
	bootAddr, bootID, err := parseAddressFromMultiaddr(bootNode)
	if err != nil {
		log.WithFields(log.Fields{
			"bootNode": bootNode,
			"error":    err,
		}).Error("parse Address from trustedNode failed")
		return err
	}
	node.peerstore.AddAddr(
		bootID,
		bootAddr,
		peerstore.TempAddrTTL,
	)
	if node.host.Addrs()[0].String() != bootAddr.String() {
		var success = false
		for i := 0; i < 3; i++ {
			err := netService.Hello(bootID)
			if err != nil {
				time.Sleep(time.Second)
				continue
			}
			success = true
			break
		}
		if !success {
			log.WithFields(log.Fields{
				"bootNode": bootNode,
				"error":    err,
			}).Error("say hello to bootNode failed")
			return errors.New("say hello to bootNode failed")
		}
		log.WithFields(log.Fields{
			"bootNode": bootNode,
		}).Debug("say hello to a node success")
		node.peerstore.AddAddr(
			bootID,
			bootAddr,
			peerstore.PermanentAddrTTL)
		// Update the routing table.
		node.routeTable.Update(bootID)
	}
	return nil
}

func parseAddressFromMultiaddr(address multiaddr.Multiaddr) (multiaddr.Multiaddr, peer.ID, error) {

	addr, err := multiaddr.NewMultiaddr(
		strings.Split(address.String(), "/ipfs/")[0],
	)
	if err != nil {
		return nil, "", err
	}

	b58, err := address.ValueForProtocol(multiaddr.P_IPFS)
	if err != nil {
		return nil, "", err
	}

	id, err := peer.IDB58Decode(b58)
	if err != nil {
		return nil, "", err
	}

	return addr, id, nil

}
