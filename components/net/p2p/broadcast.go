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
	log "github.com/sirupsen/logrus"
)

// Broadcast broadcast message
func (ns *NetService) Broadcast(name string, msg proto.Message) {

	node := ns.node
	data, err := proto.Marshal(msg)
	if err != nil {
		return
	}

	log.Info("Broadcast: start broadcast msg...", msg)
	allNode := node.routeTable.ListPeers()
	log.Info("Broadcast: allNode -> ", allNode)

	for i := 0; i < len(allNode); i++ {

		nodeID := allNode[i]
		addrs := node.peerstore.PeerInfo(nodeID).Addrs
		log.Info(node.host.Addrs()[0].String())
		if len(addrs) == 0 || node.host.Addrs()[0].String() == addrs[0].String() {
			log.Warn("Broadcast: skip self")
			continue
		}
		go func() {
			ns.SendMsg(name, data, addrs[0].String())
		}()
	}

}
