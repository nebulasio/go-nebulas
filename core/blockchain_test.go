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

	"github.com/nebulasio/go-nebulas/crypto"
	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/crypto/keystore/secp256k1"
	"github.com/nebulasio/go-nebulas/util"

	"github.com/stretchr/testify/assert"
)

func signBlock(block *Block) {
	block.header.alg = keystore.SECP256K1
}

func TestBlockChain_FindCommonAncestorWithTail(t *testing.T) {
	neb := testNeb(t)
	bc := neb.chain

	coinbase11, _ := AddressParse("2fe3f9f51f9a05dd5f7c5329127f7c917917149b4e16b0b8")
	coinbase12, _ := AddressParse("1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c")
	coinbase111, _ := AddressParse("59fc526072b09af8a8ca9732dae17132c4e9127e43cf2232")
	coinbase221, _ := AddressParse("48f981ed38910f1232c1bab124f650c482a57271632db9e3")
	coinbase222, _ := AddressParse("333cb3ed8c417971845382ede3cf67a0a96270c05fe2f700")
	coinbase1111, _ := AddressParse("7da9dabedb4c6e121146fb4250a9883d6180570e63d6b080")
	coinbase11111, _ := AddressParse("98a3eed687640b75ec55bf5c9e284371bdcaeab943524d51")

	//add from reward
	block0, _ := bc.NewBlock(coinbase11)
	block0.header.timestamp = BlockInterval
	block0.Seal()
	signBlock(block0)
	assert.Nil(t, bc.BlockPool().Push(block0))
	assert.Equal(t, bc.tailBlock.Hash().String(), "9443a0846ac180932533b55420877eb3618d05f1120f3a5c54dc5d3a0b314452")

	/*
		genesis -- 0 -- 11 -- 111 -- 1111
					 \_ 12 -- 221
					       \_ 222 tail
	*/
	block11, err := bc.NewBlock(coinbase11)
	assert.Nil(t, err)
	block11.header.timestamp = BlockInterval * 2
	block12, err := bc.NewBlock(coinbase12)
	assert.Nil(t, err)
	block12.header.timestamp = BlockInterval * 3
	assert.Nil(t, block11.Seal())
	signBlock(block11)
	assert.Equal(t, block11.Hash().String(), "e1a59257cbc5cd2a0f4d14470c0d507db2b2eeb13126fdcd32bf08c787f9929c")
	assert.Nil(t, block12.Seal())
	signBlock(block12)
	assert.Equal(t, block12.Hash().String(), "71821a5a452effcca2b579a800402aeaaa3fc4364b6a01684ef7c5d1d46de35d")

	assert.Nil(t, bc.BlockPool().Push(block11))
	assert.Equal(t, bc.tailBlock.Hash(), block11.Hash())
	block111, err := bc.NewBlock(coinbase111)
	assert.Nil(t, err)
	block111.header.timestamp = BlockInterval * 4
	assert.Nil(t, block111.Seal())
	signBlock(block111)

	assert.Nil(t, bc.BlockPool().Push(block12))
	assert.Equal(t, bc.tailBlock.Hash(), block12.Hash())
	block221, _ := bc.NewBlock(coinbase221)
	block221.header.timestamp = BlockInterval * 5
	block222, _ := bc.NewBlock(coinbase222)
	block222.header.timestamp = BlockInterval * 6
	assert.Nil(t, block221.Seal())
	signBlock(block221)
	assert.Equal(t, block221.Hash().String(), "2e554cb87cff9cc5c685049afd2e45d02ba493ea0056f852c478fee2fea590ad")
	assert.Nil(t, block222.Seal())
	signBlock(block222)
	assert.Equal(t, block222.Hash().String(), "65910f3e367c931c1d6bd4a9ef4f0d8b6cb44d7a429f19abf1f53a636732b3c8")

	assert.Nil(t, bc.BlockPool().Push(block111))
	block1111, _ := bc.NewBlock(coinbase1111)
	block1111.header.timestamp = BlockInterval * 7
	assert.Nil(t, block1111.Seal())
	signBlock(block1111)

	assert.Nil(t, bc.BlockPool().Push(block221))
	assert.Nil(t, bc.BlockPool().Push(block222))
	assert.Equal(t, bc.tailBlock.Hash(), block221.Hash())

	common2, err := bc.FindCommonAncestorWithTail(block222)
	assert.Nil(t, err)
	assert.Equal(t, common2.String(), block12.String())

	common3, err := bc.FindCommonAncestorWithTail(block111)
	assert.Nil(t, err)
	assert.Equal(t, common3.Hash(), block0.Hash())

	common4, err := bc.FindCommonAncestorWithTail(bc.tailBlock)
	assert.Nil(t, err)
	assert.Equal(t, common4.Hash(), bc.tailBlock.Hash())

	common5, err := bc.FindCommonAncestorWithTail(block12)
	assert.Nil(t, err)
	assert.Equal(t, common5.Height(), block12.Height())

	result := bc.Dump(4)
	assert.Equal(t, result, "["+block221.String()+","+block12.String()+","+block0.String()+","+bc.genesisBlock.String()+"]")

	assert.Nil(t, bc.BlockPool().Push(block1111))
	tails := bc.DetachedTailBlocks()
	for _, v := range tails {
		if v.Hash().Equals(block221.Hash()) ||
			v.Hash().Equals(block222.Hash()) ||
			v.Hash().Equals(block1111.Hash()) {
			continue
		}
		assert.Equal(t, true, false)
	}
	assert.Equal(t, len(tails), 3)

	block11111, _ := bc.NewBlock(coinbase11111)
	block11111.header.timestamp = BlockInterval * 8
	assert.Nil(t, block11111.Seal())
	signBlock(block11111)
	assert.Nil(t, bc.BlockPool().Push(block11111))
}

