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
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	syncpb "github.com/nebulasio/go-nebulas/sync/pb"
	"github.com/nebulasio/go-nebulas/util/byteutils"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/net"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

const (
	chunkDataStatusFinished = int64(-1)
	chunkDataStatusNotStart = int64(0)
)

// Errors
var (
	ErrInvalidChainChunksMessageData    = errors.New("invalid ChainChunks message data")
	ErrWrongChainChunksMessageData      = errors.New("wrong ChainChunks message data")
	ErrInvalidChainChunkDataMessageData = errors.New("invalid ChainChunkData message data")
	ErrWrongChainChunkDataMessageData   = errors.New("wrong ChainChunkData message data")
	ErrInvalidChunkHeaderSourcePeer     = errors.New("invalid chunk headers source peer")
)

// Task is a sync task
type Task struct {
	quitCh                                  chan bool
	statusCh                                chan bool
	blockChain                              *core.BlockChain
	syncPointBlock                          *core.Block
	netService                              net.Service
	chunk                                   *Chunk
	syncMutex                               sync.Mutex
	chainSyncPeers                          []string
	maxConsistentChunkHeadersCount          int
	maxConsistentChunkHeaders               *syncpb.ChunkHeaders
	maxConsistentChunkHeadersChainSyncPeers map[string][]string
	chunkHeadersRootHashCounter             map[string]int
	receivedChunkHeadersRootHashPeers       map[string]bool

	chainSyncDoneCh               chan bool
	chainChunkDataSyncPosition    int
	chainChunkDataProcessPosition int
	chainChunkData                map[int]*syncpb.ChunkData
	chainChunkDataStatus          map[int]int64
	chinGetChunkDataDoneCh        chan bool

	// debug fields.
	chainSyncRetryCount int
}

// NewTask return a new sync task
func NewTask(blockChain *core.BlockChain, netService net.Service, chunk *Chunk) *Task {
	return &Task{
		quitCh:                                  make(chan bool, 1),
		statusCh:                                make(chan bool, 1),
		blockChain:                              blockChain,
		syncPointBlock:                          blockChain.LIB(),
		netService:                              netService,
		chunk:                                   chunk,
		chainSyncPeers:                          nil,
		maxConsistentChunkHeadersCount:          0,
		maxConsistentChunkHeaders:               nil,
		maxConsistentChunkHeadersChainSyncPeers: make(map[string][]string),
		chunkHeadersRootHashCounter:             make(map[string]int),
		receivedChunkHeadersRootHashPeers:       make(map[string]bool),
		chainSyncDoneCh:                         make(chan bool, 1),
		chainChunkDataSyncPosition:              0,
		chainChunkDataProcessPosition:           0,
		chainChunkData:                          make(map[int]*syncpb.ChunkData),
		chainChunkDataStatus:                    make(map[int]int64),
		chinGetChunkDataDoneCh:                  make(chan bool, 1),
		// debug fields.
		chainSyncRetryCount: 0,
	}
}

// Start the sync task
func (st *Task) Start() {
	go st.startSyncLoop()
}

// Stop the sync task
func (st *Task) Stop() {
	st.quitCh <- true
}

