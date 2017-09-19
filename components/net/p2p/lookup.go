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
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-peerstore"
	"github.com/libp2p/go-libp2p-net"
	"github.com/nebulasio/go-nebulas/components/net/messages"
	"github.com/libp2p/go-libp2p-kbucket"
	log "github.com/sirupsen/logrus"
)

const lookupProtocolID = "/nebulas/lookup/1.0.0"

type LookupService struct {
	node *Node
}

// register lookup service
func (node *Node) RegisterLookupService() *LookupService {
	ls := &LookupService{node}
	node.host.SetStreamHandler(lookupProtocolID, ls.LookupHandler)
	log.Infof("node register lookup service success...")
	return ls
}

// Lookup from a node
func (node *Node) Lookup(pid peer.ID) ([]peerstore.PeerInfo, error) {

	log.Infof("node start lookup from peer %s", node.host.Addrs(), pid)
	s, err := node.host.NewStream(node.context, pid, lookupProtocolID)
	if err != nil {
		return nil , err
	}
	defer s.Close()

	wrappedStream := messages.WrapStream(s)
	messages, err := handleStream(wrappedStream)
	if err != nil {
		log.Errorf("Lookup error, can not receive data from node : %s", pid)
		return nil, err
	}
	return messages.Msg, nil
}

func handleStream(ws *messages.WrappedStream) (messages.LookupMessage, error) {
	var msg messages.LookupMessage
	err := ws.Dec.Decode(&msg)
	if err != nil {
		return msg, err
	}
	return msg, nil
}


// handle lookup request
func (p *LookupService) LookupHandler(s net.Stream) {
	defer s.Close()
	pid := s.Conn().RemotePeer()
	log.Debug("Receiving request for peers from", pid)

	peers := p.node.routeTable.NearestPeers(kbucket.ConvertPeerID(pid), p.node.config.maxSyncNodes)
	log.Debug("lookup nearest peers", peers)

	var peerList []peerstore.PeerInfo
	for i := range peers {
		peerInfo := p.node.peerstore.PeerInfo(peers[i])
		peerList = append(peerList, peerInfo)
	}

	msg := &messages.LookupMessage{
		Msg:    peerList,
	}

	messages.WrapStream(s).Enc.Encode(msg)
	messages.WrapStream(s).W.Flush()

	// Update the routing table.
	p.node.routeTable.Update(pid)
}