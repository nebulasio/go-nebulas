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

	libnet "github.com/libp2p/go-libp2p-net"
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

}

func (sm *StreamManager) Stop() {

}

func (sm *StreamManager) Add(node *Node, s libnet.Stream) {
	sID := s.Conn().RemotePeer()
	stream := NewStream(sID, s, node)

	sm.allStreams.Store(sID, stream)
	stream.StartLoop()
}

func (sm *StreamManager) Remove(s *Stream) {
	sID := s.pid
	sm.allStreams.Delete(sID)
}