func (st *Task) startSyncLoop() {
	for {
		// start chain sync.
		st.chunkHeadersRequest()

		syncTicker := time.NewTicker(10 * time.Second)

	SYNC_STEP_1:
		for {
			select {
			case <-st.quitCh:
				logging.VLog().Info("Stopped sync loop.")
				return
			case <-syncTicker.C:
				if !st.hasEnoughChunkHeaders() {
					st.reset()
					st.setSyncPointToLastChunk()
					st.chunkHeadersRequest()
					continue
				}
			case <-st.chainSyncDoneCh:
				// go to next step.
				logging.VLog().WithFields(logrus.Fields{
					"chainSyncPeers":                    st.chainSyncPeers,
					"chainSyncRetryCount":               st.chainSyncRetryCount,
					"maxConsistentChunkHeadersCount":    st.maxConsistentChunkHeadersCount,
					"maxConsistentChunkHeadersRootHash": byteutils.Hex(st.maxConsistentChunkHeaders.Root),
					"countOfChunkHeaders":               len(st.maxConsistentChunkHeaders.ChunkHeaders),
				}).Info("ChainSync Finished. Move to GetChainData.")
				break SYNC_STEP_1
			}
		}

		// start get chunk data.
		logging.VLog().Info("Starting GetChainData from peers.")

		st.chainSyncRetryCount = 0
		st.sendChunkDataRequest()

		getChunkTimeoutTicker := time.NewTicker(10 * time.Second)

	SYNC_STEP_2:
		for {
			select {
			case <-st.quitCh:
				logging.VLog().Info("Stopped sync loop.")
				return
			case <-getChunkTimeoutTicker.C:
				if st.chainSyncRetryCount > GetChunkDataTimeout {
					logging.CLog().WithFields(logrus.Fields{
						"from":                st.syncPointBlock,
						"chainSyncPeers":      st.chainSyncPeers,
						"chainSyncRetryCount": st.chainSyncRetryCount,
					}).Warn("Get chunk data timeout. Go to next one.")
					st.reset()
					st.setSyncPointToNewTail()
					break SYNC_STEP_2
				}
				// for the timeout peer, send message again.
				st.checkChainGetChunkTimeout()
			case <-st.chinGetChunkDataDoneCh:
				// finished.
				logging.VLog().Info("GetChainData Finished.")
				if len(st.maxConsistentChunkHeaders.ChunkHeaders) == 0 {
					st.statusCh <- true
					return
				}
				logging.CLog().WithFields(logrus.Fields{
					"from": st.syncPointBlock,
					"to":   st.blockChain.TailBlock(),
				}).Info("Finish a sync subtask. Go to next one.")
				st.reset()
				st.setSyncPointToNewTail()
				break SYNC_STEP_2
			}
		}
	}
}

func (st *Task) reset() {
	st.syncMutex.Lock()
	defer st.syncMutex.Unlock()

	st.chainSyncPeers = nil
	st.maxConsistentChunkHeadersCount = 0
	st.maxConsistentChunkHeaders = nil
	st.maxConsistentChunkHeadersChainSyncPeers = make(map[string][]string)
	st.chunkHeadersRootHashCounter = make(map[string]int)
	st.receivedChunkHeadersRootHashPeers = make(map[string]bool)
	st.chainChunkDataStatus = make(map[int]int64)
	st.chainChunkDataSyncPosition = 0
	st.chainChunkDataProcessPosition = 0
	st.chainChunkData = make(map[int]*syncpb.ChunkData)
}

func (st *Task) setSyncPointToNewTail() {
	st.chainSyncRetryCount = 0
	if st.syncPointBlock.Height() > st.blockChain.TailBlock().Height() {
		st.syncPointBlock = st.blockChain.TailBlock()
	}
}

func (st *Task) setSyncPointToLastChunk() {
	if st.chainSyncRetryCount < 2 {
		// for the first retry, keep current tail.
		// TODO: for testing perpose, could be deleted.
		return
	}

	// the first block height of chunk is 1.
	lastChunkBlockHeight := uint64(1)
	if st.syncPointBlock.Height()+1 > core.ChunkSize {
		lastChunkBlockHeight = st.syncPointBlock.Height() - uint64(core.ChunkSize)
	}

	st.syncPointBlock = st.blockChain.GetBlockOnCanonicalChainByHeight(lastChunkBlockHeight)
}

