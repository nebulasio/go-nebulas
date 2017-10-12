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
	"time"

	"github.com/libp2p/go-libp2p-peer"
	log "github.com/sirupsen/logrus"
)

/*
Discovery node can discover other node or can be discovered by another node
and then update the routing table.
*/
func (net *NetService) Discovery(ctx context.Context) {

	//FIXME  the sync routing table rate can be dynamic
	second := 3 * time.Second
	ticker := time.NewTicker(second)
	log.Infof("Discovery: node start discovery per %s...", second)
	for {
		select {
		case <-ticker.C:
			net.syncRoutingTable()
		case <-ctx.Done():
			log.Info("Discovery: discovery service halting")
			return
		}
	}
}

//sync route table
func (net *NetService) syncRoutingTable() {
	node := net.node
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
		nodeID := allNode[randomList[i]]
		if !asked[nodeID] {
			asked[nodeID] = true
			go func() {
				net.syncSingleNode(nodeID)
			}()
		}
	}
}

// sync single node routing table by peer.ID
func (net *NetService) syncSingleNode(nodeID peer.ID) {
	node := net.node
	log.Info("syncSingleNode: sync route -> ", nodeID)
	// skip self
	if nodeID == node.id {
		return
	}
	nodeInfo := node.peerstore.PeerInfo(nodeID)
	if len(nodeInfo.Addrs) != 0 {
		// net.syncRouteInfoFromSingleNode(nodeID)
		net.SyncRoutes(nodeID)
	} else {
		node.routeTable.Remove(nodeID)
	}
}

// func (net *NetService) syncRouteInfoFromSingleNode(nodeID peer.ID) {

// 	node := net.node
// 	node.SyncRoutes(nodeID)
// 	// for i := range reply {
// 	// 	if node.routeTable.Find(reply[i].ID) != "" || len(reply[i].Addrs) == 0 {
// 	// 		log.Warnf("syncRouteInfoFromSingleNode: node %s is already exist in route table", reply[i].ID)
// 	// 		continue
// 	// 	}
// 	// 	// Ping the peer.
// 	// 	node.peerstore.AddAddr(
// 	// 		reply[i].ID,
// 	// 		reply[i].Addrs[0],
// 	// 		peerstore.TempAddrTTL,
// 	// 	)
// 	// 	err := net.Hello(reply[i].ID)
// 	// 	if err != nil {
// 	// 		log.Errorf("syncRouteInfoFromSingleNode: ping peer %s fail %s", reply[i].ID, err)
// 	// 		continue
// 	// 	}
// 	// 	node.peerstore.AddAddr(
// 	// 		reply[i].ID,
// 	// 		reply[i].Addrs[0],
// 	// 		peerstore.PermanentAddrTTL,
// 	// 	)

// 	// 	// Update the routing table.
// 	// 	node.routeTable.Update(reply[i].ID)
// 	// }

// }
