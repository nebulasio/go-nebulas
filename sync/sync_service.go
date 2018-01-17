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

package sync

import (
	"errors"
	"sync"

	"github.com/nebulasio/go-nebulas/util/byteutils"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/net"
	"github.com/nebulasio/go-nebulas/net/p2p"
	"github.com/nebulasio/go-nebulas/sync/pb"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

var (
	ErrInvalidChainSyncMessageData     = errors.New("invalid ChainSync message data")
	ErrInvalidChainGetChunkMessageData = errors.New("invalid ChainGetChunk message data")
)

type SyncService struct {
	blockChain *core.BlockChain
	netService p2p.Manager
	chunk      *Chunk
	quitCh     chan bool
	messageCh  chan net.Message

	activeSyncTask      *SyncTask
	activeSyncTaskMutex sync.Mutex
}

// NewSyncService return new SyncService.
func NewSyncService(blockChain *core.BlockChain, netService p2p.Manager) *SyncService {
	return &SyncService{
		blockChain:     blockChain,
		netService:     netService,
		chunk:          NewChunk(blockChain),
		quitCh:         make(chan bool, 1),
		activeSyncTask: nil,
		messageCh:      make(chan net.Message, 128),
	}
}

// Start start sync service.
func (ss *SyncService) Start() {
	logging.VLog().Info("Starting Sync Service.")

	// register the network handler.
	netService := ss.netService
	netService.Register(net.NewSubscriber(ss, ss.messageCh, net.ChainSync))
	netService.Register(net.NewSubscriber(ss, ss.messageCh, net.ChainChunks))
	netService.Register(net.NewSubscriber(ss, ss.messageCh, net.ChainGetChunk))
	netService.Register(net.NewSubscriber(ss, ss.messageCh, net.ChainChunkData))

	// start loop().
	go ss.startLoop()
}

// Stop stop sync service.
func (ss *SyncService) Stop() {
	// deregister the network handler.
	netService := ss.netService
	netService.Deregister(net.NewSubscriber(ss, ss.messageCh, net.ChainSync))
	netService.Deregister(net.NewSubscriber(ss, ss.messageCh, net.ChainChunks))
	netService.Deregister(net.NewSubscriber(ss, ss.messageCh, net.ChainGetChunk))
	netService.Deregister(net.NewSubscriber(ss, ss.messageCh, net.ChainChunkData))

	ss.StopActiveSync()

	ss.quitCh <- true
}

func (ss *SyncService) StartActiveSync() bool {
	// lock.
	ss.activeSyncTaskMutex.Lock()
	defer ss.activeSyncTaskMutex.Unlock()

	if ss.IsActiveSyncing() {
		return false
	}

	ss.activeSyncTask = NewSyncTask(ss.blockChain, ss.netService, ss.chunk)
	ss.activeSyncTask.Start()

	logging.CLog().WithFields(logrus.Fields{
		"syncpoint": ss.activeSyncTask.syncPointBlock,
	}).Info("Started ActiveSyncTask.")
	return true
}

func (ss *SyncService) StopActiveSync() {
	if ss.activeSyncTask == nil {
		return
	}

	ss.activeSyncTask.Stop()
	ss.activeSyncTask = nil
}

func (ss *SyncService) IsActiveSyncing() bool {
	if ss.activeSyncTask == nil {
		return false
	}

	return true
}

func (ss *SyncService) WaitingForFinish() error {
	if ss.activeSyncTask == nil {
		return nil
	}

	err := <-ss.activeSyncTask.statusCh

	logging.CLog().WithFields(logrus.Fields{
		"tail": ss.blockChain.TailBlock(),
	}).Info("ActiveSyncTask Finished.")

	ss.activeSyncTask = nil
	return err
}

func (ss *SyncService) startLoop() {
	logging.CLog().Info("Started Sync Service.")
	for {
		select {
		case <-ss.quitCh:
			if ss.activeSyncTask != nil {
				ss.activeSyncTask.Stop()
			}
			logging.CLog().Info("Stopped Sync Service.")
			return
		case message := <-ss.messageCh:
			switch message.MessageType() {
			case net.ChainSync:
				ss.onChainSync(message)
			case net.ChainChunks:
				ss.onChainChunks(message)
			case net.ChainGetChunk:
				ss.onChainGetChunk(message)
			case net.ChainChunkData:
				ss.onChainChunkData(message)
			default:
				logging.VLog().WithFields(logrus.Fields{
					"messageName": message.MessageType(),
				}).Debug("Received unknown message.")
			}
		}
	}
}

func (ss *SyncService) onChainSync(message net.Message) {
	// handle ChainSync message.
	chunkSync := new(syncpb.Sync)
	err := proto.Unmarshal(message.Data().([]byte), chunkSync)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
			"pid": message.MessageFrom(),
		}).Debug("Invalid ChainSync message data.")
		ss.netService.ClosePeer(message.MessageFrom(), ErrInvalidChainSyncMessageData)
		return
	}

	// generate Chunks message.
	chunks, err := ss.chunk.generateChunkHeaders(chunkSync.TailBlockHash)
	if err != nil && err != ErrTooSmallGapToSync {
		logging.VLog().WithFields(logrus.Fields{
			"err":  err,
			"pid":  message.MessageFrom(),
			"hash": byteutils.Hex(chunkSync.TailBlockHash),
		}).Debug("Failed to generate chunk headers.")
		return
	}

	ss.sendChainChunks(message.MessageFrom(), chunks)
}

func (ss *SyncService) onChainChunks(message net.Message) {
	if ss.activeSyncTask == nil {
		return
	}

	ss.activeSyncTask.processChunkHeaders(message)
}

func (ss *SyncService) onChainGetChunk(message net.Message) {
	// handle ChainGetChunk message.
	chunkHeader := new(syncpb.ChunkHeader)
	err := proto.Unmarshal(message.Data().([]byte), chunkHeader)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
			"pid": message.MessageFrom(),
		}).Debug("Invalid ChainGetChunk message data.")
		ss.netService.ClosePeer(message.MessageFrom(), ErrInvalidChainGetChunkMessageData)
		return
	}

	chunkData, err := ss.chunk.generateChunkData(chunkHeader)
	if err != nil {
		if err == ErrWrongChunkHeaderRootHash {
			ss.netService.ClosePeer(message.MessageFrom(), err)
		}
		return
	}

	ss.sendChainChunkData(message.MessageFrom(), chunkData)
}

func (ss *SyncService) onChainChunkData(message net.Message) {
	if ss.activeSyncTask == nil {
		return
	}

	ss.activeSyncTask.processChunkData(message)
}

func (ss *SyncService) sendChainChunks(peerID string, chunks *syncpb.ChunkHeaders) {
	data, err := proto.Marshal(chunks)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Debug("Failed to marshal syncpb.ChunkHeaders.")
		return
	}

	ss.netService.SendMessageToPeer(net.ChainChunks, data, net.MessagePriorityLow, peerID)
}

func (ss *SyncService) sendChainChunkData(peerID string, chunkData *syncpb.ChunkData) {
	data, err := proto.Marshal(chunkData)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Debug("Failed to marshal syncpb.ChunkData.")
		return
	}

	ss.netService.SendMessageToPeer(net.ChainChunkData, data, net.MessagePriorityLow, peerID)
}