func (st *Task) chunkHeadersRequest() {
	logging.VLog().WithFields(logrus.Fields{
		"syncPointBlockHeight": st.syncPointBlock.Height(),
		"syncPointBlockHash":   st.syncPointBlock.Hash().String(),
	}).Infof("Starting ChainSync at %d times.", st.chainSyncRetryCount)

	st.chainSyncRetryCount++

	chunkSync := &syncpb.Sync{
		TailBlockHash: st.syncPointBlock.Hash(),
	}

	data, err := proto.Marshal(chunkSync)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err":       err,
			"syncpoint": st.syncPointBlock,
		}).Debug("Failed to serialize sync message")
		return
	}

	// send message to peers.
	st.chainSyncPeers = st.netService.SendMessageToPeers(net.ChunkHeadersRequest, data,
		net.MessagePriorityLow, new(net.ChainSyncPeersFilter))
}

func (st *Task) processChunkHeaders(message net.Message) {
	// lock.
	st.syncMutex.Lock()
	defer st.syncMutex.Unlock()

	if st.hasEnoughChunkHeaders() {
		return
	}

	// verify the peers.
	if st.chainSyncPeers == nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": ErrInvalidChunkHeaderSourcePeer,
			"pid": message.MessageFrom(),
		}).Debug("Invalid ChainChunkHeaders message source peer, chinSyncPeers is NIL.")
		st.netService.ClosePeer(message.MessageFrom(), ErrInvalidChunkHeaderSourcePeer)
		return
	}

	isValidSourcePeer := false
	for _, prettyID := range st.chainSyncPeers { // TODO: why not map?
		if prettyID == message.MessageFrom() {
			isValidSourcePeer = true
			break
		}
	}

	if isValidSourcePeer == false {
		logging.VLog().WithFields(logrus.Fields{
			"err": ErrInvalidChunkHeaderSourcePeer,
			"pid": message.MessageFrom(),
		}).Debug("Invalid ChainChunkHeaders message source peer.")
		st.netService.ClosePeer(message.MessageFrom(), ErrInvalidChunkHeaderSourcePeer)
		return
	}

	chunkHeaders := new(syncpb.ChunkHeaders)
	if err := proto.Unmarshal(message.Data(), chunkHeaders); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
			"pid": message.MessageFrom(),
		}).Debug("Invalid ChainChunkHeaders message data.")
		st.netService.ClosePeer(message.MessageFrom(), ErrInvalidChainChunksMessageData)
		return
	}

	// verify chunk headers.
	if ok, err := verifyChunkHeaders(chunkHeaders); ok == false {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
			"pid": message.MessageFrom(),
		}).Debug("Wrong ChainChunkHeaders message data.")
		st.netService.ClosePeer(message.MessageFrom(), ErrWrongChainChunksMessageData)
		return
	}

	rootHash := byteutils.Hex(chunkHeaders.Root)

	hashPeerKey := fmt.Sprintf("%s-%s", rootHash, message.MessageFrom())
	if st.receivedChunkHeadersRootHashPeers[hashPeerKey] == true {
		logging.VLog().WithFields(logrus.Fields{
			"rootHash": rootHash,
			"pid":      message.MessageFrom(),
		}).Debug("Duplicated ChainChunkHeaders message data.")
		return
	}

	count := st.chunkHeadersRootHashCounter[rootHash] + 1
	st.chunkHeadersRootHashCounter[rootHash] += count
	st.receivedChunkHeadersRootHashPeers[hashPeerKey] = true
	st.maxConsistentChunkHeadersChainSyncPeers[rootHash] = append(st.maxConsistentChunkHeadersChainSyncPeers[rootHash], message.MessageFrom())

	isMax := false
	if count > st.maxConsistentChunkHeadersCount {
		isMax = true
		st.maxConsistentChunkHeadersCount = count
		st.maxConsistentChunkHeaders = chunkHeaders
	}

	logging.VLog().WithFields(logrus.Fields{
		"rootHash": rootHash,
		"count":    count,
		"isMax":    isMax,
		"pid":      message.MessageFrom(),
	}).Debug("Processed ChainChunkHeaders message data.")

	if st.hasEnoughChunkHeaders() {
		st.chainSyncDoneCh <- true
	}
}

