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

package core

import (
	"sync"

	"github.com/nebulasio/go-nebulas/components/net"
	log "github.com/sirupsen/logrus"
)

// BlockPool a pool of all received blocks from network.
// Blocks will be sent to Consensus when it passes signature verification.
type BlockPool struct {
	inner            *sync.Map
	receiveMessageCh chan net.Message
	receivedBlockCh  chan *Block
	quitCh           chan int
}

// NewBlockPool return new #BlockPool instance.
func NewBlockPool() *BlockPool {
	bp := &BlockPool{
		inner:            new(sync.Map),
		receiveMessageCh: make(chan net.Message, 128),
		receivedBlockCh:  make(chan *Block, 128),
	}
	return bp
}

// ReceivedBlockCh return received block chan.
func (pool *BlockPool) ReceivedBlockCh() chan *Block {
	return pool.receivedBlockCh
}

// RegisterInNetwork register message subscriber in network.
func (pool *BlockPool) RegisterInNetwork(nm net.Manager) {
	nm.Register(net.NewSubscriber(pool, pool.receiveMessageCh, net.MessageTypeNewBlock))
}

// Range calls f sequentially for each key and value present in the map.
// If f returns false, range stops the iteration.
func (pool *BlockPool) Range(f func(key, value interface{}) bool) {
	pool.inner.Range(f)
}

// Delete delete key from pool.
func (pool *BlockPool) Delete(keys ...HexHash) {
	for _, key := range keys {
		pool.inner.Delete(key)
	}
}

// Start start loop.
func (pool *BlockPool) Start() {
	go pool.loop()
}

// Stop stop loop.
func (pool *BlockPool) Stop() {
	pool.quitCh <- 0
}

func (pool *BlockPool) loop() {
	log.Info("BlockPool.loop: running.")
	count := 0
	for {
		select {
		case <-pool.quitCh:
			log.Info("BlockPool.loop: quit.")
			return
		case msg := <-pool.receiveMessageCh:
			count++
			log.Debugf("BlockPool.loop: received message. Count=%d", count)
			if msg.MessageType() != net.MessageTypeNewBlock {
				log.WithFields(log.Fields{
					"messageType": msg.MessageType(),
					"message":     msg,
				}).Error("BlockPool.loop: received unregistered message, pls check code.")
				continue
			}

			// verify signature.
			block := msg.Data().(*Block)
			if block.VerifySign() == false {
				log.WithFields(log.Fields{
					"block": block,
				}).Error("BlockPool.loop: the signature of block is invalid.")
				continue
			}

			// send to chan.
			pool.inner.Store(block.Hash().Hex(), block)
			pool.receivedBlockCh <- block
		}
	}
}

// AddLocalBlock add local minted block.
func (pool *BlockPool) AddLocalBlock(block *Block) {
	if block.VerifySign() == false {
		log.WithFields(log.Fields{
			"block": block,
		}).Error("BlockPool.AddLocalBlock: the signature of block is invalid.")
		return
	}

	// send to chan.
	pool.inner.Store(block.Hash().Hex(), block)
	pool.receivedBlockCh <- block
}
