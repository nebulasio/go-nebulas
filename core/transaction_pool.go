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
	"github.com/hashicorp/golang-lru"
)

// TransactionPool contains a pool of transactions
type TransactionPool struct {
	inner *lru.Cache
}

// NewTransactionPool create a new TransactionPool
func NewTransactionPool() *TransactionPool {
	txPool := &TransactionPool{}
	txPool.inner, _ = lru.New(1024)
	return txPool
}

// Put put transaction to pool
func (txPool *TransactionPool) Put(tx *Transaction) {
	txPool.inner.Add(tx.hash.Hex(), tx)
}
