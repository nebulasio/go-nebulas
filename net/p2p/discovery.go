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
discovery node can discover other node or can be discovered by another node
and then update the routing table.
*/
func (net *NetService) discovery(ctx context.Context) {

	//FIXME  the sync routing table rate can be dynamic
	second := 5 * time.Second
	ticker := time.NewTicker(second)
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
	// TODO: should set seed?
	randomList := rand.Perm(len(allNode))
	var nodeAccount int
	if len(allNode) > node.config.MaxSyncNodes {
		nodeAccount = node.config.MaxSyncNodes
	} else {
		nodeAccount = len(allNode)
	}

	for i := 0; i < nodeAccount; i++ {
		nodeID := allNode[randomList[i]]
		if !asked[nodeID] {
			asked[nodeID] = true
			go net.syncSingleNode(nodeID)
		}
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
		//key, err := GenerateKey(nodeInfo.Addrs[0], nodeID)
		//if err != nil {
		//	return
		//}
		if _, ok := node.stream[nodeID.Pretty()]; ok {
			net.SyncRoutes(nodeID)
		}

	} else {
		node.routeTable.Remove(nodeID)
	}
}
