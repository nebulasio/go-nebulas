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
	"math"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/net"

	"github.com/nebulasio/go-nebulas/util/logging"
)

// const
var (
	LimitToSync = 1
	SyncBlock   = "syncblock"
	SyncReply   = "syncreply"
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
	LimitToSync = int(math.Sqrt(float64(len(allNode))))
	logging.VLog().Info("Sync: allNode -> ", allNode)
	if len(allNode) < LimitToSync {
		logging.VLog().Warn("Sync: node not enough.")
		return ErrNodeNotEnough
	}

	count := 0
	for i := 0; i < len(allNode); i++ {
		nodeID := allNode[i]
		addrs := node.peerstore.PeerInfo(nodeID).Addrs
		if len(addrs) > 0 {
			if node.host.Addrs()[0] == addrs[0] {
				logging.VLog().Warn("Sync: skip self")
				continue
			}

			key := nodeID.Pretty()
			if _, ok := node.stream.Load(key); ok {
				count++
				go func() {
					ns.SendMsg(SyncBlock, data, key)
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

	logging.VLog().Info("SendSyncReply: send sync addrs -> ", key)
	pb, _ := blocks.ToProto()
	data, _ := proto.Marshal(pb)
	if _, ok := ns.node.stream.Load(key); ok {
		go func() {
			ns.SendMsg(SyncReply, data, key)
		}()
		return
	}
	logging.VLog().Errorf("send syncReply to addrs %s fail", key)

}
