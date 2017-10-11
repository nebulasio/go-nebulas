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
	"github.com/nebulasio/go-nebulas/common/pdeq"
	"github.com/nebulasio/go-nebulas/components/net"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	log "github.com/sirupsen/logrus"
)

// TransactionPool cache txs, is thread safe
type TransactionPool struct {
	receivedMessageCh chan net.Message
	quitCh            chan int
	mu                sync.RWMutex

	size  int
	cache *pdeq.Pdeq
	all   map[HexHash]*Transaction
	bc    *BlockChain
}

func less(a interface{}, b interface{}) bool {
	txa := a.(*Transaction)
	txb := b.(*Transaction)
	if byteutils.Equal(txa.From(), txb.From()) {
		return txa.Nonce() < txb.Nonce()
	}
	// TODO(shshang): use gas price instead
	return txa.DataLen() < txb.DataLen()
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
		cache:             pdeq.NewPdeq(less),
		all:               make(map[HexHash]*Transaction),
	}
	return txPool
}

// RegisterInNetwork register message subscriber in network.
func (pool *TransactionPool) RegisterInNetwork(nm net.Manager) {
	nm.Register(net.NewSubscriber(pool, pool.receivedMessageCh, net.MessageTypeNewTx))
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

			if msg.MessageType() != net.MessageTypeNewTx {
				log.WithFields(log.Fields{
					"func":        "TxPool.loop",
					"messageType": msg.MessageType(),
					"message":     msg,
				}).Error("TxPool.loop: received unregistered message, pls check code.")
				continue
			}

			tx := msg.Data().(*Transaction)
			pool.Put(tx)
		}
	}
}

// Put tx into pool
// verify chainID, hash, sign, and duplication
// if cache is full, delete the lowest priority tx
func (pool *TransactionPool) Put(tx *Transaction) error {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	return pool.put(tx)
}

func (pool *TransactionPool) put(tx *Transaction) error {
	// verify chainID
	if tx.chainID != pool.bc.chainID {
		return errors.New("cannot cache transactions in different chain")
	}
	// verify hash & sign of tx
	if err := tx.Verify(); err != nil {
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

// PutTxs put many txs in once
func (pool *TransactionPool) PutTxs(txs []*Transaction) {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	for _, v := range txs {
		pool.put(v)
	}
}

// Get n verified transaction from pool for packing a new block
// 1. always choose highest priority txs
// 2. tx chosen cannot exist in the chain of parent block already
// 3. the nonce of tx chosen must be one more than current nonce
func (pool *TransactionPool) Get(n int, parent *Block) []*Transaction {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	return pool.get(n, parent)
}

func (pool *TransactionPool) get(n int, parent *Block) []*Transaction {
	// store the new block's txs
	var res []*Transaction
	// store the txs which should be giveback to pool
	var giveback []*Transaction
	// store account's current nonce in new block's txs
	nonces := make(map[HexHash]uint64)
	// choose valid tx until n txs are selected
	for pool.cache.Len() > 0 && len(res) < n {
		// get highest priority tx
		tx := pool.cache.PopMin().(*Transaction)
		from := tx.from.address.Hex()
		delete(pool.all, tx.hash.Hex())
		// verify if tx exists in the chain of the parent block
		if proof, _ := parent.txsTrie.Prove(tx.hash); proof != nil {
			// existed, drop the tx
			continue
		}
		// verify if tx's nonce is one more than current nonce
		account := new(corepb.Account)
		// current nonce, default 0
		var curNonce uint64
		// get account's current nonce in parent block
		if accBytes, _ := parent.stateTrie.Get(tx.from.address); accBytes != nil {
			// account existed
			if err := proto.Unmarshal(accBytes, account); err != nil {
				panic("account in stateTrie cannot be unmarshaled correctly")
			}
			curNonce = account.Nonce
		}
		// get account's current nonce in new block's txs
		if nonce, ok := nonces[from]; ok {
			curNonce = nonce
		}
		if tx.nonce == curNonce+1 {
			// right tx, accepted
			nonces[from] = tx.nonce
			res = append(res, tx)
		} else if tx.nonce > curNonce+1 {
			// tx with bigger nonce is future tx, giveback it to tx pool
			giveback = append(giveback, tx)
		} // drop tx with smaller nonce
	}
	// giveback future txs
	for _, v := range giveback {
		pool.put(v)
	}
	return res
}
