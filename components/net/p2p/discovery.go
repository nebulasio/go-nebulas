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
	"time"
	"math/rand"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-peerstore"
	log "github.com/sirupsen/logrus"
)

/*
node can discover other node or can be discovered by another node
and then update the routing table.
*/
func (node *Node) Discovery(ctx context.Context) {

	//FIXME  the sync routing table rate can be dynamic
	second := 5 * time.Second
	ticker := time.NewTicker(second)
	log.Infof("Discovery: node start discovery per %s...", second)
	for {
		select {
		case <-ticker.C:
			node.syncRoutingTable()
		case <-ctx.Done():
			log.Info("Discovery: discovery service halting")
			return
		}
	}
}

//sync route table
func (node *Node) syncRoutingTable() {
	log.Infof("syncRoutingTable: node start sync routing table...")
	asked := make(map[peer.ID]bool)
	allNode := node.routeTable.ListPeers()
	log.Infof("syncRoutingTable: node %s routing table: %s", node.host.Addrs(), allNode)
	randomList := rand.Perm(len(allNode))
	var nodeAccount int
	if len(allNode) > node.config.maxSyncNodes {
		nodeAccount = node.config.maxSyncNodes
	} else {
		nodeAccount = len(allNode)
	}

	for i := 0; i < nodeAccount; i++ {
		nodeId := allNode[randomList[i]]
		if !asked[nodeId] {
			asked[nodeId] = true
			go func() {
				node.syncSingleNode(nodeId)
			}()
		}
	}
}

// sync single node routing table by peer.ID
func (node *Node) syncSingleNode(nodeId peer.ID) {
	// skip self
	if nodeId == node.id {
		return
	}
	nodeInfo := node.peerstore.PeerInfo(nodeId)
	if len(nodeInfo.Addrs) != 0 {
		node.syncRouteInfoFromSingleNode(nodeId)
	} else {
		node.routeTable.Remove(nodeId)
	}
}

func (node *Node) syncRouteInfoFromSingleNode(nodeId peer.ID) {

	reply, err := node.Lookup(nodeId)
	if err != nil {
		log.Errorf("")
		return
	}
	for i := range reply {
		if node.routeTable.Find(reply[i].ID) != "" || len(reply[i].Addrs) == 0 {
			continue
		}
		// Ping the peer.
		err := node.Ping(reply[i].ID)
		if err != nil {
			continue
		}
		node.peerstore.SetAddrs(
			reply[i].ID,
			reply[i].Addrs,
			peerstore.ProviderAddrTTL,
		)

		// Update the routing table.
		node.routeTable.Update(reply[i].ID)
	}

}
