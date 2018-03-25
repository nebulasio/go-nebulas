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

	coinbase12, _ := AddressParse("n1FkntVUMPAsESuCAAPK711omQk19JotBjM")
	coinbase11, _ := AddressParse("n1JNHZJEUvfBYfjDRD14Q73FX62nJAzXkMR")
	coinbase221, _ := AddressParse("n1Kjom3J4KPsHKKzZ2xtt8Lc9W5pRDjeLcW")
	coinbase111, _ := AddressParse("n1TV3sU6jyzR4rJ1D7jCAmtVGSntJagXZHC")
	coinbase222, _ := AddressParse("n1WwqBXVMuYC3mFCEEuFFtAXad6yxqj4as4")
	coinbase1111, _ := AddressParse("n1Zn6iyyQRhqthmCfqGBzWfip1Wx8wEvtrJ")
	coinbase11111, _ := AddressParse("n1Zn6iyyQRhqthmCfqGBzWfip1Wx8wEvtrJ")

	//add from reward
	block0, _ := bc.NewBlock(coinbase11)
	block0.header.timestamp = BlockInterval
	block0.Seal()
	signBlock(block0)
	assert.Nil(t, bc.BlockPool().Push(block0))
	assert.Equal(t, bc.tailBlock.Hash().String(), "1111e101a97e0fb05ee1c96db0eba69b4a491af15f820d8df06f65379a2de9ef")

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
	assert.Nil(t, block12.Seal())
	signBlock(block12)

	assert.Nil(t, bc.BlockPool().Push(block11))
	assert.Equal(t, bc.tailBlock.Hash(), block11.Hash())
	block111, err := bc.NewBlock(coinbase111)
	assert.Nil(t, err)
	block111.header.timestamp = BlockInterval * 4
	assert.Nil(t, block111.Seal())
	signBlock(block111)

	assert.Equal(t, bc.tailBlock.Hash(), block11.Hash())

	assert.Nil(t, bc.BlockPool().Push(block12))
	assert.Equal(t, bc.tailBlock.Hash(), block12.Hash())

	block221, _ := bc.NewBlock(coinbase221)
	block221.header.timestamp = BlockInterval * 5
	block222, _ := bc.NewBlock(coinbase222)
	block222.header.timestamp = BlockInterval * 6
	assert.Nil(t, block221.Seal())
	signBlock(block221)
	assert.Equal(t, block221.Hash().String(), "0da11c6b15932b426f264e0a862ebbee318cc2ade632d480d43e3e2615013495")
	assert.Nil(t, block222.Seal())
	signBlock(block222)
	assert.Equal(t, block222.Hash().String(), "078a5acc899a7f209b4c816a8bfacbada91f54617f2f57b384399c74feb97639")

	assert.Nil(t, bc.BlockPool().Push(block111))
	block1111, _ := bc.NewBlock(coinbase1111)
	block1111.header.timestamp = BlockInterval * 7
	assert.Nil(t, block1111.Seal())
	signBlock(block1111)

	assert.Nil(t, bc.BlockPool().Push(block221))
	assert.Nil(t, bc.BlockPool().Push(block222))
	assert.Equal(t, bc.tailBlock.Hash(), block222.Hash())

	common2, err := bc.FindCommonAncestorWithTail(block221)
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
	assert.Equal(t, result, "["+block222.String()+","+block12.String()+","+block0.String()+","+bc.genesisBlock.String()+"]")

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

	expectedGasUsed, _ := util.NewUint128FromInt(20000)

	result, err := bc.SimulateTransactionExecution(tx)
	assert.Nil(t, err)
	assert.Equal(t, ErrInsufficientBalance, result.Err)
	assert.Equal(t, expectedGasUsed, result.GasUsed)
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
