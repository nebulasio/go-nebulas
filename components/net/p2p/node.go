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
	"crypto/rand"
	"fmt"
	"io"
	mrand "math/rand"
	"strings"
	"sync"
	"time"

	"errors"

	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-kbucket"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-peerstore"
	"github.com/libp2p/go-libp2p-swarm"
	"github.com/libp2p/go-libp2p/p2p/host/basic"
	"github.com/multiformats/go-multiaddr"
	log "github.com/sirupsen/logrus"
)

// Node the node can be used as both the client and the server
type Node struct {
	host       *basichost.BasicHost
	id         peer.ID
	peerstore  peerstore.Peerstore
	routeTable *kbucket.RoutingTable
	context    context.Context
	config     *Config
	running    bool
}

// NewNode start a local node and join the node to network
func NewNode(config *Config) (*Node, error) {

	node := &Node{}
	node.config = config
	node.context = context.Background()
	log.Info("NewNode: node make Host success")
	err := node.makeHost()
	if err != nil {
		log.Error("NewNode: start node fail, can not make a basic host", err)
	}

	//node.start()
	return node, nil
}

// Start start the node and say hello to bootNodes, then start node discovery service
func (node *Node) Start() error {

	log.Info("Start: node create success...")
	log.Infof("Start: node info {id -> %s, address -> %s}", node.id, node.host.Addrs())
	if node.running {
		return errors.New("Start: node already running")
	}
	node.running = true
	log.Info("Start: node start join p2p network...")

	node.RegisterNetService()

	var wg sync.WaitGroup
	for _, bootNode := range node.config.BootNodes {
		wg.Add(1)
		go func(bootNode multiaddr.Multiaddr) {
			defer wg.Done()
			err := node.SayHello(bootNode)
			if err != nil {
				log.Error("Start: can not say hello to trusted node.", bootNode, err)
			}

		}(bootNode)
	}
	wg.Wait()

	go func() {
		node.Discovery(node.context)
	}()

	log.Infof("Start: node start and join to p2p network success and listening for connections on port %d... ", node.config.Port)

	return nil
}

func (node *Node) makeHost() error {

	ctx := node.context
	randseed := node.config.Randseed
	var r io.Reader
	if randseed == 0 {
		r = rand.Reader
	} else {
		r = mrand.New(mrand.NewSource(randseed))
	}

	priv, pub, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)

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

	node.routeTable = kbucket.NewRoutingTable(
		node.config.bucketsize,
		kbucket.ConvertPeerID(node.id),
		node.config.latency,
		node.peerstore,
	)

	node.routeTable.Update(node.id)

	address, err := multiaddr.NewMultiaddr(
		fmt.Sprintf(
			"/ip4/%s/tcp/%d",
			node.config.IP,
			node.config.Port,
		),
	)
	if err != nil {
		return err
	}

	network, err := swarm.NewNetwork(
		ctx,
		[]multiaddr.Multiaddr{address},
		node.id,
		node.peerstore,
		nil,
	)

	options := &basichost.HostOpts{}

	log.Infof("makeHost: boot node pretty id is %s", node.id.Pretty())
	node.host, err = basichost.NewHost(node.context, network, options)
	return err
}

// SayHello Say hello to trustedNode
func (node *Node) SayHello(bootNode multiaddr.Multiaddr) error {
	bootAddr, bootID, err := parseAddressFromMultiaddr(bootNode)
	if err != nil {
		log.Error("SayHello: parse Address from trustedNode failed", bootNode, err)
		return err
	}
	if node.id != bootID {
		for i := 0; i < 3; i++ {
			node.peerstore.AddAddr(
				bootID,
				bootAddr,
				peerstore.TempAddrTTL,
			)
			err := node.Hello(bootID)
			if err != nil {
				time.Sleep(time.Second)
				continue
			}
			break
		}
		if err != nil {
			log.Error("SayHello: ping to seedNode failed", bootNode, err)
		}
		log.Info("SayHello: node say hello to boot node success... ")
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
