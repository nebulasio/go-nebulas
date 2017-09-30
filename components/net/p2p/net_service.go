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
	nnet "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	peerstore "github.com/libp2p/go-libp2p-peerstore"
	log "github.com/sirupsen/logrus"
)

const ProtocolID = "/neb/1.0.0"

type NetService struct {
	node *Node
}

func (node *Node) RegisterNetService() *NetService {
	ns := &NetService{node}
	node.host.SetStreamHandler(ProtocolID, ns.msgHandler)
	log.Infof("NetService: node register net service success...")
	return ns
}

func (ns *NetService) msgHandler(s nnet.Stream) {
	// TODO handle msg coming
}

func (node *Node) SendMsg(msg []byte, pid peer.ID) {
	// TODO Send message
}

func (node *Node) Hello(pid peer.ID) error {
	// TODO Say Hello
	return nil
}

func (node *Node) SyncRoutes(pid peer.ID) ([]peerstore.PeerInfo, error) {
	return nil, nil
}
