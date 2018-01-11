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
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"path"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/multiformats/go-multiaddr"
	"github.com/nebulasio/go-nebulas/util/logging"

	kbucket "github.com/libp2p/go-libp2p-kbucket"
	peer "github.com/libp2p/go-libp2p-peer"
	peerstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
)

type RouteTable struct {
	quitCh                   chan bool
	peerStore                peerstore.Peerstore
	routeTable               *kbucket.RoutingTable
	maxPeersCountForSyncResp int
	maxPeersCountToSync      int
	cacheFilePath            string
	seedNodes                []ma.Multiaddr
	node                     *Node
	streamManager            *StreamManager
	latestUpdatedAt          int64
}

func NewRouteTable(config *Config, node *Node) *RouteTable {
	table := &RouteTable{
		quitCh:                   make(chan bool, 1),
		peerStore:                peerstore.NewPeerstore(),
		maxPeersCountForSyncResp: MaxPeersCountForSyncResp,
		maxPeersCountToSync:      config.MaxSyncNodes,
		cacheFilePath:            path.Join(config.RoutingTableDir, RouteTableCacheFileName),
		seedNodes:                config.BootNodes,
		node:                     node,
		streamManager:            node.streamManager,
		latestUpdatedAt:          0,
	}

	table.routeTable = kbucket.NewRoutingTable(
		config.Bucketsize,
		kbucket.ConvertPeerID(node.id),
		config.Latency,
		table.peerStore,
	)

	table.routeTable.Update(node.id)
	table.peerStore.AddPubKey(node.id, node.networkKey.GetPublic())
	table.peerStore.AddPrivKey(node.id, node.networkKey)

	return table
}

func (table *RouteTable) Start() {
	go table.syncLoop()
}

func (table *RouteTable) Stop() {
	table.quitCh <- true
}

func (table *RouteTable) Peers() map[peer.ID][]ma.Multiaddr {
	peers := make(map[peer.ID][]ma.Multiaddr)
	for _, pid := range table.peerStore.Peers() {
		peers[pid] = table.peerStore.Addrs(pid)
	}
	return peers
}

func (table *RouteTable) syncLoop() {
	// Load Route Table.
	table.LoadSeedNodes()
	table.LoadRouteTableFromFile()

	// trigger first sync.
	table.SyncRouteTable()

	syncLoopTicker := time.NewTicker(RouteTableSyncLoopInterval)
	saveRouteTableToDiskTicker := time.NewTicker(RouteTableSaveToDiskInterval)
	latestUpdatedAt := table.latestUpdatedAt

	for {
		select {
		case <-table.quitCh:
			logging.CLog().Info("Stopping Route Table Sync Loop.")
			return
		case <-syncLoopTicker.C:
			table.SyncRouteTable()
		case <-saveRouteTableToDiskTicker.C:
			if latestUpdatedAt < table.latestUpdatedAt {
				table.SaveRouteTableToFile()
			}
		}
	}
}

func (table *RouteTable) AddPeerInfo(pidStr string, addrStr []string) error {
	pid := peer.ID(pidStr)
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

	table.peerStore.AddAddrs(pid, addrs, peerstore.PermanentAddrTTL)
	table.routeTable.Update(pid)

	return nil
}

func (table *RouteTable) AddPeer(pid peer.ID, addr ma.Multiaddr) {
	logging.CLog().Infof("Adding: %s,%s", pid.Pretty(), addr.String())
	table.peerStore.AddAddr(pid, addr, peerstore.PermanentAddrTTL)
	table.routeTable.Update(pid)
}

func (table *RouteTable) AddPeerAddr(addr ma.Multiaddr) {
	id, err := MultiaddrToPeerID(addr)
	if err != nil {
		return
	}
	table.AddPeer(id, addr)
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
	peers := table.routeTable.NearestPeers(kbucket.ConvertPeerID(pid), table.maxPeersCountForSyncResp)

	ret := make([]peerstore.PeerInfo, len(peers))
	for i, v := range peers {
		ret[i] = table.peerStore.PeerInfo(v)
	}
	return ret
}

func (table *RouteTable) LoadSeedNodes() {
	for _, addr := range table.seedNodes {
		table.AddPeerAddr(addr)
	}
}

func (table *RouteTable) LoadRouteTableFromFile() {
	file, err := os.Open(table.cacheFilePath)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"cacheFilePath": table.cacheFilePath,
			"err":           err,
		}).Warn("Failed to open Route Table Cache file.")
		return
	}
	defer file.Close()

	// read line by line.
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") {
			continue
		}

		addr, err := ma.NewMultiaddr(line)
		if err != nil {
			// ignore.
			logging.VLog().WithFields(logrus.Fields{
				"err":  err,
				"text": line,
			}).Warn("Invalid address in Route Table Cache file.")
			continue
		}

		table.AddPeerAddr(addr)
	}
}

func (table *RouteTable) SaveRouteTableToFile() {
	file, err := os.Create(table.cacheFilePath)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"cacheFilePath": table.cacheFilePath,
			"err":           err,
		}).Warn("Failed to open Route Table Cache file.")
		return
	}
	defer file.Close()

	// write header.
	file.WriteString(fmt.Sprintf("# %s\n", time.Now().String()))

	peers := table.routeTable.ListPeers()
	for _, v := range peers {
		for _, addr := range table.peerStore.Addrs(v) {
			file.WriteString(addr.String())
		}
	}
}

func (table *RouteTable) SyncRouteTable() {
	// sync with seed nodes.
	for _, addr := range table.seedNodes {
		pid, err := MultiaddrToPeerID(addr)
		if err != nil {
			continue
		}
		table.SyncWithPeer(pid)
	}

	// random peer selection.
	rand.Seed(time.Now().UnixNano())
	selectedPeersCount := table.maxPeersCountToSync
	peers := table.routeTable.ListPeers()
	peersCount := len(peers)

	selectedPeersIdx := make(map[int]bool)
	for i := 0; i < selectedPeersCount; i++ {
		ri := 0
		for {
			ri := rand.Intn(peersCount)
			if selectedPeersIdx[ri] == false {
				break
			}
		}
		selectedPeersIdx[ri] = true
		pid := peers[ri]

		table.SyncWithPeer(pid)
	}
}

func (table *RouteTable) SyncWithPeer(pid peer.ID) {
	stream := table.streamManager.Find(pid)

	if stream == nil {
		stream = NewStreamFromPID(pid, table.node)
		table.streamManager.AddStream(stream)
	}

	stream.SyncRoute()
}
