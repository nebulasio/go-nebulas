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
	"hash/crc32"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/gogo/protobuf/proto"
	libnet "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	"github.com/nebulasio/go-nebulas/util/logging"
)

// StreamManager manages all streams
type StreamManager struct {
	quitCh           chan bool
	allStreams       *sync.Map
	activePeersCount int32
}

// NewStreamManager return a new stream manager
func NewStreamManager() *StreamManager {
	return &StreamManager{
		quitCh:           make(chan bool, 1),
		allStreams:       new(sync.Map),
		activePeersCount: 0,
	}
}

// Count return active peers count in the stream manager
func (sm *StreamManager) Count() int32 {
	return sm.activePeersCount
}

// Start stream manager service
func (sm *StreamManager) Start() {
	logging.CLog().Info("Starting NetService StreamManager...")

	go sm.loop()
}

// Stop stream manager service
func (sm *StreamManager) Stop() {
	logging.CLog().Info("Stopping NetService StreamManager...")

	sm.quitCh <- true
}

// Add a new stream into the stream manager
func (sm *StreamManager) Add(s libnet.Stream, node *Node) {
	stream := NewStream(s, node)
	sm.AddStream(stream)
}

// AddStream into the stream manager
func (sm *StreamManager) AddStream(stream *Stream) {
	logging.VLog().WithFields(logrus.Fields{
		"steam": stream.String(),
	}).Debug("Added a new stream.")

	atomic.AddInt32(&sm.activePeersCount, 1)
	sm.allStreams.Store(stream.pid.Pretty(), stream)
	stream.StartLoop()
}

// Remove the stream with the given pid from the stream manager
func (sm *StreamManager) Remove(pid peer.ID) {
	logging.VLog().WithFields(logrus.Fields{
		"pid": pid.Pretty(),
	}).Debug("Removing a stream.")

	atomic.AddInt32(&sm.activePeersCount, -1)
	sm.allStreams.Delete(pid.Pretty())
}

// RemoveStream from the stream manager
func (sm *StreamManager) RemoveStream(s *Stream) {
	sm.Remove(s.pid)
}

// FindByPeerID find the stream with the given peerID
func (sm *StreamManager) FindByPeerID(peerID string) *Stream {
	v, _ := sm.allStreams.Load(peerID)
	if v == nil {
		return nil
	}
	return v.(*Stream)
}

// Find the stream with the given pid
func (sm *StreamManager) Find(pid peer.ID) *Stream {
	return sm.FindByPeerID(pid.Pretty())
}

func (sm *StreamManager) loop() {
	logging.CLog().Info("Started NetService StreamManager.")

	ticker := time.NewTicker(time.Second * 30)
	for {
		select {
		case <-sm.quitCh:
			logging.CLog().Info("Stopped Stream Manager Loop.")
			return
		case <-ticker.C:
			// TODO: @robin streams cleanup if needed.
			logging.CLog().Info("TODO: streams cleanup is not implemented.")
		}
	}
}

// BroadcastMessage broadcast the meesage
func (sm *StreamManager) BroadcastMessage(messageName string, messageContent Serializable, priority int) {
	pb, _ := messageContent.ToProto()
	data, err := proto.Marshal(pb)
	if err != nil {
		return
	}

	dataCheckSum := crc32.ChecksumIEEE(data)

	sm.allStreams.Range(func(key, value interface{}) bool {
		stream := value.(*Stream)
		if stream.IsHandshakeSucceed() && !HasRecvMessage(stream, dataCheckSum) {
			stream.SendMessage(messageName, data, priority)
		}
		return true
	})
}

// RelayMessage relay the message
func (sm *StreamManager) RelayMessage(messageName string, messageContent Serializable, priority int) {
	pb, _ := messageContent.ToProto()
	data, err := proto.Marshal(pb)
	if err != nil {
		return
	}

	dataCheckSum := crc32.ChecksumIEEE(data)

	sm.allStreams.Range(func(key, value interface{}) bool {
		stream := value.(*Stream)
		if stream.IsHandshakeSucceed() && !HasRecvMessage(stream, dataCheckSum) {
			stream.SendMessage(messageName, data, priority)
		}
		return true
	})
}

// SendMessageToPeers send the message to the peers filtered by the filter algorithm
func (sm *StreamManager) SendMessageToPeers(messageName string, data []byte, priority int, filter PeerFilterAlgorithm) []string {
	allPeers := make(PeersSlice, 0)

	sm.allStreams.Range(func(key, value interface{}) bool {
		stream := value.(*Stream)
		if stream.IsHandshakeSucceed() {
			allPeers = append(allPeers, value)
		}
		return true
	})

	selectedPeers := filter.Filter(allPeers)
	selectedPeersPrettyID := make([]string, 0)

	for _, v := range selectedPeers {
		stream := v.(*Stream)
		if err := stream.SendMessage(messageName, data, priority); err == nil {
			selectedPeersPrettyID = append(selectedPeersPrettyID, stream.pid.Pretty())
		}
	}

	return selectedPeersPrettyID
}

// CloseStream with the given pid and reason
func (sm *StreamManager) CloseStream(peerID string, reason error) {
	stream := sm.FindByPeerID(peerID)
	if stream != nil {
		stream.Close(reason)
	}
}
