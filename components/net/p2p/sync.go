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
	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/core"
	log "github.com/sirupsen/logrus"
)

// PreSync do something ready to sync
func (ns *NetService) PreSync(tail *core.Block) {
	node := ns.node
	pb, _ := tail.ToProto()
	data, err := proto.Marshal(pb)
	if err != nil {
		return
	}

	allNode := node.routeTable.ListPeers()
	log.Info("PreSync: allNode -> ", allNode)

	for i := 0; i < len(allNode); i++ {
		nodeID := allNode[i]
		addrs := node.peerstore.PeerInfo(nodeID).Addrs
		if len(addrs) > 0 && node.host.Addrs()[0] == addrs[0] {
			log.Warn("PreSync: skip self")
			continue
		}
		go func() {
			ns.SendMsg("syncblock", data, addrs[0].String())
		}()
	}
}

// SendSyncReply send sync reply message to remote peer
func (ns *NetService) SendSyncReply(data []byte) {
	addrs := ns.node.syncList
	for i := 0; i < len(addrs); i++ {
		go func() {
			ns.SendMsg("syncreply", data, addrs[i])
		}()
	}
}
