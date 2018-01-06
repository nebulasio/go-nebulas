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
	"time"

	libnet "github.com/libp2p/go-libp2p-net"
)

// StreamStore is for stream cache
type StreamStore struct {
	key       string
	conn      int
	stream    libnet.Stream
	timestamp int64
}

func streamEliminationAlgorithm(a interface{}, b interface{}) bool {
	sa := a.(*StreamStore)
	sb := b.(*StreamStore)
	return sa.timestamp < sb.timestamp
}

// NewStreamStore return a new streamStore
func NewStreamStore(key string, conn int, stream libnet.Stream) *StreamStore {
	return &StreamStore{key, conn, stream, time.Now().Unix()}
}

func (node *Node) manageStreamStore() {
	second := 30 * time.Second
	ticker := time.NewTicker(second)
	for {
		select {
		case <-ticker.C:
			node.clearStreamStore()
			node.cleanPeerStore()
		case <-node.ns.quitCh:
			return
		}
	}
}

func (node *Node) cleanPeerStore() {
	for _, v := range node.peerstore.Peers() {
		if _, ok := node.stream.Load(v.Pretty()); !ok {
			if !InArray(v.Pretty(), node.bootIds) {
				node.peerstore.ClearAddrs(v)
			}
		}
	}
}

func (node *Node) clearStreamStore() {
	// do clear streamStore only when the count of stream in cache exceed the cache size.
	if node.streamCache.Len() > node.config.StreamStoreSize {
		overflowSize := node.streamCache.Len() - node.config.StreamStoreSize
		for i := 0; i < overflowSize; i++ {
			streamStore := node.streamCache.PopMin().(*StreamStore)
			key := streamStore.key

			if streamStore, ok := node.stream.Load(key); ok {
				streamStore.(*StreamStore).stream.Close()
				node.stream.Delete(key)
			}
		}
	}
}