func (st *Task) sendChunkDataRequest() {
	// lock.
	st.syncMutex.Lock()
	defer st.syncMutex.Unlock()

	if len(st.maxConsistentChunkHeaders.ChunkHeaders) == 0 {
		logging.VLog().WithFields(logrus.Fields{
			"maxConsistentChunkHeadersCount":    st.maxConsistentChunkHeadersCount,
			"maxConsistentChunkHeadersRootHash": byteutils.Hex(st.maxConsistentChunkHeaders.Root),
			"countOfChunkHeaders":               len(st.maxConsistentChunkHeaders.ChunkHeaders),
		}).Info("ChunkHeaders is empty, no need to sync.")

		// done.
		st.chinGetChunkDataDoneCh <- true
		return
	}

	currentSyncChunkDataCount := 0
	chainChunkDataSyncPosition := 0
	for i := 0; i < len(st.maxConsistentChunkHeaders.ChunkHeaders) && currentSyncChunkDataCount < ConcurrentSyncChunkDataCount; i++ {
		if st.chainChunkDataStatus[i] == chunkDataStatusNotStart {
			currentSyncChunkDataCount++
			chainChunkDataSyncPosition = i
			st.chunkDataRequest(i)
		}
	}

	st.chainChunkDataSyncPosition = chainChunkDataSyncPosition
}

func (st *Task) checkChainGetChunkTimeout() {
	// lock.
	st.syncMutex.Lock()
	defer st.syncMutex.Unlock()

	logging.VLog().WithFields(logrus.Fields{
		"syncPointBlockHeight": st.syncPointBlock.Height(),
		"syncPointBlockHash":   st.syncPointBlock.Hash().String(),
	}).Infof("Get Chunk at %d times.", st.chainSyncRetryCount)

	st.chainSyncRetryCount++

	for i := 0; i <= st.chainChunkDataSyncPosition; i++ {
		t := st.chainChunkDataStatus[i]

		if t == chunkDataStatusFinished || t == chunkDataStatusNotStart {
			continue
		}

		logging.VLog().WithFields(logrus.Fields{
			"rootHash": byteutils.Hex(st.maxConsistentChunkHeaders.Root),
			"timout":   time.Now().Unix() - st.chainChunkDataStatus[i],
		}).Debugf("Get Chunk %d Timout. Retry.", i)

		st.chunkDataRequest(i)
	}
}

func (st *Task) chunkDataRequest(chunkHeaderIndex int) {
	chunkHeader := st.maxConsistentChunkHeaders.ChunkHeaders[chunkHeaderIndex]
	data, err := proto.Marshal(chunkHeader)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Warn("Failed to marshal ChunkHeader.")
		return
	}

	// random
	peers := st.maxConsistentChunkHeadersChainSyncPeers[byteutils.Hex(st.maxConsistentChunkHeaders.Root)]
	idx := rand.Intn(len(peers))
	st.netService.SendMessageToPeer(net.ChunkDataRequest, data, net.MessagePriorityLow, peers[idx])

	st.chainChunkDataStatus[chunkHeaderIndex] = time.Now().Unix()

	logging.VLog().WithFields(logrus.Fields{
		"peers": peers,
	}).Debugf("Send to get chain chunk %d.", chunkHeaderIndex)
}

