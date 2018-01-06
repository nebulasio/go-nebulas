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
	libnet "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	byteutils "github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

// const message name
const (
	NetworkID      = "networkid"
	NetworkIDReply = "renetworkid"
)

func (node *Node) handleNetworkIDMsg(data []byte, pid peer.ID, s libnet.Stream) {

	networkID := byteutils.Uint32(data)
	node.networkIDCache.Add(pid.Pretty(), networkID)

	networkIDData := byteutils.FromUint32(node.Config().NetworkID)
	if err := node.sendMsgWithStream(NetworkIDReply, networkIDData, s); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"error": err,
		}).Error("failed to send networkID message")
	}

}

func (node *Node) handleReNetworkIDMsg(data []byte, pid peer.ID) {
	networkID := byteutils.Uint32(data)
	node.networkIDCache.Add(pid.Pretty(), networkID)
}

func (node *Node) checkNetworkID(target string) bool {

	targetNetworkID, ok := node.networkIDCache.Get(target)
	if ok {
		logging.VLog().WithFields(logrus.Fields{
			"targetNetworkID": targetNetworkID,
			"result":          node.config.NetworkID & targetNetworkID.(uint32),
		}).Info("check networkID ok")

		if node.config.NetworkID&targetNetworkID.(uint32) > 0 {
			return true
		}
	}

	return false
}
