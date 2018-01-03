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
	"hash/crc32"

	"github.com/gogo/protobuf/proto"
	peer "github.com/libp2p/go-libp2p-peer"
	"github.com/nebulasio/go-nebulas/net"
	byteutils "github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

// Broadcast broadcast message
func (ns *NetService) Broadcast(name string, msg net.Serializable) {
	node := ns.node
	if node.synchronizing {
		return
	}
	ns.distribute(name, msg, false)
}

// Relay relay message
func (ns *NetService) Relay(name string, msg net.Serializable) {
	node := ns.node
	if node.synchronizing {
		return
	}
	ns.distribute(name, msg, true)
}

func (ns *NetService) nodeNotInRelayness(relayness []peer.ID, peers []peer.ID) []peer.ID {
	var list []peer.ID
	for _, p := range peers {
		if !InArray(p, relayness) {
			list = append(list, p)
		}
	}
	return list
}

func (ns *NetService) distribute(name string, msg net.Serializable, relay bool) {
	node := ns.node
	pbMsg, _ := msg.ToProto()
	data, err := proto.Marshal(pbMsg)
	if err != nil {
		return
	}

	var relayness []peer.ID
	dataChecksum := crc32.ChecksumIEEE(data)
	peers, exists := node.relayness.Get(dataChecksum)
	if exists {
		relayness = peers.([]peer.ID)
	}
	var allNode []peer.ID
	transfer := node.routeTable.ListPeers()
	if relay {
		allNode = ns.nodeNotInRelayness(relayness, node.routeTable.ListPeers())
		transfer = allNode
	}
	logging.VLog().WithFields(logrus.Fields{
		"msg":      msg,
		"transfer": transfer,
	}).Info("distribute: start distribute msg.")

	ns.doMsgTransfer(transfer, relayness, dataChecksum, name, data)

	if relay {
		ns.doRelay(allNode, relayness, dataChecksum)
	}
}

func (ns *NetService) doMsgTransfer(transfer []peer.ID, relayness []peer.ID, dataChecksum uint32, name string, data []byte) {
	node := ns.node
	for i := 0; i < len(transfer); i++ {
		nodeID := transfer[i]
		if InArray(nodeID, relayness) {
			logging.VLog().Infof("msgTransfer:  nodeID %s has already have the same message", nodeID)
			continue
		}
		addrs := node.peerstore.PeerInfo(nodeID).Addrs
		if len(addrs) == 0 || node.host.Addrs()[0].String() == addrs[0].String() {
			logging.VLog().Info("msgTransfer: skip self")
			continue
		}
		if len(addrs) > 0 {
			node.relayness.Add(dataChecksum, append(relayness, nodeID))
			go ns.SendMsg(name, data, nodeID.Pretty())
		}
	}
}

func (ns *NetService) doRelay(nodes []peer.ID, relayness []peer.ID, dataChecksum uint32) {
	node := ns.node
	for i := 0; i < len(nodes); i++ {
		nodeID := nodes[i]
		if InArray(nodeID, relayness) {
			logging.VLog().Infof("distribute: relay nodeID %s has already have the same message", nodeID)
			continue
		}
		addrs := node.peerstore.PeerInfo(nodeID).Addrs
		if len(addrs) == 0 || node.host.Addrs()[0].String() == addrs[0].String() {
			logging.VLog().Info("distribute: relay skip self")
			continue
		}
		if len(addrs) > 0 {
			node.relayness.Add(dataChecksum, append(relayness, nodeID))
			go ns.SendMsg(NewHashMsg, byteutils.FromUint32(dataChecksum), nodeID.Pretty())
		}
	}
}

// BroadcastNetworkID broadcast networkID when changed.
func (ns *NetService) BroadcastNetworkID(msg []byte) {
	node := ns.node
	if node.synchronizing {
		return
	}

	allNode := node.routeTable.ListPeers()
	for _, v := range allNode {
		go ns.SendMsg(NetworkID, msg, v.Pretty())
	}
}
