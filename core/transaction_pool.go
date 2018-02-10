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
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/common/pdeque"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/net"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

// TransactionPool cache txs, is thread safe
type TransactionPool struct {
	receivedMessageCh chan net.Message
	quitCh            chan int

	size  int
	cache *pdeque.PriorityDeque
	all   map[byteutils.HexHash]*Transaction
	bc    *BlockChain

	ns net.Service
	mu sync.RWMutex

	gasPrice *util.Uint128 // the lowest gasPrice.
	gasLimit *util.Uint128 // the maximum gasLimit.

	eventEmitter *EventEmitter
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
func NewTransactionPool(size int) (*TransactionPool, error) {
	txPool := &TransactionPool{
		receivedMessageCh: make(chan net.Message, size),
		quitCh:            make(chan int, 1),
		size:              size,
		cache:             pdeque.NewPriorityDeque(less),
		all:               make(map[byteutils.HexHash]*Transaction),
		gasPrice:          TransactionGasPrice,
		gasLimit:          TransactionMaxGas,
	}
	return txPool, nil
}

// SetGasConfig config the lowest gasPrice and the maximum gasLimit.
func (pool *TransactionPool) SetGasConfig(gasPrice, gasLimit *util.Uint128) {
	if gasPrice == nil || gasPrice.Cmp(util.NewUint128().Int) <= 0 {
		pool.gasPrice = TransactionGasPrice
	} else {
		pool.gasPrice = gasPrice
	}
	if gasLimit == nil || gasLimit.Cmp(util.NewUint128().Int) == 0 || gasLimit.Cmp(TransactionMaxGas.Int) > 0 {
		pool.gasLimit = TransactionMaxGas
	} else {
		pool.gasLimit = gasLimit
	}
}

// RegisterInNetwork register message subscriber in network.
func (pool *TransactionPool) RegisterInNetwork(ns net.Service) {
	ns.Register(net.NewSubscriber(pool, pool.receivedMessageCh, true, MessageTypeNewTx, net.MessageWeightNewTx))
	pool.ns = ns
}

func (pool *TransactionPool) setBlockChain(bc *BlockChain) {
	pool.bc = bc
}

func (pool *TransactionPool) setEventEmitter(emitter *EventEmitter) {
	pool.eventEmitter = emitter
}

// Start start loop.
func (pool *TransactionPool) Start() {
	logging.CLog().WithFields(logrus.Fields{
		"size": pool.size,
	}).Info("Starting TransactionPool...")

	go pool.loop()
}

// Stop stop loop.
func (pool *TransactionPool) Stop() {
	logging.CLog().WithFields(logrus.Fields{
		"size": pool.size,
	}).Info("Stop TransactionPool.")

	pool.quitCh <- 0
}

func (pool *TransactionPool) loop() {
	logging.CLog().WithFields(logrus.Fields{
		"size": pool.size,
	}).Info("Started TransactionPool.")

	timerChan := time.NewTicker(time.Second).C
	for {
		select {
		case <-timerChan:
			metricsCachedTx.Update(int64(len(pool.receivedMessageCh)))
		case <-pool.quitCh:
			logging.CLog().WithFields(logrus.Fields{
				"size": pool.size,
			}).Info("Stopped TransactionPool.")
			return
		case msg := <-pool.receivedMessageCh:
			if msg.MessageType() != MessageTypeNewTx {
				logging.VLog().WithFields(logrus.Fields{
					"messageType": msg.MessageType(),
					"message":     msg,
					"err":         "not new tx msg",
				}).Debug("Received unregistered message.")
				continue
			}

			tx := new(Transaction)
			pbTx := new(corepb.Transaction)
			if err := proto.Unmarshal(msg.Data(), pbTx); err != nil {
				logging.VLog().WithFields(logrus.Fields{
					"msgType": msg.MessageType(),
					"msg":     msg,
					"err":     err,
				}).Debug("Failed to unmarshal data.")
				continue
			}
			if err := tx.FromProto(pbTx); err != nil {
				logging.VLog().WithFields(logrus.Fields{
					"msgType": msg.MessageType(),
					"msg":     msg,
					"err":     err,
				}).Debug("Failed to recover a tx from proto data.")
				continue
			}

			/* 			logging.VLog().WithFields(logrus.Fields{
				"tx":   tx,
				"type": msg.MessageType(),
			}).Debug("Received a new tx.") */

			if err := pool.PushAndRelay(tx); err != nil {
				logging.VLog().WithFields(logrus.Fields{
					"func":        "TxPool.loop",
					"messageType": msg.MessageType(),
					"transaction": tx,
					"err":         err,
				}).Debug("Failed to push a tx into tx pool.")
				continue
			}
		}
	}
}

// GetTransaction return transaction of given hash from transaction pool.
func (pool *TransactionPool) GetTransaction(hash byteutils.Hash) *Transaction {
	return pool.all[hash.Hex()]
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

	pool.ns.Relay(MessageTypeNewTx, tx, net.MessagePriorityNormal)
	return nil
}

// PushAndBroadcast push tx into pool and broadcast it
func (pool *TransactionPool) PushAndBroadcast(tx *Transaction) error {
	if err := pool.Push(tx); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"tx":  tx,
			"err": err,
		}).Debug("Failed to push a new tx into tx pool")
		return err
	}

	pool.ns.Broadcast(MessageTypeNewTx, tx, net.MessagePriorityNormal)
	return nil
}

func (pool *TransactionPool) push(tx *Transaction) error {
	// verify non-dup tx
	if _, ok := pool.all[tx.hash.Hex()]; ok {
		metricsDuplicateTx.Inc(1)
		return ErrDuplicatedTransaction
	}

	// if tx's gasPrice below the pool config lowest gasPrice, return ErrBelowGasPrice
	if tx.gasPrice.Cmp(pool.gasPrice.Int) < 0 {
		metricsTxPoolBelowGasPrice.Inc(1)
		return ErrBelowGasPrice
	}
	if tx.gasLimit.Cmp(pool.gasLimit.Int) > 0 {
		metricsTxPoolOutOfGasLimit.Inc(1)
		return ErrOutOfGasLimit
	}

	// verify hash & sign of tx
	if err := tx.VerifyIntegrity(pool.bc.chainID); err != nil {
		metricsInvalidTx.Inc(1)
		return err
	}

	// cache the verified tx
	pool.cache.Insert(tx)
	pool.all[tx.hash.Hex()] = tx
	// delete tx with lowest priority if cache is full
	if pool.cache.Len() > pool.size {
		tx := pool.cache.PopMax().(*Transaction)
		delete(pool.all, tx.hash.Hex())
	}

	// trigger pending transaction
	event := &Event{
		Topic: TopicPendingTransaction,
		Data:  tx.String(),
	}
	pool.eventEmitter.Trigger(event)

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
