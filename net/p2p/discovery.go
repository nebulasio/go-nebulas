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
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-peer"
	peerstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

var (
	filename = "/routingtable"
)

/*
discovery node can discover other node or can be discovered by another node
and then update the routing table.
*/
func (net *NetService) discovery(ctx context.Context) {

	//FIXME  the sync routing table rate can be dynamic
	interval := 30 * time.Second
	ticker := time.NewTicker(interval)
	time.Sleep(1 * time.Second)
	net.loadRoutingTableFromDisk()
	net.syncRoutingTable()
	go net.persistRoutingTable()
	for {
		select {
		case <-ticker.C:
			net.syncRoutingTable()
		case <-net.quitCh:
			logging.VLog().Info("discovery service halting")
			return
		}
	}
}

func (net *NetService) persistRoutingTable() {
	ticker := time.NewTicker(60 * time.Second)
	for {
		select {
		case <-ticker.C:
			net.persistRt()
		case <-net.quitCh:
			return
		}
	}
}

func (net *NetService) persistRt() {
	node := net.node
	allnode := node.routeTable.ListPeers()
	var nodes []string
	for _, v := range allnode {
		if len(node.PeerStore().Addrs(v)) > 0 {
			addr := node.PeerStore().Addrs(v)[0]
			tmp := fmt.Sprintf("%s/ipfs/%s", addr, v.Pretty())
			nodes = append(nodes, tmp)
		}
	}
	str := strings.Join(nodes, ",")
	if err := ioutil.WriteFile(node.config.RoutingTableDir+filename, []byte(str), os.ModePerm); err != nil {
		logging.VLog().Warn("failed to persist routing table")
	}
}

func (net *NetService) loadRoutingTableFromDisk() {
	node := net.node
	b, err := ioutil.ReadFile(node.config.RoutingTableDir + filename)
	if err != nil {
		logging.VLog().Warn("failed to load routing table from disk")
		return
	}
	contents := strings.Split(string(b), ",")
	for _, v := range contents {

		multiaddr, err := ma.NewMultiaddr(v)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err": err,
			}).Warn("new multiaddr failed")
			continue
		}

		addr, ID, err := net.parseAddressFromMultiaddr(multiaddr)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"multiaddr": multiaddr,
				"error":     err,
			}).Warn("parse address failed")
			continue
		}

		node.peerstore.AddAddr(
			ID,
			addr,
			peerstore.ProviderAddrTTL,
		)

		if err := net.Hello(ID); err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"ID":    ID,
				"addr":  addr,
				"error": err,
			}).Warn("say hello to node failed")
			continue
		}

		node.peerstore.AddAddr(
			ID,
			addr,
			peerstore.PermanentAddrTTL)
		node.routeTable.Update(ID)
	}

}

//sync route table
func (net *NetService) syncRoutingTable() {
	node := net.node
	asked := make(map[peer.ID]bool)
	allNode := node.routeTable.ListPeers()
	rand.Seed(time.Now().UnixNano())
	randomList := rand.Perm(len(allNode))
	var nodeAccount int
	if len(allNode) > node.config.MaxSyncNodes {
		nodeAccount = node.config.MaxSyncNodes
	} else {
		nodeAccount = len(allNode)
	}

	if nodeAccount > 0 {
		for i := 0; i < nodeAccount; i++ {
			nodeID := allNode[randomList[i]]
			if !asked[nodeID] {
				asked[nodeID] = true
				go net.syncSingleNode(nodeID)
			}
		}
	} else if nodeAccount == 0 && len(node.Config().BootNodes) > 0 { // If disconnect from the network, say hello to seed node, reconnect to the network.
		var wg sync.WaitGroup
		for _, bootNode := range node.config.BootNodes {
			wg.Add(1)
			go func(bootNode ma.Multiaddr) {
				defer wg.Done()
				err := net.SayHello(bootNode)
				if err != nil {
					logging.VLog().Error("net.start: can not say hello to trusted node.", bootNode, err)
				}

			}(bootNode)
		}
		wg.Wait()
	}

}

// sync single node routing table by peer.ID
func (net *NetService) syncSingleNode(nodeID peer.ID) {
	node := net.node
	// skip self
	if nodeID == node.id {
		return
	}

	if _, ok := node.stream.Load(nodeID.Pretty()); ok {
		net.SyncRoutes(nodeID)
	} else {
		// if stream not exist, create new connection to remote node.
		net.Hello(nodeID)
	}
}
