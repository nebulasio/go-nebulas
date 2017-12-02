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
	"errors"
	"sync"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/common/pdeque"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/net"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	log "github.com/sirupsen/logrus"
)

// TransactionPool cache txs, is thread safe
type TransactionPool struct {
	receivedMessageCh chan net.Message
	quitCh            chan int
	mu                sync.RWMutex

	size  int
	cache *pdeque.PriorityDeque
	all   map[byteutils.HexHash]*Transaction
	bc    *BlockChain

	nm net.Manager
	//TODO: miner should have the lowest gasPrice & maximum gasLimt
}

func less(a interface{}, b interface{}) bool {
	txa := a.(*Transaction)
	txb := b.(*Transaction)
	if txa.from.Equals(txb.from) {
		return txa.Nonce() < txb.Nonce()
	}
	if txa.gasPrice.Cmp(txb.gasPrice.Int) != 0 {
		// txa.gasPrice < txb.gasPrice
		return txa.GasPrice().Cmp(txb.GasPrice().Int) == -1
	}
	// txa.gasLimit < txb.gasLimit
	return txa.GasLimit().Cmp(txb.GasLimit().Int) == -1
}

// NewTransactionPool create a new TransactionPool
func NewTransactionPool(size int) *TransactionPool {
	if size == 0 {
		panic("cannot new txpool with size == 0")
	}
	txPool := &TransactionPool{
		receivedMessageCh: make(chan net.Message, 128),
		quitCh:            make(chan int, 1),
		size:              size,
		cache:             pdeque.NewPriorityDeque(less),
		all:               make(map[byteutils.HexHash]*Transaction),
	}
	return txPool
}

// RegisterInNetwork register message subscriber in network.
func (pool *TransactionPool) RegisterInNetwork(nm net.Manager) {
	nm.Register(net.NewSubscriber(pool, pool.receivedMessageCh, MessageTypeNewTx))
	pool.nm = nm
}

func (pool *TransactionPool) setBlockChain(bc *BlockChain) {
	pool.bc = bc
}

// Start start loop.
func (pool *TransactionPool) Start() {
	go pool.loop()
}

// Stop stop loop.
func (pool *TransactionPool) Stop() {
	pool.quitCh <- 0
}

func (pool *TransactionPool) loop() {
	log.WithFields(log.Fields{
		"func": "TxPool.loop",
	}).Debug("running.")

	count := 0
	for {
		select {
		case <-pool.quitCh:
			log.WithFields(log.Fields{
				"func": "TxPool.loop",
			}).Info("quit.")
			return
		case msg := <-pool.receivedMessageCh:
			count++
			log.WithFields(log.Fields{
				"func": "TxPool.loop",
			}).Debugf("received message. Count=%d", count)

			if msg.MessageType() != MessageTypeNewTx {
				log.WithFields(log.Fields{
					"func":        "TxPool.loop",
					"messageType": msg.MessageType(),
					"message":     msg,
				}).Error("TxPool.loop: received unregistered message, pls check code.")
				continue
			}

			tx := new(Transaction)
			pbTx := new(corepb.Transaction)
			if err := proto.Unmarshal(msg.Data().([]byte), pbTx); err != nil {
				log.Error("TxPool.loop:: unmarshal data occurs error, ", err)
				continue
			}
			if err := tx.FromProto(pbTx); err != nil {
				log.Error("TxPool.loop:: get block from proto occurs error: ", err)
				continue
			}
			if err := pool.PushAndRelay(tx); err != nil {
				log.WithFields(log.Fields{
					"func":        "TxPool.loop",
					"messageType": msg.MessageType(),
					"transaction": tx,
					"err":         err,
				}).Warn("TxPool.loop: invalid transaction, drop it.")
				continue
			}
		}
	}
}

// Push tx into pool
func (pool *TransactionPool) Push(tx *Transaction) error {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	return pool.push(tx)
}

// PushAndRelay push tx into pool and relay it
func (pool *TransactionPool) PushAndRelay(tx *Transaction) error {
	if err := pool.Push(tx); err != nil {
		return err
	}
	pool.nm.Relay(MessageTypeNewTx, tx)
	return nil
}

// PushAndBroadcast push tx into pool and broadcast it
func (pool *TransactionPool) PushAndBroadcast(tx *Transaction) error {
	if err := pool.Push(tx); err != nil {
		return err
	}
	pool.nm.Broadcast(MessageTypeNewTx, tx)
	return nil
}

func (pool *TransactionPool) push(tx *Transaction) error {
	// verify hash & sign of tx
	if err := tx.Verify(pool.bc.chainID); err != nil {
		return err
	}

	// verify non-dup tx
	if _, ok := pool.all[tx.hash.Hex()]; ok {
		return errors.New("duplicate tx")
	}
	// cache the verified tx
	pool.cache.Insert(tx)
	pool.all[tx.hash.Hex()] = tx
	// delete tx with lowest priority if cache is full
	if pool.cache.Len() > pool.size {
		tx := pool.cache.PopMax().(*Transaction)
		delete(pool.all, tx.hash.Hex())
	}
	return nil
}

// Pop a transaction from pool
func (pool *TransactionPool) Pop() *Transaction {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	return pool.pop()
}

func (pool *TransactionPool) pop() *Transaction {
	if pool.cache.Len() > 0 {
		tx := pool.cache.PopMin().(*Transaction)
		delete(pool.all, tx.hash.Hex())
		return tx
	}
	return nil
}

// Empty return if the pool is empty
func (pool *TransactionPool) Empty() bool {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	return pool.cache.Len() == 0
}
