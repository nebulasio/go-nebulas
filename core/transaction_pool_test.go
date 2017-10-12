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
	"testing"

	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/crypto/keystore/ecdsa"
	"github.com/stretchr/testify/assert"
)

func TestTransactionPool(t *testing.T) {
	from, _ := Parse("91da63ba4ec9f6a08636d9abd443f64b4855be7fa9e44aa2")
	other, _ := Parse("c7a9cbc5d69126fa4354ac05e16d5e781c8ab4dc61850cd8")
	txs := []*Transaction{
		NewTransaction(1, *from, Address{[]byte("to")}, 0, 10, []byte("datadata")),
		NewTransaction(1, *other, Address{[]byte("to")}, 0, 1, []byte("datadata")),
		NewTransaction(1, *from, Address{[]byte("to")}, 0, 1, []byte("da")),

		NewTransaction(1, *from, Address{[]byte("to")}, 0, 2, []byte("da")),
		NewTransaction(0, *from, Address{[]byte("to")}, 0, 0, []byte("da")),

		NewTransaction(1, *other, Address{[]byte("to")}, 0, 1, []byte("data")),
		NewTransaction(1, *from, Address{[]byte("to")}, 0, 1, []byte("datadata")),
	}
	ks := keystore.DefaultKS
	arr := make([][]byte, 3)
	arr[0] = []byte{59, 144, 87, 239, 199, 27, 51, 230, 209, 177, 177, 166, 161, 23, 23, 195, 197, 245, 56, 156, 171, 40, 209, 7, 25, 1, 32, 0, 75, 69, 145, 30}
	arr[1] = []byte{208, 98, 189, 16, 69, 97, 14, 44, 112, 56, 253, 61, 195, 100, 88, 245, 99, 14, 70, 22, 173, 172, 243, 186, 46, 128, 18, 39, 93, 125, 27, 186}
	for _, pdata := range arr {
		priv, _ := ecdsa.ToPrivateKey(pdata)
		pubdata, _ := ecdsa.FromPublicKey(&priv.PublicKey)
		addr, _ := NewAddressFromPublicKey(pubdata)
		ps := ecdsa.NewPrivateStoreKey(priv)
		ks.SetKey(addr.ToHex(), ps, []byte("passphrase"))
		ks.Unlock(addr.ToHex(), []byte("passphrase"), 10)
	}
	txPool := NewTransactionPool(3)
	bc := NewBlockChain(1)
	txPool.setBlockChain(bc)
	txs[0].Sign()
	assert.Nil(t, txPool.Push(txs[0]))
	// put dup tx, should fail
	assert.NotNil(t, txPool.Push(txs[0]))
	txs[1].Sign()
	assert.Nil(t, txPool.Push(txs[1]))
	txs[2].Sign()
	assert.Nil(t, txPool.Push(txs[2]))
	// put not signed tx, should fail
	assert.NotNil(t, txPool.Push(txs[3]))
	// put tx with different chainID, should fail
	txs[4].Sign()
	assert.NotNil(t, txPool.Push(txs[4]))
	// put one new, replace txs[1]
	assert.Equal(t, len(txPool.all), 3)
	assert.Equal(t, txPool.cache.Len(), 3)
	txs[6].Sign()
	assert.Nil(t, txPool.Push(txs[6]))
	assert.Equal(t, txPool.cache.Len(), 3)
	assert.Equal(t, len(txPool.all), 3)
	// get from: other, nonce: 1, data: "da"
	tx1 := txPool.Pop()
	assert.Equal(t, txs[2].from.address, tx1.from.address)
	assert.Equal(t, txs[2].nonce, tx1.nonce)
	assert.Equal(t, txs[2].data, tx1.data)
	// put one new
	assert.Equal(t, len(txPool.all), 2)
	assert.Equal(t, txPool.cache.Len(), 2)
	txs[5].Sign()
	assert.Nil(t, txPool.Push(txs[5]))
	assert.Equal(t, len(txPool.all), 3)
	assert.Equal(t, txPool.cache.Len(), 3)
	// get 2 txs, txs[5], txs[0]
	tx21 := txPool.Pop()
	tx22 := txPool.Pop()
	assert.Equal(t, txs[5].from.address, tx21.from.address)
	assert.Equal(t, txs[5].Nonce(), tx21.Nonce())
	assert.Equal(t, txs[5].data, tx21.data)
	assert.Equal(t, txs[6].from.address, tx22.from.address)
	assert.Equal(t, txs[6].Nonce(), tx22.Nonce())
	assert.Equal(t, txs[6].data, tx22.data)
	assert.Equal(t, txPool.Empty(), false)
	txPool.Pop()
	assert.Equal(t, txPool.Empty(), true)
	assert.Nil(t, txPool.Pop())
	/* 	// put new txs
	   	assert.Equal(t, len(txPool.all), 0)
	   	assert.Equal(t, txPool.cache.Len(), 0)
	   	// mock some txs in parent block
	   	pb1, _ := txs[1].ToProto()
	   	txBytes1, _ := proto.Marshal(pb1)
	   	account1 := &corepb.Account{Balance: 0, Nonce: 1}
	   	accBytes1, _ := proto.Marshal(account1)
	   	bc.tailBlock.stateTrie.Put(txs[1].from.address, accBytes1)
	   	bc.tailBlock.txsTrie.Put(txs[1].hash, txBytes1)
	   	pb2, _ := txs[2].ToProto()
	   	txBytes2, _ := proto.Marshal(pb2)
	   	account2 := &corepb.Account{Balance: 0, Nonce: 1}
	   	accBytes2, _ := proto.Marshal(account2)
	   	bc.tailBlock.stateTrie.Put(txs[2].from.address, accBytes2)
	   	bc.tailBlock.txsTrie.Put(txs[2].hash, txBytes2)
	   	// put some new txs for testing stateTrie & txsTrie
	   	txs[7].Sign()
	   	txPool.PutTxs([]*Transaction{txs[1], txs[6], txs[7]})
	   	assert.Equal(t, len(txPool.all), 3)
	   	assert.Equal(t, txPool.cache.Len(), 3)
	   	tx3 := txPool.Get(4, bc.tailBlock)
	   	assert.Equal(t, len(tx3), 1)
	   	assert.Equal(t, txs[7].from.address, tx3[0].from.address)
	   	assert.Equal(t, txs[7].Nonce(), tx3[0].Nonce())
	   	assert.Equal(t, txs[7].data, tx3[0].data)
	   	// put bigger nonce tx
	   	assert.Equal(t, len(txPool.all), 0)
	   	assert.Equal(t, txPool.cache.Len(), 0)
	   	txs[8].Sign()
	   	txPool.PutTxs([]*Transaction{txs[8]})
	   	tx4 := txPool.Get(4, bc.tailBlock)
	   	assert.Equal(t, len(tx4), 0)
	   	assert.Equal(t, len(txPool.all), 1)
	   	assert.Equal(t, txPool.cache.Len(), 1) */
}
