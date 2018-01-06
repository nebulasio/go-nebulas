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

// const
const (
	NewHashMsg = "newhashmsg"
)

func (node *Node) broadcast(name string, msg net.Serializable) {
	// node can not broadcast or relay message if it is in synchronizing.
	if node.synchronizing {
		return
	}
	node.distribute(name, msg, false)
}

func (node *Node) relay(name string, msg net.Serializable) {
	if node.synchronizing {
		return
	}
	node.distribute(name, msg, true)
}

func (node *Node) distribute(name string, msg net.Serializable, relay bool) {

	pbMsg, _ := msg.ToProto()
	data, err := proto.Marshal(pbMsg)
	if err != nil {
		return
	}

	// check relay blacklist
	var relayness []peer.ID
	dataChecksum := crc32.ChecksumIEEE(data)
	peers, exists := node.relayness.Get(dataChecksum)
	if exists {
		relayness = peers.([]peer.ID)
	}

	var allNode []peer.ID
	transfer := node.routeTable.ListPeers()
	if relay {
		allNode = node.nodeNotInRelayness(relayness, node.routeTable.ListPeers())
		transfer = allNode
	}

	logging.VLog().WithFields(logrus.Fields{
		"msg":      msg,
		"transfer": transfer,
	}).Info("distribute msg")

	node.doMsgTransfer(transfer, relayness, dataChecksum, name, data)

	if relay {
		// notice the node in my route table that i have received the message, when others received the message they won`t relay to me.
		node.doNotice(allNode, relayness, dataChecksum)
	}
}

func (node *Node) doMsgTransfer(transfer []peer.ID, relayness []peer.ID, dataChecksum uint32, name string, data []byte) {
	for i := 0; i < len(transfer); i++ {
		nodeID := transfer[i]
		if InArray(nodeID, relayness) {
			logging.VLog().WithFields(logrus.Fields{
				"nodeID": nodeID.Pretty(),
			}).Info("target node has received the same message")
			continue
		}
		addrs := node.peerstore.PeerInfo(nodeID).Addrs
		if len(addrs) == 0 || node.host.Addrs()[0].String() == addrs[0].String() {
			continue
		}
		if len(addrs) > 0 {
			node.relayness.Add(dataChecksum, append(relayness, nodeID))
			go node.sendMsg(name, data, nodeID.Pretty())
		}
	}
}

func (node *Node) doNotice(nodes []peer.ID, relayness []peer.ID, dataChecksum uint32) {
	for i := 0; i < len(nodes); i++ {
		nodeID := nodes[i]
		addrs := node.peerstore.PeerInfo(nodeID).Addrs
		if len(addrs) == 0 || node.host.Addrs()[0].String() == addrs[0].String() {
			continue
		}
		if len(addrs) > 0 {
			node.relayness.Add(dataChecksum, append(relayness, nodeID))
			go node.sendMsg(NewHashMsg, byteutils.FromUint32(dataChecksum), nodeID.Pretty())
		}
	}
}

func (node *Node) broadcastNetworkID(msg []byte) {
	if node.synchronizing {
		return
	}

	allNode := node.routeTable.ListPeers()
	for _, v := range allNode {
		go node.sendMsg(NetworkID, msg, v.Pretty())
	}
}

func (node *Node) handleNewHashMsg(data []byte, pid peer.ID) {
	var relayness []peer.ID
	peers, exists := node.relayness.Get(byteutils.Uint32(data))
	if exists {
		relayness = peers.([]peer.ID)
	}
	node.relayness.Add(byteutils.Uint32(data), append(relayness, pid))
}

func (node *Node) nodeNotInRelayness(relayness []peer.ID, peers []peer.ID) []peer.ID {
	var list []peer.ID
	for _, p := range peers {
		if !InArray(p, relayness) {
			list = append(list, p)
		}
	}
	return list
}