func (st *Task) processChunkData(message net.Message) {
	// lock.
	st.syncMutex.Lock()
	defer st.syncMutex.Unlock()

	// if maxConsistentChunkHeaders is nil, return
	if st.maxConsistentChunkHeaders == nil || st.maxConsistentChunkHeaders.ChunkHeaders == nil {
		logging.VLog().WithFields(logrus.Fields{
			"pid": message.MessageFrom(),
		}).Debug("Invalid ChainChunkData message data.")
		st.netService.ClosePeer(message.MessageFrom(), ErrInvalidChainChunkDataMessageData)
		return
	}

	chunkData := new(syncpb.ChunkData)
	if err := proto.Unmarshal(message.Data(), chunkData); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
			"pid": message.MessageFrom(),
		}).Debug("Invalid ChainChunkData message data.")
		st.netService.ClosePeer(message.MessageFrom(), ErrInvalidChainChunkDataMessageData)
		return
	}

	// verify chunk data.
	chunkDataIndex := -1
	var chunkHeader *syncpb.ChunkHeader

	for i := 0; i < len(st.maxConsistentChunkHeaders.ChunkHeaders); i++ { // TODO: why not map?
		chunkHeader = st.maxConsistentChunkHeaders.ChunkHeaders[i]
		if bytes.Compare(chunkHeader.Root, chunkData.Root) == 0 {
			chunkDataIndex = i
			break
		}
	}

	if chunkDataIndex < 0 {
		logging.VLog().WithFields(logrus.Fields{
			"pid": message.MessageFrom(),
		}).Debug("Wrong ChainChunkData message data.")
		st.netService.ClosePeer(message.MessageFrom(), ErrWrongChainChunkDataMessageData)
		return
	}

	if st.chainChunkDataStatus[chunkDataIndex] == chunkDataStatusFinished {
		logging.VLog().WithFields(logrus.Fields{
			"pid": message.MessageFrom(),
		}).Debug("Duplicated ChainChunkData message data.")
		return
	}

	if ok, err := verifyChunkData(chunkHeader, chunkData); ok == false {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
			"pid": message.MessageFrom(),
		}).Debug("Wrong ChainChunkData message data, retry.")
		st.netService.ClosePeer(message.MessageFrom(), err)
		st.chunkDataRequest(chunkDataIndex)
		return
	}

	st.chainChunkData[chunkDataIndex] = chunkData
	chunk, ok := st.chainChunkData[st.chainChunkDataProcessPosition]
	for ok {
		// startAt := time.Now().Unix()
		last, err := st.chunk.processChunkData(chunk)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err": err,
				"pid": message.MessageFrom(),
			}).Debug("Wrong ChainChunkData message data, retry.")
			st.netService.ClosePeer(message.MessageFrom(), err)
			st.chunkDataRequest(chunkDataIndex)
			return
		}

		st.syncPointBlock = last
		st.chainChunkDataProcessPosition++
		chunk, ok = st.chainChunkData[st.chainChunkDataProcessPosition]
	}

	// mark done.
	st.chainChunkDataStatus[chunkDataIndex] = chunkDataStatusFinished

	// sync next chunk.
	st.sendChainGetChunkForNext()
}

func (st *Task) sendChainGetChunkForNext() {
	nextPos := st.chainChunkDataSyncPosition + 1
	if nextPos >= len(st.maxConsistentChunkHeaders.ChunkHeaders) {
		if st.hasFinishedGetAllChunkData() {
			st.chinGetChunkDataDoneCh <- true
		}
		return
	}

	st.chainChunkDataSyncPosition = nextPos
	st.chunkDataRequest(nextPos)
}

func (st *Task) hasEnoughChunkHeaders() bool {
	chainSyncPeersCount := 0
	if st.chainSyncPeers != nil {
		chainSyncPeersCount = len(st.chainSyncPeers)
	}

	return chainSyncPeersCount > 0 && st.maxConsistentChunkHeadersCount >= int(chainSyncPeersCount/2)+1
}

func (st *Task) hasFinishedGetAllChunkData() bool {
	total := len(st.maxConsistentChunkHeaders.ChunkHeaders)
	missing := 0
	for i := 0; i < total; i++ {
		if st.chainChunkDataStatus[i] != chunkDataStatusFinished {
			missing++
		}
	}

	if missing > 0 {
		logging.VLog().WithFields(logrus.Fields{
			"totalSyncingChunkHeaders": total,
			"missingCount":             missing,
		}).Debug("Waiting for ChunkData.")
		return false
	}

	logging.VLog().Info("Received enough chunk data.")
	return true
}
