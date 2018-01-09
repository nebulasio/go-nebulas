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
	libnet "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	peerstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/nebulasio/go-nebulas/net/messages"
	netpb "github.com/nebulasio/go-nebulas/net/pb"
	byteutils "github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

// const message name
const (
	HELLO = "hello"
	OK    = "ok"
	BYE   = "bye"
)

// say hello to trustedNode
func (node *Node) sayHelloToSeed(seed ma.Multiaddr) error {

	addr, ID, err := node.parseAddressFromMultiaddr(seed)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"seed":  seed,
			"error": err,
		}).Error("Failed to parse Address from trustedNode")
		return err
	}

	node.bootIds = append(node.bootIds, ID.Pretty())
	node.peerstore.AddAddr(
		ID,
		addr,
		peerstore.ProviderAddrTTL,
	)
	if node.host.Addrs()[0].String() != addr.String() {
		if err := node.hello(ID); err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"seed": seed,
				"err":  err,
			}).Error("Failed to say hello to seed")
			return err
		}
		logging.CLog().WithFields(logrus.Fields{
			"seed": seed,
		}).Info("say hello to a node success")

		node.peerstore.AddAddr(
			ID,
			addr,
			peerstore.PermanentAddrTTL)
		// Update the routing table.
		node.routeTable.Update(ID)
	}
	return nil
}

// say hello to a peer
func (node *Node) hello(pid peer.ID) error {

	stream, err := node.host.NewStream(
		node.context,
		pid,
		ProtocolID,
	)
	if err != nil {
		return err
	}

	message := messages.NewHelloMessage(node.id.String(), ClientVersion)
	pb, _ := message.ToProto()
	data, err := proto.Marshal(pb)
	if err != nil {
		return err
	}

	if err = node.sendMsgWithStream(HELLO, data, stream); err != nil {
		return err
	}
	// call streamHandler explicitly to start loop to handle stream origined from this node.
	go node.messageHandler(stream)
	return nil
}

func (node *Node) handleOkMsg(data []byte, pid peer.ID, s libnet.Stream, addrs ma.Multiaddr, key string) bool {
	result := false
	defer func() {
		if !result {
			node.Bye(pid, []ma.Multiaddr{addrs}, s, key)
		}
	}()

	ok := new(messages.HelloMessage)
	pb := new(netpb.Hello)
	if err := proto.Unmarshal(data, pb); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to handle ok msg")
		return result
	}
	if err := ok.FromProto(pb); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to handle ok msg")
		return result
	}

	if ok.NodeID == pid.String() && ok.ClientVersion == ClientVersion {
		streamStore := NewStreamStore(key, SOK, s)
		node.stream.Store(key, streamStore)
		node.streamCache.Insert(streamStore)
		node.peerstore.AddAddr(
			pid,
			addrs,
			peerstore.PermanentAddrTTL,
		)
		node.routeTable.Update(pid)

		result = true
		return result
	}

	logging.VLog().Error("recv incorrect ok message")
	return result

}

func (node *Node) handleHelloMsg(data []byte, pid peer.ID, s libnet.Stream, addrs ma.Multiaddr, key string) bool {
	result := false
	defer func() {
		if !result {
			node.Bye(pid, []ma.Multiaddr{addrs}, s, key)
		}
	}()

	hello := new(messages.HelloMessage)
	pb := new(netpb.Hello)
	if err := proto.Unmarshal(data, pb); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to handle hello msg")
		return result
	}
	if err := hello.FromProto(pb); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to handle hello msg")
		return result
	}

	logging.VLog().WithFields(logrus.Fields{
		"hello.NodeID":  hello.NodeID,
		"pid":           pid,
		"addrs":         addrs.String(),
		"ClientVersion": hello.ClientVersion,
	}).Info("receive hello message.")

	//Todo: clientVersion backwards compatible
	if hello.NodeID == pid.String() && hello.ClientVersion == ClientVersion {
		ok := messages.NewHelloMessage(node.id.String(), ClientVersion)
		pbok, err := ok.ToProto()
		okdata, err := proto.Marshal(pbok)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err": err,
			}).Error("Failed to send ok message")
			return result
		}

		node.peerstore.AddAddr(
			pid,
			addrs,
			peerstore.PermanentAddrTTL,
		)

		if err := node.sendMsgWithStream(OK, okdata, s); err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err": err,
			}).Error("Failed to send ok message")
			return result
		}

		networkIDData := byteutils.FromUint32(node.Config().NetworkID)
		if err := node.sendMsgWithStream(NetworkID, networkIDData, s); err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err": err,
			}).Error("Failed to send networkID message")
			return result
		}

		streamStore := NewStreamStore(key, SOK, s)
		node.stream.Store(key, streamStore)
		node.streamCache.Insert(streamStore)
		node.routeTable.Update(pid)
		result = true
		return result
	}
	return result

}

// Bye say bye to a peer, and close connection.
func (node *Node) Bye(pid peer.ID, addrs []ma.Multiaddr, s libnet.Stream, key string) {
	logging.VLog().WithFields(logrus.Fields{
		"pid":  pid.Pretty(),
		"addr": addrs,
	}).Info("Say bye to a node")
	node.clearPeerStore(pid, addrs)
	node.stream.Delete(key)
	s.Close()
}
