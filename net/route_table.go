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

package net

import (
	"bufio"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/multiformats/go-multiaddr"
	netpb "github.com/nebulasio/go-nebulas/net/pb"
	"github.com/nebulasio/go-nebulas/util/logging"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	kbucket "github.com/libp2p/go-libp2p-kbucket"
	"github.com/libp2p/go-libp2p-peerstore/pstoremem"
	ma "github.com/multiformats/go-multiaddr"
)

// Route Table Errors
var (
	ErrExceedMaxSyncRouteResponse = errors.New("too many sync route table response")
)

// RouteTable route table struct.
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
	internalNodeList         []string
}

// NewRouteTable new route table.
func NewRouteTable(config *Config, node *Node) *RouteTable {
	table := &RouteTable{
		quitCh:                   make(chan bool, 1),
		peerStore:                pstoremem.NewPeerstore(),
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

// Start start route table syncLoop.
func (table *RouteTable) Start() {
	logging.CLog().Info("Starting NebService RouteTable Sync...")

	go table.syncLoop()
}

// Stop quit route table syncLoop.
func (table *RouteTable) Stop() {
	logging.CLog().Info("Stopping NebService RouteTable Sync...")

	table.quitCh <- true
}

// Peers return peers in route table.
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
	table.LoadInternalNodeList()

	// trigger first sync.
	table.SyncRouteTable()

	logging.CLog().Info("Started NebService RouteTable Sync.")

	syncLoopTicker := time.NewTicker(RouteTableSyncLoopInterval)
	saveRouteTableToDiskTicker := time.NewTicker(RouteTableSaveToDiskInterval)
	latestUpdatedAt := table.latestUpdatedAt

	for {
		select {
		case <-table.quitCh:
			logging.CLog().Info("Stopped NebService RouteTable Sync.")
			return
		case <-syncLoopTicker.C:
			table.SyncRouteTable()
		case <-saveRouteTableToDiskTicker.C:
			if latestUpdatedAt < table.latestUpdatedAt {
				table.SaveRouteTableToFile()
				latestUpdatedAt = table.latestUpdatedAt
			}
		}
	}
}

// AddPeerInfo add peer to route table.
func (table *RouteTable) AddPeerInfo(prettyID string, addrStr []string) error {
	pid, err := peer.IDB58Decode(prettyID)
	if err != nil {
		return nil
	}

	addrs := make([]ma.Multiaddr, len(addrStr))
	for i, v := range addrStr {
		addrs[i], err = multiaddr.NewMultiaddr(v)
		if err != nil {
			return err
		}
	}

	if table.routeTable.Find(pid) != "" {
		table.peerStore.SetAddrs(pid, addrs, peerstore.PermanentAddrTTL)
	} else {
		table.peerStore.AddAddrs(pid, addrs, peerstore.PermanentAddrTTL)
	}
	table.routeTable.Update(pid)
	table.onRouteTableChange()

	return nil
}

// AddPeer add peer to route table.
func (table *RouteTable) AddPeer(pid peer.ID, addr ma.Multiaddr) {
	logging.VLog().Debugf("Adding Peer: %s,%s", pid.Pretty(), addr.String())
	table.peerStore.AddAddr(pid, addr, peerstore.PermanentAddrTTL)
	table.routeTable.Update(pid)
	table.onRouteTableChange()

}

// AddPeers add peers to route table
func (table *RouteTable) AddPeers(pid string, peers *netpb.Peers) {
	// recv too many peers info. say Bye.
	if len(peers.Peers) > table.maxPeersCountForSyncResp {
		table.streamManager.CloseStream(pid, ErrExceedMaxSyncRouteResponse)
	}
	for _, v := range peers.Peers {
		table.AddPeerInfo(v.Id, v.Addrs)
	}
}

// AddIPFSPeerAddr add a peer to route table with ipfs address.
func (table *RouteTable) AddIPFSPeerAddr(addr ma.Multiaddr) {
	id, addr, err := ParseFromIPFSAddr(addr)
	if err != nil {
		return
	}
	table.AddPeer(id, addr)
}

// AddPeerStream add peer stream to peerStore.
func (table *RouteTable) AddPeerStream(s *Stream) {
	table.peerStore.AddAddr(
		s.pid,
		s.addr,
		peerstore.PermanentAddrTTL,
	)
	table.routeTable.Update(s.pid)
	table.onRouteTableChange()
}

// RemovePeerStream remove peerStream from peerStore.
func (table *RouteTable) RemovePeerStream(s *Stream) {
	table.peerStore.ClearAddrs(s.pid)
	table.routeTable.Remove(s.pid)
	table.onRouteTableChange()
}

func (table *RouteTable) onRouteTableChange() {
	table.latestUpdatedAt = time.Now().Unix()
}

// GetRandomPeers get random peers
func (table *RouteTable) GetRandomPeers(pid peer.ID) []peer.AddrInfo {

	// change sync route algorithm from `NearestPeers` to `randomPeers`
	var peers []peer.ID
	allPeers := table.routeTable.ListPeers()
	// Do not accept internal node synchronization routing requests.
	if inArray(pid.Pretty(), table.internalNodeList) {
		return []peer.AddrInfo{}
	}

	for _, v := range allPeers {
		if inArray(v.Pretty(), table.internalNodeList) == false {
			peers = append(peers, v)
		}
	}
	peers = shufflePeerID(peers)
	if len(peers) > table.maxPeersCountForSyncResp {
		peers = peers[:table.maxPeersCountForSyncResp]
	}
	ret := make([]peer.AddrInfo, len(peers))
	for i, v := range peers {
		ret[i] = table.peerStore.PeerInfo(v)
	}
	return ret
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

func shufflePeerID(pids []peer.ID) []peer.ID {

	r := rand.New(rand.NewSource(time.Now().Unix()))
	ret := make([]peer.ID, len(pids))
	perm := r.Perm(len(pids))
	for i, randIndex := range perm {
		ret[i] = pids[randIndex]
	}
	return ret
}

// LoadSeedNodes load seed nodes.
func (table *RouteTable) LoadSeedNodes() {
	for _, ipfsAddr := range table.seedNodes {
		table.AddIPFSPeerAddr(ipfsAddr)
	}
}

// LoadRouteTableFromFile load route table from file.
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

		table.AddIPFSPeerAddr(addr)
	}
}

