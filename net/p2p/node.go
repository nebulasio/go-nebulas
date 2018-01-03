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
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	mrand "math/rand"
	"net"
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
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
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
		logging.VLog().Error("start node fail, can not init node", err)
		return nil, err
	}
	logging.CLog().WithFields(logrus.Fields{
		"node.listen": node.config.Listen,
	}).Info("node init success")
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
			return errors.New("The port already in use")
		}
	}

	return nil
}

// GenerateEd25519Key generate a privKey and pubKey by ed25519.
func GenerateEd25519Key() (crypto.PrivKey, crypto.PubKey, error) {
	randseedstr := randSeed(64)
	randseed, err := hex.DecodeString(randseedstr)
	priv, pub, err := crypto.GenerateEd25519Key(
		bytes.NewReader(randseed),
	)
	return priv, pub, err
}

func (node *Node) generatePeerStore() error {
	filename := node.Config().PrivateKey
	priv, pub, err := getPeerstoreFromFile(filename)
	if err != nil {
		var randseedstr string
		if len(node.Config().BootNodes) == 0 {
			// seednode
			randseedstr = letterBytes
		} else {
			randseedstr = randSeed(64)
		}
		randseed, err := hex.DecodeString(randseedstr)
		priv, pub, err = crypto.GenerateEd25519Key(
			bytes.NewReader(randseed),
		)
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

func getPeerstoreFromFile(filename string) (crypto.PrivKey, crypto.PubKey, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, nil, errors.New("get private_key from file error")
	}
	privb, err := base64.StdEncoding.DecodeString(string(b))
	priv, err := crypto.UnmarshalPrivateKey(privb)
	if err != nil {
		return nil, nil, errors.New("get private_key from file error")
	}
	pub := priv.GetPublic()
	return priv, pub, nil
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
	node.streamCache = pdeque.NewPriorityDeque(less)
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
