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
	"github.com/nebulasio/go-nebulas/components/net"
	log "github.com/sirupsen/logrus"
)

// Broadcast broadcast message
func (ns *NetService) Broadcast(name string, msg net.Serializable) {

	if !ns.node.synchronized {
		return
	}
	node := ns.node
	pbMsg, _ := msg.ToProto()
	data, err := proto.Marshal(pbMsg)
	if err != nil {
		return
	}

	log.Info("Broadcast: start broadcast msg...", msg)
	allNode := node.routeTable.ListPeers()
	log.Info("Broadcast: allNode -> ", allNode)

	for i := 0; i < len(allNode); i++ {

		nodeID := allNode[i]
		addrs := node.peerstore.PeerInfo(nodeID).Addrs

		if len(addrs) == 0 || node.host.Addrs()[0].String() == addrs[0].String() {
			log.Warn("Broadcast: skip self")
			continue
		}
		if len(addrs) > 0 {
			key := GenerateKey(addrs[0], nodeID)
			go func() {
				ns.SendMsg(name, data, key)
			}()
		}

	}

}

// Relay message
func (ns *NetService) Relay(name string, msg net.Serializable) {
	// TODO(@leon): relay protocol
	if !ns.node.synchronized {
		return
	}
	ns.Broadcast(name, msg)
}
