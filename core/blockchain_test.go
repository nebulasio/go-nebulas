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

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

func BlockFromNetwork(block *Block) *Block {
	pb, _ := block.ToProto()
	ir, _ := proto.Marshal(pb)
	proto.Unmarshal(ir, pb)
	b := new(Block)
	b.FromProto(pb)
	return b
}

func TestBlockChain_FindCommonAncestorWithTail(t *testing.T) {
	bc := testNeb(t).chain

	ks := keystore.DefaultKS
	priv := secp256k1.GeneratePrivateKey()
	pubdata, _ := priv.PublicKey().Encoded()
	from, _ := NewAddressFromPublicKey(pubdata)
	ks.SetKey(from.String(), priv, []byte("passphrase"))
	ks.Unlock(from.String(), []byte("passphrase"), time.Second*60*60*24*365)

	key, _ := ks.GetUnlocked(from.String())
	signature, _ := crypto.NewSignature(keystore.SECP256K1)
	signature.InitSign(key.(keystore.PrivateKey))

	//add from reward
	block0, _ := bc.NewBlock(from)
	consensusState, err := bc.tailBlock.NextConsensusState(BlockInterval)
	assert.Nil(t, err)
	block0.LoadConsensusState(consensusState)
	block0.Seal()
	assert.Nil(t, bc.BlockPool().Push(block0))
	bc.SetTailBlock(block0)
	assert.Equal(t, bc.lib, bc.genesisBlock)

	coinbase11 := mockAddress()
	coinbase12 := mockAddress()
	coinbase111 := mockAddress()
	coinbase221 := mockAddress()
	coinbase222 := mockAddress()
	coinbase1111 := mockAddress()
	coinbase11111 := mockAddress()
	/*
		genesis -- 0 -- 11 -- 111 -- 1111
					 \_ 12 -- 221
					       \_ 222 tail
	*/
	block11, _ := bc.NewBlock(coinbase11)
	consensusState, err = bc.tailBlock.NextConsensusState(BlockInterval * 2)
	assert.Nil(t, err)
	block11.LoadConsensusState(consensusState)
	block11.Seal()
	block12, _ := bc.NewBlock(coinbase12)
	consensusState, err = bc.tailBlock.NextConsensusState(BlockInterval * 3)
	assert.Nil(t, err)
	block12.LoadConsensusState(consensusState)
	block12.Seal()
	assert.Nil(t, bc.BlockPool().Push(block11))
	assert.Nil(t, bc.BlockPool().Push(block12))
	bc.SetTailBlock(block12)
	assert.Equal(t, bc.lib, bc.genesisBlock)
	bc.SetTailBlock(block11)
	assert.Equal(t, bc.lib, bc.genesisBlock)
	block111, _ := bc.NewBlock(coinbase111)
	consensusState, err = bc.tailBlock.NextConsensusState(BlockInterval * 4)
	assert.Nil(t, err)
	block111.LoadConsensusState(consensusState)
	block111.Seal()
	assert.Nil(t, bc.BlockPool().Push(block111))
	bc.SetTailBlock(block12)
	assert.Equal(t, bc.lib, bc.genesisBlock)
	block221, _ := bc.NewBlock(coinbase221)
	consensusState, err = bc.tailBlock.NextConsensusState(BlockInterval * 5)
	assert.Nil(t, err)
	block221.LoadConsensusState(consensusState)
	block221.Seal()
	block222, _ := bc.NewBlock(coinbase222)
	consensusState, err = bc.tailBlock.NextConsensusState(BlockInterval * 6)
	assert.Nil(t, err)
	block222.LoadConsensusState(consensusState)
	block222.Seal()
	assert.Nil(t, bc.BlockPool().Push(block221))
	assert.Nil(t, bc.BlockPool().Push(block222))
	bc.SetTailBlock(block111)
	assert.Equal(t, bc.lib, bc.genesisBlock)
	block1111, _ := bc.NewBlock(coinbase1111)
	consensusState, err = bc.tailBlock.NextConsensusState(BlockInterval * 7)
	assert.Nil(t, err)
	block1111.LoadConsensusState(consensusState)
	block1111.Seal()
	assert.Nil(t, bc.BlockPool().Push(block1111))
	bc.SetTailBlock(block222)
	assert.Equal(t, bc.lib, bc.genesisBlock)
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
	common1, err := bc.FindCommonAncestorWithTail(BlockFromNetwork(block1111))
	assert.Nil(t, err)
	assert.Equal(t, common1.Hash(), block0.Hash())
	common2, err := bc.FindCommonAncestorWithTail(BlockFromNetwork(block221))
	assert.Nil(t, err)
	assert.Equal(t, common2.Hash(), block12.Hash())
	common3, err := bc.FindCommonAncestorWithTail(BlockFromNetwork(block222))
	assert.Nil(t, err)
	assert.Equal(t, common3.Hash(), block222.Hash())
	common4, err := bc.FindCommonAncestorWithTail(BlockFromNetwork(bc.tailBlock))
	assert.Nil(t, err)
	assert.Equal(t, common4.Hash(), bc.tailBlock.Hash())
	common5, err := bc.FindCommonAncestorWithTail(BlockFromNetwork(block12))
	assert.Nil(t, err)
	assert.Equal(t, common5.Hash(), block12.Hash())

	result := bc.Dump(4)
	assert.Equal(t, result, "["+block222.String()+","+block12.String()+","+block0.String()+","+bc.genesisBlock.String()+"]")

	bc.SetTailBlock(block1111)
	assert.Equal(t, bc.lib, bc.genesisBlock)

	block11111, _ := bc.NewBlock(coinbase11111)
	consensusState, err = bc.tailBlock.NextConsensusState(BlockInterval * 8)
	assert.Nil(t, err)
	block11111.LoadConsensusState(consensusState)
	block11111.Seal()
	assert.Nil(t, bc.BlockPool().Push(block11111))
	bc.SetTailBlock(block11111)
}

