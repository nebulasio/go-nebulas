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
	"errors"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/components/net"
	log "github.com/sirupsen/logrus"
)

var (
	ErrNodeNotEnough = errors.New("node is not enough")
)

// Sync do something ready to sync
func (ns *NetService) Sync(tail net.Serializable) error {
	node := ns.node
	pb, _ := tail.ToProto()
	data, err := proto.Marshal(pb)
	if err != nil {
		return err
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
	return nil
}

// SendSyncReply send sync reply message to remote peer
func (ns *NetService) SendSyncReply(blocks net.Serializable) {

	addrs := ns.node.syncList
	pb, _ := blocks.ToProto()
	data, _ := proto.Marshal(pb)

	for i := 0; i < len(addrs); i++ {
		go func() {
			ns.SendMsg("syncreply", data, addrs[i])
		}()
	}
}
