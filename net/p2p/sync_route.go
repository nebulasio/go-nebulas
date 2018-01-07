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
	kbucket "github.com/libp2p/go-libp2p-kbucket"
	libnet "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	peerstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/nebulasio/go-nebulas/net/messages"
	netpb "github.com/nebulasio/go-nebulas/net/pb"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

// const message name
const (
	SyncRoute      = "syncroute"
	SyncRouteReply = "resyncroute"
)

// SyncRoutes sync routing table from a peer
func (node *Node) SyncRoutes(pid peer.ID) {
	addrs := node.peerstore.PeerInfo(pid).Addrs
	if len(addrs) == 0 {
		logging.VLog().WithFields(logrus.Fields{
			"pid": pid,
		}).Warn("wrong pid addrs")
		node.hello(pid)
		return
	}
	data := []byte{}
	if err := node.sendMsg(SyncRoute, data, pid.Pretty()); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to send message")
		node.clearPeerStore(pid, addrs)
		return
	}

}

func (node *Node) handleSyncRouteMsg(data []byte, pid peer.ID, s libnet.Stream, addrs ma.Multiaddr, key string) bool {

	result := false
	defer func() {
		if !result {
			node.Bye(pid, []ma.Multiaddr{addrs}, s, key)
		}
	}()

	// get nearest peers from routeTable
	peers := node.routeTable.NearestPeers(kbucket.ConvertPeerID(pid), node.config.MaxSyncNodes)
	var peerList []*messages.PeerInfo
	for i := range peers {
		peerInfo := node.peerstore.PeerInfo(peers[i])
		if len(peerInfo.Addrs) == 0 {
			logging.VLog().WithFields(logrus.Fields{
				"nodeId": peerInfo.ID.Pretty(),
			}).Warn("node addrs is nil")
			continue
		}
		var addres []string
		for _, v := range peerInfo.Addrs {
			addres = append(addres, v.String())
		}
		peer := messages.NewPeerInfoMessage(peerInfo.ID, addres)
		peerList = append(peerList, peer)
	}

	logging.VLog().WithFields(logrus.Fields{
		"remoteId":    pid.Pretty(),
		"remoteAddrs": addrs,
		"count":       len(peerList),
	}).Info("reply sync route to remote node")

	peersMessage := messages.NewPeersMessage(peerList)

	pb, _ := peersMessage.ToProto()
	data, err := proto.Marshal(pb)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to marshal proto")
		return result
	}

	if err := node.sendMsg(SyncRouteReply, data, key); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to send msg")
		return result
	}

	node.routeTable.Update(pid)
	result = true
	return result
}

func (node *Node) handleSyncRouteReplyMsg(data []byte, pid peer.ID, s libnet.Stream, addrs ma.Multiaddr) bool {
	peers := new(messages.Peers)
	pb := new(netpb.Peers)

	if err := proto.Unmarshal(data, pb); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to unmarshal proto")
		return false
	}

	if err := peers.FromProto(pb); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to get peers from proto.")
		return false
	}

	for i := range peers.Peers() {
		id := peers.Peers()[i].ID()
		if node.routeTable.Find(id) != "" || len(peers.Peers()[i].Addrs()) == 0 {
			logging.VLog().WithFields(logrus.Fields{
				"id": id.Pretty(),
			}).Warn("node is already exist in route table")
			continue
		}
		var addres []ma.Multiaddr
		for _, v := range peers.Peers()[i].Addrs() {
			addr, _ := ma.NewMultiaddr(v)
			addres = append(addres, addr)
		}

		logging.VLog().WithFields(logrus.Fields{
			"id":    id.Pretty(),
			"addrs": addres,
		}).Info("discover new node")

		node.peerstore.AddAddrs(
			id,
			addres,
			peerstore.ProviderAddrTTL,
		)
		if err := node.hello(id); err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"id":  id.Pretty(),
				"err": err,
			}).Error("Failed to say hello to the peer")
			continue
		}

		node.routeTable.Update(id)
	}
	return true
}
