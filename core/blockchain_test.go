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
	bc, _ := NewBlockChain(testNeb())
	var c MockConsensus
	bc.SetConsensusHandler(c)

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
	block0.header.timestamp = BlockInterval
	block0.SetMiner(from)
	block0.Seal()
	assert.Nil(t, bc.BlockPool().Push(BlockFromNetwork(block0)))
	bc.SetTailBlock(block0)
	assert.Equal(t, bc.latestIrreversibleBlock, bc.genesisBlock)

	coinbase11 := &Address{[]byte("012345678901234567890011")}
	coinbase12 := &Address{[]byte("012345678901234567890012")}
	coinbase111 := &Address{[]byte("012345678901234567890111")}
	coinbase221 := &Address{[]byte("012345678901234567890221")}
	coinbase222 := &Address{[]byte("012345678901234567890222")}
	coinbase1111 := &Address{[]byte("012345678901234567891111")}
	coinbase11111 := &Address{[]byte("012345678901234567811111")}
	/*
		genesis -- 0 -- 11 -- 111 -- 1111
					 \_ 12 -- 221
					       \_ 222 tail
	*/
	block11, _ := bc.NewBlock(coinbase11)
	block11.header.timestamp = BlockInterval * 2
	block12, _ := bc.NewBlock(coinbase12)
	block12.header.timestamp = BlockInterval * 3
	block11.SetMiner(coinbase11)
	block11.Seal()
	block12.SetMiner(coinbase12)
	block12.Seal()
	assert.Nil(t, bc.BlockPool().Push(BlockFromNetwork(block11)))
	assert.Nil(t, bc.BlockPool().Push(BlockFromNetwork(block12)))
	bc.SetTailBlock(block12)
	assert.Equal(t, bc.latestIrreversibleBlock, bc.genesisBlock)
	bc.SetTailBlock(block11)
	assert.Equal(t, bc.latestIrreversibleBlock, bc.genesisBlock)
	block111, _ := bc.NewBlock(coinbase111)
	block111.header.timestamp = BlockInterval * 4
	block111.SetMiner(coinbase111)
	block111.Seal()
	assert.Nil(t, bc.BlockPool().Push(BlockFromNetwork(block111)))
	bc.SetTailBlock(block12)
	assert.Equal(t, bc.latestIrreversibleBlock, bc.genesisBlock)
	block221, _ := bc.NewBlock(coinbase221)
	block221.header.timestamp = BlockInterval * 5
	block222, _ := bc.NewBlock(coinbase222)
	block222.header.timestamp = BlockInterval * 6
	block221.SetMiner(coinbase221)
	block221.Seal()
	block222.SetMiner(coinbase222)
	block222.Seal()
	assert.Nil(t, bc.BlockPool().Push(BlockFromNetwork(block221)))
	assert.Nil(t, bc.BlockPool().Push(BlockFromNetwork(block222)))
	bc.SetTailBlock(block111)
	assert.Equal(t, bc.latestIrreversibleBlock, bc.genesisBlock)
	block1111, _ := bc.NewBlock(coinbase1111)
	block1111.header.timestamp = BlockInterval * 7
	block1111.SetMiner(coinbase1111)
	block1111.Seal()
	assert.Nil(t, bc.BlockPool().Push(BlockFromNetwork(block1111)))
	bc.SetTailBlock(block222)
	assert.Equal(t, bc.latestIrreversibleBlock, bc.genesisBlock)
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
	test := &Block{header: &BlockHeader{coinbase: &Address{}}, miner: &Address{}}
	_, err := bc.FindCommonAncestorWithTail(BlockFromNetwork(test))
	assert.Equal(t, err, ErrMissingParentBlock)
	common1, err := bc.FindCommonAncestorWithTail(BlockFromNetwork(block1111))
	assert.Nil(t, err)
	assert.Equal(t, BlockFromNetwork(common1), BlockFromNetwork(block0))
	common2, err := bc.FindCommonAncestorWithTail(BlockFromNetwork(block221))
	assert.Nil(t, err)
	assert.Equal(t, BlockFromNetwork(common2), BlockFromNetwork(block12))
	common3, err := bc.FindCommonAncestorWithTail(BlockFromNetwork(block222))
	assert.Nil(t, err)
	assert.Equal(t, BlockFromNetwork(common3), BlockFromNetwork(block222))
	common4, err := bc.FindCommonAncestorWithTail(BlockFromNetwork(bc.tailBlock))
	assert.Nil(t, err)
	assert.Equal(t, BlockFromNetwork(common4), BlockFromNetwork(bc.tailBlock))
	common5, err := bc.FindCommonAncestorWithTail(BlockFromNetwork(block12))
	assert.Nil(t, err)
	assert.Equal(t, BlockFromNetwork(common5), BlockFromNetwork(block12))

	result := bc.Dump(4)
	assert.Equal(t, result, "["+block222.String()+","+block12.String()+","+block0.String()+","+bc.genesisBlock.String()+"]")

	bc.SetTailBlock(block1111)
	assert.Equal(t, bc.latestIrreversibleBlock, bc.genesisBlock)

	block11111, _ := bc.NewBlock(coinbase11111)
	block11111.header.timestamp = BlockInterval * 8
	block11111.SetMiner(coinbase11111)
	block11111.Seal()
	assert.Nil(t, bc.BlockPool().Push(BlockFromNetwork(block11111)))
	bc.SetTailBlock(block11111)

	bc.updateLatestIrreversibleBlock(block11111)
	assert.Equal(t, bc.latestIrreversibleBlock.Hash(), block0.Hash())
}

