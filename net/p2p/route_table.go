// Copyright (C) 2018 go-nebulas authors
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
	"crypto"

	"github.com/multiformats/go-multiaddr"

	kbucket "github.com/libp2p/go-libp2p-kbucket"
	peer "github.com/libp2p/go-libp2p-peer"
	peerstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
)

type RouteTable struct {
	peerStore                   peerstore.Peerstore
	routeTable                  *kbucket.RoutingTable
	MaxNearestPeersCountForSync int
}

func NewRouteTable(config *Config, id peer.ID, networkKey crypto.PrivateKey) *RouteTable {
	table := &RouteTable{
		peerStore:                   peerstore.NewPeerstore(),
		MaxNearestPeersCountForSync: config.MaxSyncNodes,
	}

	table.routeTable = kbucket.NewRoutingTable(
		config.Bucketsize,
		kbucket.ConvertPeerID(id),
		config.Latency,
		table.peerstore,
	)

	table.routeTable.Update(id)
	table.peerStore.AddPubKey(id, networkKey.GetPublic())
	table.peerStore.AddPrivKey(id, networkKey)

	return table
}

func (table *RouteTable) AddPeerInfo(pidStr string, addrStr []string) error {
	pid := pidStr.(peer.ID)
	if table.routeTable.Find(pid) != "" {
		return nil
	}

	var err error

	addrs := make([]ma.Multiaddr, len(addrStr))
	for i, v := range addrStr {
		addrs[i], err = multiaddr.NewMultiaddr(v)
		if err != nil {
			return err
		}
	}

	table.peerStore.AddAddrs(pid, addrs)
	table.routeTable.Update(pid)

	return nil
}

func (table *RouteTable) AddPeer(pid peer.ID, addr ma.Multiaddr) {
	table.peerStore.AddAddr(
		pid,
		addr,
		peerstore.PermanentAddrTTL,
	)
	table.routeTable.Update(pid)
}

func (table *RouteTable) AddPeerStream(s *Stream) {
	table.peerStore.AddAddr(
		s.pid,
		s.addr,
		peerstore.PermanentAddrTTL,
	)
	table.routeTable.Update(s.pid)
}

func (table *RouteTable) RemovePeerStream(s *Stream) {
	table.peerStore.AddAddr(s.pid, s.addr, 0)
	table.routeTable.Remove(s.pid)
}

func (table *RouteTable) GetNearestPeers(pid peer.ID) []peerstore.PeerInfo {
	peers := table.routeTable.NearestPeers(kbucket.ConvertPeerID(pid), table.MaxNearestPeersCountForSync)

	ret := make([]peerstore.PeerInfo, len(peers))
	for i, v := range peers {
		ret[i] = table.peerStore.PeerInfo(v)
	}
	return ret
}
