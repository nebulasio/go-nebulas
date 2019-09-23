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
	"errors"
	"fmt"
	"hash/crc32"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/gogo/protobuf/proto"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/nebulasio/go-nebulas/util/logging"
)

// const
const (
	CleanupInterval = time.Second * 60
	// MaxStreamNum      = 500
	// ReservedStreamNum = 50 // of MaxStreamNum
)

// var
var (
	ErrExceedMaxStreamNum = errors.New("too many streams connected")
	ErrElimination        = errors.New("eliminated for low value")
	ErrDeprecatedStream   = errors.New("deprecated stream")
)

// StreamManager manages all streams
type StreamManager struct {
	mu                sync.Mutex
	quitCh            chan bool
	allStreams        *sync.Map
	activePeersCount  int32
	maxStreamNum      int32
	reservedStreamNum int32
}

// NewStreamManager return a new stream manager
func NewStreamManager(config *Config) *StreamManager {
	return &StreamManager{
		quitCh:            make(chan bool, 1),
		allStreams:        new(sync.Map),
		activePeersCount:  0,
		maxStreamNum:      config.StreamLimits,
		reservedStreamNum: config.ReservedStreamLimits,
	}
}

// Count return active peers count in the stream manager
func (sm *StreamManager) Count() int32 {
	return sm.activePeersCount
}

// Start stream manager service
func (sm *StreamManager) Start() {
	logging.CLog().Info("Starting NebService StreamManager...")

	go sm.loop()
}

// Stop stream manager service
func (sm *StreamManager) Stop() {
	logging.CLog().Info("Stopping NebService StreamManager...")

	sm.quitCh <- true
}

// Add a new stream into the stream manager
func (sm *StreamManager) Add(s network.Stream, node *Node) {
	stream := NewStream(s, node)
	sm.AddStream(stream)
}

// AddStream into the stream manager
func (sm *StreamManager) AddStream(stream *Stream) {

	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.activePeersCount >= sm.maxStreamNum {
		if stream.stream != nil {
			stream.stream.Close()
		}
		return
	}

	// check & close old stream
	if v, ok := sm.allStreams.Load(stream.pid.Pretty()); ok {
		old, _ := v.(*Stream)

		logging.VLog().WithFields(logrus.Fields{
			"pid": old.pid.Pretty(),
		}).Debug("Removing old stream.")

		sm.activePeersCount--
		sm.allStreams.Delete(old.pid.Pretty())

		if old.stream != nil {
			old.stream.Close()
		}
	}

	logging.VLog().WithFields(logrus.Fields{
		"stream": stream.String(),
	}).Debug("Added a new stream.")

	sm.activePeersCount++
	sm.allStreams.Store(stream.pid.Pretty(), stream)
	stream.StartLoop()
}

// Remove the stream with the given pid from the stream manager
// func (sm *StreamManager) Remove(pid peer.ID) {

// 	sm.mu.Lock()
// 	defer sm.mu.Unlock()

// 	logging.VLog().WithFields(logrus.Fields{
// 		"pid": pid.Pretty(),
// 	}).Debug("Removing a stream.")

// 	if _, ok := sm.allStreams.Load(pid.Pretty()); !ok {
// 		// caused by close in AddStream
// 		return
// 	}

// 	sm.activePeersCount--
// 	sm.allStreams.Delete(pid.Pretty())
// }

// RemoveStream from the stream manager
func (sm *StreamManager) RemoveStream(s *Stream) {

	sm.mu.Lock()
	defer sm.mu.Unlock()

	v, ok := sm.allStreams.Load(s.pid.Pretty())
	if !ok {
		return
	}

	exist, _ := v.(*Stream)
	if s != exist {
		return
	}

	logging.VLog().WithFields(logrus.Fields{
		"pid": s.pid.Pretty(),
	}).Debug("Removing a stream.")

	sm.activePeersCount--
	sm.allStreams.Delete(s.pid.Pretty())
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
	logging.CLog().Info("Started NebService StreamManager.")

	ticker := time.NewTicker(CleanupInterval)
	for {
		select {
		case <-sm.quitCh:
			logging.CLog().Info("Stopped Stream Manager Loop.")
			return
		case <-ticker.C:
			sm.cleanup()
		}
	}
}

// BroadcastMessage broadcast the message
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
		stream.close(reason)
	}
}

// cleanup eliminating low value streams if reaching the limit
func (sm *StreamManager) cleanup() {

	if sm.activePeersCount < sm.maxStreamNum {
		logging.VLog().WithFields(logrus.Fields{
			"maxNum":      sm.maxStreamNum,
			"reservedNum": sm.reservedStreamNum,
			"currentNum":  sm.activePeersCount,
		}).Debug("No need for streams cleanup.")
		return
	}

	// total number of each msg type
	msgTotal := make(map[string]int)

	// weight of each msg type
	msgWeight := make(map[string]MessageWeight)
	msgWeight[ROUTETABLE] = MessageWeightRouteTable

	svs := make(StreamValueSlice, 0)

	sm.allStreams.Range(func(key, value interface{}) bool {
		stream := value.(*Stream)

		// t type, c count
		for t, c := range stream.msgCount {
			msgTotal[t] += c
			if _, ok := msgWeight[t]; ok {
				continue
			}

			v, _ := stream.node.netService.dispatcher.subscribersMap.Load(t)
			if m, ok := v.(*sync.Map); ok {
				m.Range(func(key, value interface{}) bool {
					msgWeight[t] = key.(*Subscriber).MessageWeight()
					return false
				})
			}
		}

		svs = append(svs, &StreamValue{
			stream: stream,
		})

		return true
	})

	// check length
	if len(svs) <= int(sm.maxStreamNum-sm.reservedStreamNum) {
		logging.CLog().WithFields(logrus.Fields{
			"streamValueSliceLength": len(svs),
		}).Debug("StreamValueSlice length is not enough, return directly.")
		return
	}

	for _, sv := range svs {
		for t, c := range sv.stream.msgCount {
			w, _ := msgWeight[t]
			sv.value += float64(c) * float64(w) / float64(msgTotal[t])
		}
	}

	sort.Sort(sort.Reverse(svs))
	logging.VLog().WithFields(logrus.Fields{
		"maxNum":           sm.maxStreamNum,
		"reservedNum":      sm.reservedStreamNum,
		"currentNum":       sm.activePeersCount,
		"msgTotal":         msgTotal,
		"msgWeight":        msgWeight,
		"streamValueSlice": svs,
	}).Debug("Sorting streams before the cleanup.")

	eliminated := svs[sm.maxStreamNum-sm.reservedStreamNum:]
	for _, sv := range eliminated {
		sv.stream.close(ErrElimination)
	}

	svs = svs[:sm.maxStreamNum-sm.reservedStreamNum]
	logging.VLog().WithFields(logrus.Fields{
		"eliminatedNum": len(eliminated),
		"retained":      svs,
	}).Debug("Streams cleanup is done.")
}

// StreamValue value of stream in the past CleanupInterval
type StreamValue struct {
	stream *Stream
	value  float64
}

// StreamValueSlice StreamValue slice
type StreamValueSlice []*StreamValue

func (s StreamValueSlice) Len() int           { return len(s) }
func (s StreamValueSlice) Less(i, j int) bool { return s[i].value < s[j].value }
func (s StreamValueSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s *StreamValue) String() string {
	return s.stream.addr.String() + ":" +
		strconv.FormatFloat(s.value, 'f', 3, 64) + ":" +
		fmt.Sprintf("%v", s.stream.msgCount)
}