func TestBlockChain_FetchDescendantInCanonicalChain(t *testing.T) {
	bc, _ := NewBlockChain(testNeb())
	var c MockConsensus
	bc.SetConsensusHandler(c)
	coinbase := &Address{[]byte("012345678901234567890000")}
	/*
		genesisi -- 1 - 2 - 3 - 4 - 5 - 6
		         \_ block - block1
	*/
	block, _ := bc.NewBlock(coinbase)
	block.header.timestamp = BlockInterval
	block.SetMiner(coinbase)
	block.Seal()
	bc.BlockPool().Push(block)
	bc.SetTailBlock(block)

	block1, _ := bc.NewBlock(coinbase)
	block1.header.timestamp = BlockInterval * 2
	block1.SetMiner(coinbase)
	block1.Seal()
	bc.BlockPool().Push(block1)
	bc.SetTailBlock(block1)

	var blocks []*Block
	for i := 0; i < 6; i++ {
		block, _ := bc.NewBlock(coinbase)
		block.header.timestamp = BlockInterval * int64(i+3)
		blocks = append(blocks, block)
		block.SetMiner(coinbase)
		block.Seal()
		bc.BlockPool().Push(block)
		bc.SetTailBlock(block)
	}
	blocks24, _ := bc.FetchDescendantInCanonicalChain(3, blocks[0])
	assert.Equal(t, BlockFromNetwork(blocks24[0]), BlockFromNetwork(blocks[1]))
	assert.Equal(t, BlockFromNetwork(blocks24[1]), BlockFromNetwork(blocks[2]))
	assert.Equal(t, BlockFromNetwork(blocks24[2]), BlockFromNetwork(blocks[3]))
	blocks46, _ := bc.FetchDescendantInCanonicalChain(10, blocks[2])
	assert.Equal(t, len(blocks46), 3)
	assert.Equal(t, BlockFromNetwork(blocks46[0]), BlockFromNetwork(blocks[3]))
	assert.Equal(t, BlockFromNetwork(blocks46[1]), BlockFromNetwork(blocks[4]))
	assert.Equal(t, BlockFromNetwork(blocks46[2]), BlockFromNetwork(blocks[5]))
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

	bc, _ := NewBlockChain(testNeb())
	tx := NewTransaction(bc.ChainID(), from, to, util.NewUint128FromInt(0), 1, TxPayloadBinaryType, payload, TransactionGasPrice, util.NewUint128FromInt(200000))

	_, err = bc.EstimateGas(tx)
	assert.Nil(t, err)
}

func TestTailBlock(t *testing.T) {
	bc, err := NewBlockChain(testNeb())
	assert.Nil(t, err)
	block, err := bc.loadTailFromStorage()
	assert.Nil(t, err)
	assert.Equal(t, bc.tailBlock, block)
}

func TestGetPrice(t *testing.T) {
	bc, err := NewBlockChain(testNeb())
	assert.Nil(t, err)
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
	lowerGasPrice := util.NewUint128FromBigInt(util.NewUint128().Sub(TransactionGasPrice.Int, util.NewUint128FromInt(1).Int))
	tx1 := NewTransaction(bc.ChainID(), from, from, util.NewUint128(), 1, TxPayloadBinaryType, []byte("nas"), lowerGasPrice, util.NewUint128FromInt(200000))
	tx1.Sign(signature)
	tx2 := NewTransaction(bc.ChainID(), from, from, util.NewUint128(), 2, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, util.NewUint128FromInt(200000))
	tx2.Sign(signature)
	block.transactions = append(block.transactions, tx1)
	block.transactions = append(block.transactions, tx2)
	block.miner = from
	block.Seal()
	block.Sign(signature)
	bc.SetTailBlock(block)
	bc.storeBlockToStorage(block)
	block, err = bc.NewBlock(from)
	assert.Nil(t, err)
	block.miner = from
	block.Seal()
	block.Sign(signature)
	bc.SetTailBlock(block)
	bc.storeBlockToStorage(block)
	assert.Equal(t, bc.GasPrice(), lowerGasPrice)
}