func TestBlockChain_SimulateTransactionExecution(t *testing.T) {
	priv := secp256k1.GeneratePrivateKey()
	pubdata, _ := priv.PublicKey().Encoded()
	from, _ := NewAddressFromPublicKey(pubdata)
	to := &Address{from.address}

	payload, err := NewBinaryPayload(nil).ToBytes()
	assert.Nil(t, err)

	neb := testNeb(t)
	bc := neb.chain
	gasLimit, _ := util.NewUint128FromInt(200000)
	tx, _ := NewTransaction(bc.ChainID(), from, to, util.NewUint128(), 1, TxPayloadBinaryType, payload, TransactionGasPrice, gasLimit)

	expectedGasUsed, _, _ := tx.CalculateMinGasExpected(nil)

	gasUsed, _, exeErr, err := bc.SimulateTransactionExecution(tx)
	assert.Nil(t, err)
	assert.Equal(t, ErrInsufficientBalance, exeErr)
	assert.Equal(t, expectedGasUsed, gasUsed)
}

func TestTailBlock(t *testing.T) {
	neb := testNeb(t)
	bc := neb.chain
	block, err := bc.LoadTailFromStorage()
	assert.Nil(t, err)
	assert.Equal(t, bc.tailBlock.Hash(), block.Hash())
}

func TestGetPrice(t *testing.T) {
	neb := testNeb(t)
	bc := neb.chain
	assert.Equal(t, bc.GasPrice(), TransactionGasPrice)

	ks := keystore.DefaultKS
	from := mockAddress()
	key, err := ks.GetUnlocked(from.String())
	assert.Nil(t, err)
	signature, err := crypto.NewSignature(keystore.SECP256K1)
	assert.Nil(t, err)
	signature.InitSign(key.(keystore.PrivateKey))
	block, err := bc.NewBlock(from)
	assert.Nil(t, err)
	GasPriceDetla, _ := util.NewUint128FromInt(1)
	lowerGasPrice, err := TransactionGasPrice.Sub(GasPriceDetla)
	assert.Nil(t, err)
	gasLimit, _ := util.NewUint128FromInt(200000)
	tx1, _ := NewTransaction(bc.ChainID(), from, from, util.NewUint128(), 1, TxPayloadBinaryType, []byte("nas"), lowerGasPrice, gasLimit)
	tx1.Sign(signature)
	tx2, _ := NewTransaction(bc.ChainID(), from, from, util.NewUint128(), 2, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, gasLimit)
	tx2.Sign(signature)
	block.transactions = append(block.transactions, tx1)
	block.transactions = append(block.transactions, tx2)
	block.Seal()
	block.Sign(signature)
	bc.SetTailBlock(block)
	assert.Equal(t, bc.GasPrice(), lowerGasPrice)
}
