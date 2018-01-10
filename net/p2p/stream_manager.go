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
	"sync"
	"time"

	libnet "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	"github.com/nebulasio/go-nebulas/util/logging"
)

type StreamManager struct {
	quitCh     chan bool
	allStreams *sync.Map
}

func NewStreamManager() *StreamManager {
	return &StreamManager{
		quitCh:     make(chan bool, 1),
		allStreams: new(sync.Map),
	}
}

func (sm *StreamManager) Start() {
	go sm.loop()
}

func (sm *StreamManager) Stop() {
	sm.quitCh <- true
}

func (sm *StreamManager) Add(s libnet.Stream, node *Node) {
	stream := NewStream(s.Conn().RemotePeer(), s, node)
	sm.AddStream(stream)
}

func (sm *StreamManager) AddStream(stream *Stream) {
	sm.allStreams.Store(stream.pid, stream)
	stream.StartLoop()
}

func (sm *StreamManager) Remove(pid peer.ID) {
	sm.allStreams.Delete(pid)
}

func (sm *StreamManager) RemoveStream(s *Stream) {
	sm.Remove(s.pid)
}

func (sm *StreamManager) Find(pid peer.ID) *Stream {
	return sm.allStreams.Load(pid)
}

func (sm *StreamManager) loop() {
	logging.CLog().Info("Starting Stream Manager Loop.")

	ticker := time.NewTicker(time.Second * 5)
	for {
		select {
		case <-sm.quitCh:
			logging.CLog().Info("Stopping Stream Manager Loop.")
			return
		case <-ticker.C:
			// TODO: @robin cleanup connections.
			logging.VLog().Warn("TODO: cleanup connections is not implemented.")
		}
	}
}
