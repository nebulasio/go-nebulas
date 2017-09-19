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
	"github.com/libp2p/go-libp2p-host"
)

type LookupService struct {
	Host host.Host
}

// register ping service
func (node *Node) RegisterLookupService() *LookupService {
	ls := &LookupService{node.host}
	node.host.SetStreamHandler(protocolID, ls.LookupHandler)
	return ls
}

//TODO Lookup from a node
func (node *Node) Lookup(pid peer.ID) ([]peerstore.PeerInfo, error) {

	return nil, nil
}

//TODO handle lookup request
func (p *LookupService) LookupHandler(s net.Stream) {

}