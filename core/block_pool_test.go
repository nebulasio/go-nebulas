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
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/stretchr/testify/assert"
)

type MockConsensus struct {
	storage storage.Storage
}

func (c MockConsensus) FastVerifyBlock(block *Block) error {
	block.miner = block.Coinbase()
	return nil
}
func (c MockConsensus) VerifyBlock(block *Block, parent *Block) error {
	block.miner = block.Coinbase()
	return nil
}

func TestBlockPool(t *testing.T) {
	storage, _ := storage.NewMemoryStorage()
	bc, err := NewBlockChain(0, storage)
	cons := &MockConsensus{storage}
	bc.SetConsensusHandler(cons)
	pool := bc.bkPool
	assert.Equal(t, pool.blockCache.Len(), 0)

	ks := keystore.DefaultKS
	priv := secp256k1.GeneratePrivateKey()
	pubdata, _ := priv.PublicKey().Encoded()
	from, _ := NewAddressFromPublicKey(pubdata)
	ks.SetKey(from.ToHex(), priv, []byte("passphrase"))
	ks.Unlock(from.ToHex(), []byte("passphrase"), time.Second*60*60*24*365)
	to := &Address{from.address}
	key, _ := ks.GetUnlocked(from.ToHex())
	signature, _ := crypto.NewSignature(keystore.SECP256K1)
	signature.InitSign(key.(keystore.PrivateKey))
	bc.tailBlock.begin()
	bc.tailBlock.accState.GetOrCreateUserAccount(from.Bytes()).AddBalance(util.NewUint128FromInt(1000000))
	bc.tailBlock.header.stateRoot = bc.tailBlock.accState.RootHash()
	bc.tailBlock.commit()
	bc.storeBlockToStorage(bc.tailBlock)

	validators, err := TraverseDynasty(bc.tailBlock.dposContext.dynastyTrie)
	assert.Nil(t, err)

	addr := &Address{validators[1]}
	block0, err := NewBlock(0, addr, bc.tailBlock)
	assert.Nil(t, err)
	block0.header.timestamp = bc.tailBlock.header.timestamp + BlockInterval
	block0.SetMiner(addr)
	block0.Seal()

	tx1 := NewTransaction(0, from, to, util.NewUint128FromInt(1), 1, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, util.NewUint128FromInt(200000))
	tx1.Sign(signature)
	tx2 := NewTransaction(0, from, to, util.NewUint128FromInt(2), 2, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, util.NewUint128FromInt(200000))
	tx2.Sign(signature)
	tx3 := NewTransaction(0, from, to, util.NewUint128FromInt(3), 3, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, util.NewUint128FromInt(200000))
	tx3.Sign(signature)
	err = bc.txPool.Push(tx1)
	assert.NoError(t, err)
	err = bc.txPool.Push(tx2)
	assert.NoError(t, err)
	err = bc.txPool.Push(tx3)
	assert.NoError(t, err)

	addr = &Address{validators[2]}
	block1, _ := NewBlock(0, addr, block0)
	block1.header.timestamp = block0.header.timestamp + BlockInterval
	block1.CollectTransactions(1)
	block1.SetMiner(addr)
	block1.Seal()

	addr = &Address{validators[3]}
	block2, _ := NewBlock(0, addr, block1)
	block2.header.timestamp = block1.header.timestamp + BlockInterval
	block2.CollectTransactions(1)
	block2.SetMiner(addr)
	block2.Seal()

	addr = &Address{validators[4]}
	block3, _ := NewBlock(0, addr, block2)
	block3.header.timestamp = block2.header.timestamp + BlockInterval
	block3.CollectTransactions(1)
	block3.SetMiner(addr)
	block3.Seal()

	addr = &Address{validators[5]}
	block4, _ := NewBlock(0, addr, block3)
	block4.header.timestamp = block3.header.timestamp + BlockInterval
	block4.CollectTransactions(1)
	block4.SetMiner(addr)
	block4.Seal()

	err = pool.Push(block0)
	assert.NoError(t, err)
	assert.Equal(t, pool.blockCache.Len(), 0)

	err = pool.Push(block3)
	assert.Equal(t, pool.blockCache.Len(), 1)
	assert.NoError(t, err)
	err = pool.Push(block4)
	assert.Equal(t, pool.blockCache.Len(), 2)
	assert.NoError(t, err)
	err = pool.Push(block2)
	assert.Equal(t, pool.blockCache.Len(), 3)
	assert.NoError(t, err)

	err = pool.Push(block1)
	assert.NoError(t, err)
	assert.Equal(t, pool.blockCache.Len(), 0)

	bc.SetTailBlock(block0)
	nonce := bc.tailBlock.GetNonce(from.Bytes())
	assert.Equal(t, nonce, uint64(0))

	bc.SetTailBlock(block1)
	nonce = bc.tailBlock.GetNonce(from.Bytes())
	assert.Equal(t, nonce, uint64(1))
	bc.SetTailBlock(block2)
	nonce = bc.tailBlock.GetNonce(from.Bytes())
	assert.Equal(t, nonce, uint64(2))
	bc.SetTailBlock(block3)
	nonce = bc.tailBlock.GetNonce(from.Bytes())
	assert.Equal(t, nonce, uint64(3))

}
