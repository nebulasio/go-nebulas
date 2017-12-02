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
	storage, _ := storage.NewMemoryStorage()
	bc, _ := NewBlockChain(0, storage)
	var c MockConsensus
	bc.SetConsensusHandler(c)

	ks := keystore.DefaultKS
	priv := secp256k1.GeneratePrivateKey()
	pubdata, _ := priv.PublicKey().Encoded()
	from, _ := NewAddressFromPublicKey(pubdata)
	to := &Address{from.address}
	ks.SetKey(from.ToHex(), priv, []byte("passphrase"))
	ks.Unlock(from.ToHex(), []byte("passphrase"), time.Second*60*60*24*365)

	key, _ := ks.GetUnlocked(from.ToHex())
	signature, _ := crypto.NewSignature(keystore.SECP256K1)
	signature.InitSign(key.(keystore.PrivateKey))

	//add from reward
	block0 := bc.NewBlock(from)
	block0.Seal()
	bc.BlockPool().Push(block0)
	bc.SetTailBlock(block0)

	tx1 := NewTransaction(0, from, to, util.NewUint128FromInt(1), 1, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, util.NewUint128FromInt(200000))
	tx1.Sign(signature)
	tx2 := NewTransaction(0, from, to, util.NewUint128FromInt(1), 1, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, util.NewUint128FromInt(200000))
	tx2.timestamp = tx1.timestamp + 1
	tx2.Sign(signature)
	tx3 := NewTransaction(0, from, to, util.NewUint128FromInt(1), 2, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, util.NewUint128FromInt(200000))
	tx3.timestamp = tx3.timestamp + 1
	tx3.Sign(signature)
	bc.txPool.Push(tx1)
	bc.txPool.Push(tx2)
	bc.txPool.Push(tx3)

	coinbase11 := &Address{[]byte("012345678901234567890011")}
	coinbase12 := &Address{[]byte("012345678901234567890012")}
	coinbase111 := &Address{[]byte("012345678901234567890111")}
	coinbase221 := &Address{[]byte("012345678901234567890221")}
	coinbase222 := &Address{[]byte("012345678901234567890222")}
	coinbase1111 := &Address{[]byte("012345678901234567891111")}
	/*
		genesis -- 0 -- 11 -- 111 -- 1111
					 \_ 12 -- 221
					       \_ 222 tail
	*/
	block11 := bc.NewBlock(coinbase11)
	block12 := bc.NewBlock(coinbase12)
	block11.CollectTransactions(1)
	block11.Seal()
	block12.CollectTransactions(1)
	block12.Seal()
	bc.BlockPool().Push(block11)
	bc.BlockPool().Push(block12)
	bc.SetTailBlock(block12)
	assert.Equal(t, bc.txPool.cache.Len(), 1)
	bc.SetTailBlock(block11)
	assert.Equal(t, bc.txPool.cache.Len(), 2)
	block111 := bc.NewBlock(coinbase111)
	block111.CollectTransactions(0)
	block111.Seal()
	bc.BlockPool().Push(block111)
	bc.SetTailBlock(block12)
	block221 := bc.NewBlock(coinbase221)
	block222 := bc.NewBlock(coinbase222)
	block221.CollectTransactions(0)
	block221.Seal()
	block222.CollectTransactions(0)
	block222.Seal()
	bc.BlockPool().Push(block221)
	bc.BlockPool().Push(block222)
	bc.SetTailBlock(block111)
	block1111 := bc.NewBlock(coinbase1111)
	block1111.CollectTransactions(0)
	block1111.Seal()
	bc.BlockPool().Push(block1111)
	bc.SetTailBlock(block222)
	test := &Block{
		header: &BlockHeader{
			coinbase: &Address{},
		},
	}
	_, err := bc.FindCommonAncestorWithTail(BlockFromNetwork(test))
	assert.NotNil(t, err)
	common1, _ := bc.FindCommonAncestorWithTail(BlockFromNetwork(block1111))
	assert.Equal(t, BlockFromNetwork(common1), BlockFromNetwork(block0))
	common2, _ := bc.FindCommonAncestorWithTail(BlockFromNetwork(block221))
	assert.Equal(t, BlockFromNetwork(common2), BlockFromNetwork(block12))
	common3, _ := bc.FindCommonAncestorWithTail(BlockFromNetwork(block222))
	assert.Equal(t, BlockFromNetwork(common3), BlockFromNetwork(block222))
	common4, _ := bc.FindCommonAncestorWithTail(BlockFromNetwork(bc.tailBlock))
	assert.Equal(t, BlockFromNetwork(common4), BlockFromNetwork(bc.tailBlock))
	common5, _ := bc.FindCommonAncestorWithTail(BlockFromNetwork(block12))
	assert.Equal(t, BlockFromNetwork(common5), BlockFromNetwork(block12))
}

func TestBlockChain_FetchDescendantInCanonicalChain(t *testing.T) {
	storage, _ := storage.NewMemoryStorage()
	bc, _ := NewBlockChain(0, storage)
	var c MockConsensus
	bc.SetConsensusHandler(c)
	coinbase := &Address{[]byte("012345678901234567890000")}
	/*
		genesisi -- 1 - 2 - 3 - 4 - 5 - 6
		         \_ block - block1
	*/
	block := bc.NewBlock(coinbase)
	block.header.timestamp = 0
	block.CollectTransactions(0)
	block.Seal()
	bc.BlockPool().Push(block)
	block1 := bc.NewBlock(coinbase)
	block1.header.timestamp = 1
	block1.CollectTransactions(0)
	block1.Seal()
	bc.BlockPool().Push(block1)

	var blocks []*Block
	for i := 0; i < 6; i++ {
		block := bc.NewBlock(coinbase)
		if i > 0 {
			block.header.timestamp = blocks[i-1].header.timestamp + 1
		}
		blocks = append(blocks, block)
		block.CollectTransactions(0)
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
	_, err := bc.FetchDescendantInCanonicalChain(3, block)
	assert.NotNil(t, err)
	blocks0, err0 := bc.FetchDescendantInCanonicalChain(3, blocks[5])
	assert.Equal(t, len(blocks0), 0)
	assert.Nil(t, err0)
}
