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
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/net"

	log "github.com/sirupsen/logrus"
)

// const
const (
	LimitToSync = 1
)

// errors
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
	log.Info("Sync: allNode -> ", allNode)
	if len(allNode) < LimitToSync {
		log.Warn("Sync: node not enough.")
		return ErrNodeNotEnough
	}

	count := 0
	for i := 0; i < len(allNode); i++ {
		nodeID := allNode[i]
		addrs := node.peerstore.PeerInfo(nodeID).Addrs
		if len(addrs) > 0 {
			if node.host.Addrs()[0] == addrs[0] {
				log.Warn("Sync: skip self")
				continue
			}

			key := nodeID.Pretty()
			if _, ok := node.stream.Load(key); ok {
				count++
				go func() {
					ns.SendMsg("syncblock", data, key)
				}()
			}
		}
	}
	if count < LimitToSync {
		return ErrNodeNotEnough
	}
	return nil
}

// SendSyncReply send sync reply message to remote peer
func (ns *NetService) SendSyncReply(key string, blocks net.Serializable) {

	// addrs := ns.node.syncList
	log.Info("SendSyncReply: send sync addrs -> ", key)
	pb, _ := blocks.ToProto()
	data, _ := proto.Marshal(pb)
	for {
		if _, ok := ns.node.stream.Load(key); ok {
			go func() {
				ns.SendMsg("syncreply", data, key)
			}()
			return
		}
		log.Info("SendSyncReply: sleep for 1 second")
		time.Sleep(1 * time.Second)
	}

}
