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

	logging "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-kbucket"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-peerstore"
	"github.com/libp2p/go-libp2p-swarm"
	"github.com/libp2p/go-libp2p/p2p/host/basic"
	"github.com/multiformats/go-multiaddr"
)

var log = logging.Logger("node")

/*
	the node can be used as both the client and the server
*/
type Node struct {
	host       *basichost.BasicHost
	id         peer.ID
	peerstore  peerstore.Peerstore
	routeTable *kbucket.RoutingTable
	context    context.Context
	config     *Config
}

// start a local node and join the node to network
func NewNode(config *Config) (*Node, error) {

	node := &Node{}
	node.context = context.Background()
	err := node.makeHost()
	if err != nil {
		log.Error("start node fail, can not make a basic host", err)
	}
	node.routeTable = kbucket.NewRoutingTable(
		config.bucketsize,
		kbucket.ConvertPeerID(node.id),
		config.latency,
		node.peerstore,
	)
	//node.start()
	return node, nil
}

// start the node and say hello to seedNodes, then start node discovery service
func Start(node *Node) {

	node.RegisterPingService()

	var wg sync.WaitGroup
	for _, seedNode := range node.config.seedNodes {
		wg.Add(1)
		go func(seedNode multiaddr.Multiaddr) {
			defer wg.Done()
			err := node.Hello(seedNode)
			if err != nil {
				log.Error("can not say hello to seed node.", seedNode, err)
			}

		}(seedNode)
	}
	wg.Wait()

	go node.Discovery(node.context)

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

	node.host, err = basichost.NewHost(node.context, network, options)
	return err
}

//TODO Say hello to seedNode
func (node *Node) Hello(seedNode multiaddr.Multiaddr) error {
	seedAddr, seedId, err := parseAddressFromMultiaddr(seedNode)
	if err != nil {
		log.Error("parse Address from seedNode failed", seedNode, err)
		return err
	}
	if node.id != seedId {
		for i := 0; i < 3; i++ {
			node.peerstore.AddAddr(
				seedId,
				seedAddr,
				peerstore.TempAddrTTL,
			)
			err := node.Ping(seedId)
			if err != nil {
				time.Sleep(time.Second)
				continue
			}
			break
		}
		if err != nil {
			log.Error("ping to seedNode failed", seedNode, err)
		}
		node.peerstore.SetAddr(
			seedId,
			seedAddr,
			peerstore.PermanentAddrTTL,
		)
		// Update the routing table.
		node.routeTable.Update(seedId)
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
