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
	"math"
	"reflect"

	"github.com/gogo/protobuf/proto"
	peer "github.com/libp2p/go-libp2p-peer"
	"github.com/nebulasio/go-nebulas/net"
	byteutils "github.com/nebulasio/go-nebulas/util/byteutils"
	log "github.com/sirupsen/logrus"
)

// Broadcast broadcast message
func (ns *NetService) Broadcast(name string, msg net.Serializable) {
	/* 	node := ns.node */
	/* 	if !node.synchronized {
		return
	} */
	ns.distribute(name, msg, false)
}

// Relay relay message
func (ns *NetService) Relay(name string, msg net.Serializable) {
	/* 	node := ns.node
	   	if !node.synchronized {
	   		return
	   	} */
	ns.distribute(name, msg, true)
}

func (ns *NetService) nodeNotInRelayness(relayness []peer.ID, peers []peer.ID) []peer.ID {
	var list []peer.ID
	for _, p := range peers {
		if !inArray(p, relayness) {
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
	allNode := ns.nodeNotInRelayness(relayness, node.routeTable.ListPeers())
	transfer := allNode
	if relay {
		transfer = allNode[:int(math.Sqrt(float64(len(allNode))))]
	}
	log.WithFields(log.Fields{
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
		if inArray(nodeID, relayness) {
			log.Warnf("distribute:  nodeID %s has already have the same message", nodeID)
			continue
		}
		addrs := node.peerstore.PeerInfo(nodeID).Addrs
		if len(addrs) == 0 || node.host.Addrs()[0].String() == addrs[0].String() {
			log.Warn("distribute: skip self")
			continue
		}
		if len(addrs) > 0 {
			// key, err := GenerateKey(addrs[0], nodeID)
			// if err != nil {
			// 	log.Warn("distribute:  the addrs format is incorrect")
			// 	continue
			// }
			node.relayness.Add(dataChecksum, append(relayness, nodeID))
			go ns.SendMsg(name, data, nodeID.Pretty())
		}
	}
}

func (ns *NetService) doRelay(nodes []peer.ID, relayness []peer.ID, dataChecksum uint32) {
	node := ns.node
	for i := 0; i < len(nodes); i++ {
		nodeID := nodes[i]
		if inArray(nodeID, relayness) {
			log.Warnf("distribute: relay nodeID %s has already have the same message", nodeID)
			continue
		}
		addrs := node.peerstore.PeerInfo(nodeID).Addrs
		if len(addrs) == 0 || node.host.Addrs()[0].String() == addrs[0].String() {
			log.Warn("distribute: relay skip self")
			continue
		}
		if len(addrs) > 0 {
			//key, err := GenerateKey(addrs[0], nodeID)
			//if err != nil {
			//	log.Warn("distribute: relay  the addrs format is incorrect")
			//	continue
			//}
			node.relayness.Add(dataChecksum, append(relayness, nodeID))
			go ns.SendMsg(NewHashMsg, byteutils.FromUint32(dataChecksum), nodeID.Pretty())
		}
	}
}

func inArray(obj interface{}, array interface{}) bool {
	arrayValue := reflect.ValueOf(array)

	if reflect.TypeOf(array).Kind() == reflect.Array || reflect.TypeOf(array).Kind() == reflect.Slice {
		for i := 0; i < arrayValue.Len(); i++ {
			if arrayValue.Index(i).Interface() == obj {
				return true
			}
		}
	}
	return false
}
