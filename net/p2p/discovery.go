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
	"path"
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
	routingTableCacheFile = "ROUTINGTABLE.TMP"
)

/*
discovery node can discover other node or can be discovered by another node
and then update the routing table.
*/
func (node *Node) discovery(ctx context.Context) {

	//FIXME  the sync routing table rate can be dynamic
	interval := 30 * time.Second
	ticker := time.NewTicker(interval)
	time.Sleep(1 * time.Second)
	node.loadRoutingTableFromDisk()
	node.syncRoutingTable()
	go node.persistRoutingTable()
	for {
		select {
		case <-ticker.C:
			node.syncRoutingTable()
		case <-node.ns.quitCh:
			logging.VLog().Info("discovery service halting")
			return
		}
	}
}

func (node *Node) persistRoutingTable() {
	ticker := time.NewTicker(60 * time.Second)
	for {
		select {
		case <-ticker.C:
			node.saveRoutingTableToDisk()
		case <-node.ns.quitCh:
			return
		}
	}
}

func (node *Node) saveRoutingTableToDisk() {
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
	if err := ioutil.WriteFile(node.getRoutingTableCacheFilePath(), []byte(str), 0644); err != nil {
		logging.VLog().Warn("failed to persist routing table")
	}
}

func (node *Node) loadRoutingTableFromDisk() {

	b, err := ioutil.ReadFile(node.getRoutingTableCacheFilePath())
	if err != nil {
		logging.VLog().Warn("failed to load routing table from disk")
		return
	}
	contents := strings.Split(string(b), ",")
	count := 0
	for _, v := range contents {

		multiaddr, err := ma.NewMultiaddr(v)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err": err,
			}).Warn("failed to new multiaddr")
			continue
		}

		addr, ID, err := node.parseAddressFromMultiaddr(multiaddr)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"multiaddr": multiaddr,
				"error":     err,
			}).Warn("failed to parse address")
			continue
		}

		node.peerstore.AddAddr(
			ID,
			addr,
			peerstore.ProviderAddrTTL,
		)

		if err := node.hello(ID); err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"ID":    ID,
				"addr":  addr,
				"error": err,
			}).Warn("failed to say hello to node")
			continue
		}

		node.peerstore.AddAddr(
			ID,
			addr,
			peerstore.PermanentAddrTTL)
		node.routeTable.Update(ID)
		count++
	}

	logging.CLog().Infof("loaded %d nodes to routing table from disk.", count)

}

func (node *Node) getRoutingTableCacheFilePath() string {
	return path.Join(node.config.RoutingTableDir, routingTableCacheFile)
}

//sync route table
func (node *Node) syncRoutingTable() {

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
				go node.syncSingleNode(nodeID)
			}
		}
	} else if nodeAccount == 0 && len(node.Config().BootNodes) > 0 { // If disconnect from the network, say hello to seed node, reconnect to the network.
		var wg sync.WaitGroup
		for _, seed := range node.config.BootNodes {
			wg.Add(1)
			go func(seed ma.Multiaddr) {
				defer wg.Done()
				err := node.sayHelloToSeed(seed)
				if err != nil {
					logging.VLog().WithFields(logrus.Fields{
						"seed": seed,
						"err":  err,
					}).Error("failed to say hello to seed")
				}

			}(seed)
		}
		wg.Wait()
	}

}

// sync single node routing table by peer.ID
func (node *Node) syncSingleNode(nodeID peer.ID) {
	// skip self
	if nodeID == node.id {
		return
	}

	if _, ok := node.stream.Load(nodeID.Pretty()); ok {
		node.SyncRoutes(nodeID)
	} else {
		// if stream not exist, create new connection to remote node.
		node.hello(nodeID)
	}
}