// SaveRouteTableToFile save route table to file.
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
			line := fmt.Sprintf("%s/ipfs/%s\n", addr, v.Pretty())
			file.WriteString(line)
		}
	}
}

// SyncRouteTable sync route table.
func (table *RouteTable) SyncRouteTable() {
	syncedPeers := make(map[peer.ID]bool)

	// sync with seed nodes.
	for _, ipfsAddr := range table.seedNodes {
		pid, _, err := ParseFromIPFSAddr(ipfsAddr)
		if err != nil {
			continue
		}
		table.SyncWithPeer(pid)
		syncedPeers[pid] = true
	}

	// random peer selection.
	peers := table.routeTable.ListPeers()
	peersCount := len(peers)
	if peersCount <= 1 {
		return
	}

	peersCountToSync := table.maxPeersCountToSync

	if peersCount < peersCountToSync {
		peersCountToSync = peersCount
	}
	selectedPeersIdx := make(map[int]bool)
	for i := 0; i < peersCountToSync/2; i++ {
		ri := 0

		for {
			ri = rand.Intn(peersCountToSync)
			if selectedPeersIdx[ri] == false {
				break
			}
		}

		selectedPeersIdx[ri] = true
		pid := peers[ri]

		if syncedPeers[pid] == false {
			table.SyncWithPeer(pid)
			syncedPeers[pid] = true
		}
	}
}

// SyncWithPeer sync route table with a peer.
func (table *RouteTable) SyncWithPeer(pid peer.ID) {
	if pid == table.node.id {
		return
	}

	stream := table.streamManager.Find(pid)

	if stream == nil {
		stream = NewStreamFromPID(pid, table.node)
		table.streamManager.AddStream(stream)
	}

	stream.SyncRoute()
}

//LoadInternalNodeList Load Internal Node list from file
func (table *RouteTable) LoadInternalNodeList() {
	file, err := os.Open(RouteTableInternalNodeFileName)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Warn("Failed to open internal list file.")
		return
	}
	defer file.Close()

	// read line by line.
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) > 0 {
			table.internalNodeList = append(table.internalNodeList, line)
		}
	}

	logging.VLog().WithFields(logrus.Fields{
		"internalNodeList": table.internalNodeList,
	}).Info("Loaded internal node list.")
}
