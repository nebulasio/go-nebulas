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
	"math/rand"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-peer"
	ma "github.com/multiformats/go-multiaddr"
	log "github.com/sirupsen/logrus"
)

/*
discovery node can discover other node or can be discovered by another node
and then update the routing table.
*/
func (net *NetService) discovery(ctx context.Context) {

	//FIXME  the sync routing table rate can be dynamic
	interval := 30 * time.Second
	ticker := time.NewTicker(interval)
	net.syncRoutingTable()
	for {
		select {
		case <-ticker.C:
			net.syncRoutingTable()
		case <-net.quitCh:
			log.Info("discovery service halting")
			return
		}
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
					log.Error("net.start: can not say hello to trusted node.", bootNode, err)
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
	nodeInfo := node.peerstore.PeerInfo(nodeID)
	if len(nodeInfo.Addrs) != 0 {
		if _, ok := node.stream.Load(nodeID.Pretty()); ok {
			net.SyncRoutes(nodeID)
		} else {
			// if stream not exist, create new connection to remote node.
			net.Hello(nodeID)
		}

	} else {
		node.routeTable.Remove(nodeID)
	}
}
