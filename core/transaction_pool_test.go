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

	"time"

	"github.com/nebulasio/go-nebulas/crypto"
	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/crypto/keystore/secp256k1"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/stretchr/testify/assert"
)

func TestTransactionPool(t *testing.T) {
	ks := keystore.DefaultKS
	priv1 := secp256k1.GeneratePrivateKey()
	pubdata1, _ := priv1.PublicKey().Encoded()
	from, _ := NewAddressFromPublicKey(pubdata1)
	ks.SetKey(from.String(), priv1, []byte("passphrase"))
	ks.Unlock(from.String(), []byte("passphrase"), time.Second*60*60*24*365)
	key1, _ := ks.GetUnlocked(from.String())
	signature1, _ := crypto.NewSignature(keystore.SECP256K1)
	signature1.InitSign(key1.(keystore.PrivateKey))

	priv2 := secp256k1.GeneratePrivateKey()
	pubdata2, _ := priv2.PublicKey().Encoded()
	other, _ := NewAddressFromPublicKey(pubdata2)
	ks.SetKey(other.String(), priv2, []byte("passphrase"))
	ks.Unlock(other.String(), []byte("passphrase"), time.Second*60*60*24*365)
	key2, _ := ks.GetUnlocked(other.String())
	signature2, _ := crypto.NewSignature(keystore.SECP256K1)
	signature2.InitSign(key2.(keystore.PrivateKey))

	heighPrice := util.NewUint128FromBigInt(util.NewUint128().Mul(TransactionGasPrice.Int, util.NewUint128FromInt(2).Int))
	txPool, _ := NewTransactionPool(3)
	bc, _ := NewBlockChain(testNeb())
	txPool.setBlockChain(bc)

	txs := []*Transaction{
		NewTransaction(bc.ChainID(), from, &Address{[]byte("to")}, util.NewUint128(), 10, TxPayloadBinaryType, []byte("datadata"), TransactionGasPrice, util.NewUint128FromInt(200000)),
		NewTransaction(bc.ChainID(), other, &Address{[]byte("to")}, util.NewUint128(), 1, TxPayloadBinaryType, []byte("datadata"), heighPrice, util.NewUint128FromInt(200000)),
		NewTransaction(bc.ChainID(), from, &Address{[]byte("to")}, util.NewUint128(), 1, TxPayloadBinaryType, []byte("da"), TransactionGasPrice, util.NewUint128FromInt(200000)),

		NewTransaction(bc.ChainID(), from, &Address{[]byte("to")}, util.NewUint128(), 2, TxPayloadBinaryType, []byte("da"), TransactionGasPrice, util.NewUint128FromInt(200000)),
		NewTransaction(bc.ChainID()+1, from, &Address{[]byte("to")}, util.NewUint128(), 0, TxPayloadBinaryType, []byte("da"), TransactionGasPrice, util.NewUint128FromInt(200000)),

		NewTransaction(bc.ChainID(), other, &Address{[]byte("to")}, util.NewUint128(), 1, TxPayloadBinaryType, []byte("data"), TransactionGasPrice, util.NewUint128FromInt(200000)),
		NewTransaction(bc.ChainID(), from, &Address{[]byte("to")}, util.NewUint128(), 1, TxPayloadBinaryType, []byte("datadata"), heighPrice, util.NewUint128FromInt(200000)),
	}

	assert.Nil(t, txs[0].Sign(signature1))
	assert.Nil(t, txPool.Push(txs[0]))
	// put dup tx, should fail
	assert.NotNil(t, txPool.Push(txs[0]))
	assert.Nil(t, txs[1].Sign(signature2))
	assert.Nil(t, txPool.Push(txs[1]))
	assert.Nil(t, txs[2].Sign(signature1))
	assert.Nil(t, txPool.Push(txs[2]))
	// put not signed tx, should fail
	assert.NotNil(t, txPool.Push(txs[3]))
	// put tx with different chainID, should fail
	assert.Nil(t, txs[4].Sign(signature1))
	assert.NotNil(t, txPool.Push(txs[4]))
	// put one new, replace txs[1]
	assert.Equal(t, len(txPool.all), 3)
	assert.Equal(t, txPool.cache.Len(), 3)
	assert.Nil(t, txs[6].Sign(signature1))
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
	assert.Nil(t, txs[5].Sign(signature2))
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
}

func TestGasConfig(t *testing.T) {
	txPool, _ := NewTransactionPool(3)
	txPool.SetGasConfig(nil, nil)
	assert.Equal(t, txPool.gasPrice, TransactionGasPrice)
	assert.Equal(t, txPool.gasLimit, TransactionMaxGas)
	txPool.SetGasConfig(util.NewUint128FromInt(1), util.NewUint128FromInt(1))
	assert.Equal(t, txPool.gasPrice, util.NewUint128FromInt(1))
	assert.Equal(t, txPool.gasLimit, util.NewUint128FromInt(1))
}

func TestPushTxs(t *testing.T) {
	ks := keystore.DefaultKS
	priv1 := secp256k1.GeneratePrivateKey()
	pubdata1, _ := priv1.PublicKey().Encoded()
	from, _ := NewAddressFromPublicKey(pubdata1)
	ks.SetKey(from.String(), priv1, []byte("passphrase"))
	ks.Unlock(from.String(), []byte("passphrase"), time.Second*60*60*24*365)
	key1, _ := ks.GetUnlocked(from.String())
	signature1, _ := crypto.NewSignature(keystore.SECP256K1)
	signature1.InitSign(key1.(keystore.PrivateKey))

	priv2 := secp256k1.GeneratePrivateKey()
	pubdata2, _ := priv2.PublicKey().Encoded()
	to, _ := NewAddressFromPublicKey(pubdata2)
	ks.SetKey(to.String(), priv2, []byte("passphrase"))
	ks.Unlock(to.String(), []byte("passphrase"), time.Second*60*60*24*365)
	key2, _ := ks.GetUnlocked(to.String())
	signature2, _ := crypto.NewSignature(keystore.SECP256K1)
	signature2.InitSign(key2.(keystore.PrivateKey))

	txPool, _ := NewTransactionPool(3)
	bc, _ := NewBlockChain(testNeb())
	txPool.setBlockChain(bc)
	MaxGasPlus1 := util.NewUint128FromBigInt(util.NewUint128().Add(TransactionMaxGas.Int, util.NewUint128FromInt(1).Int))
	txs := []*Transaction{
		NewTransaction(bc.ChainID(), from, to, util.NewUint128(), 10, TxPayloadBinaryType, []byte("datadata"), util.NewUint128FromInt(10^6-1), TransactionMaxGas),
		NewTransaction(bc.ChainID(), from, to, util.NewUint128(), 10, TxPayloadBinaryType, []byte("datadata"), TransactionGasPrice, MaxGasPlus1),
	}
	assert.Equal(t, txPool.push(txs[0]), ErrBelowGasPrice)
	assert.Equal(t, txPool.push(txs[1]), ErrOutOfGasLimit)
}