func TestBlockChain_FetchDescendantInCanonicalChain(t *testing.T) {
	bc := testNeb(t).chain
	coinbase := mockAddress()
	/*
		genesisi -- 1 - 2 - 3 - 4 - 5 - 6
		         \_ block - block1
	*/
	block, _ := bc.NewBlock(coinbase)
	consensusState, err := bc.tailBlock.NextConsensusState(BlockInterval)
	assert.Nil(t, err)
	block.LoadConsensusState(consensusState)
	block.Seal()
	bc.BlockPool().Push(block)
	bc.SetTailBlock(block)

	block1, _ := bc.NewBlock(coinbase)
	consensusState, err = bc.tailBlock.NextConsensusState(BlockInterval * 2)
	assert.Nil(t, err)
	block1.LoadConsensusState(consensusState)
	block1.Seal()
	bc.BlockPool().Push(block1)
	bc.SetTailBlock(block1)

	var blocks []*Block
	for i := 0; i < 6; i++ {
		block, _ := bc.NewBlock(coinbase)
		blocks = append(blocks, block)
		consensusState, err = bc.tailBlock.NextConsensusState(BlockInterval * int64(i+3))
		assert.Nil(t, err)
		block.LoadConsensusState(consensusState)
		block.Seal()
		bc.BlockPool().Push(block)
		bc.SetTailBlock(block)
	}
	blocks24, _ := bc.FetchDescendantInCanonicalChain(3, blocks[0])
	assert.Equal(t, blocks24[0].Hash(), blocks[1].Hash())
	assert.Equal(t, blocks24[1].Hash(), blocks[2].Hash())
	assert.Equal(t, blocks24[2].Hash(), blocks[3].Hash())
	blocks46, _ := bc.FetchDescendantInCanonicalChain(10, blocks[2])
	assert.Equal(t, len(blocks46), 3)
	assert.Equal(t, blocks46[0].Hash(), blocks[3].Hash())
	assert.Equal(t, blocks46[1].Hash(), blocks[4].Hash())
	assert.Equal(t, blocks46[2].Hash(), blocks[5].Hash())
	blocks13, _ := bc.FetchDescendantInCanonicalChain(3, bc.genesisBlock)
	assert.Equal(t, len(blocks13), 3)
	blocks0, err0 := bc.FetchDescendantInCanonicalChain(3, blocks[5])
	assert.Equal(t, len(blocks0), 0)
	assert.Nil(t, err0)
}

func TestBlockChain_EstimateGas(t *testing.T) {
	priv := secp256k1.GeneratePrivateKey()
	pubdata, _ := priv.PublicKey().Encoded()
	from, _ := NewAddressFromPublicKey(pubdata)
	to := &Address{from.address}

	payload, err := NewBinaryPayload(nil).ToBytes()
	assert.Nil(t, err)

	bc := testNeb(t).chain
	gasLimit, _ := util.NewUint128FromInt(200000)
	tx, _ := NewTransaction(bc.ChainID(), from, to, util.NewUint128(), 1, TxPayloadBinaryType, payload, TransactionGasPrice, gasLimit)

	_, err = bc.EstimateGas(tx)
	assert.Nil(t, err)
}

func TestTailBlock(t *testing.T) {
	bc := testNeb(t).chain
	block, err := bc.LoadTailFromStorage()
	assert.Nil(t, err)
	assert.Equal(t, bc.tailBlock, block)
}

func TestGetPrice(t *testing.T) {
	bc := testNeb(t).chain
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
	bc.StoreBlockToStorage(block)
	block, err = bc.NewBlock(from)
	assert.Nil(t, err)
	block.Seal()
	block.Sign(signature)
	bc.SetTailBlock(block)
	bc.StoreBlockToStorage(block)
	assert.Equal(t, bc.GasPrice(), lowerGasPrice)
}
