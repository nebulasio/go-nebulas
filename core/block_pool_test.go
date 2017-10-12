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

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/crypto/keystore/ecdsa"
	"github.com/stretchr/testify/assert"
)

type Cons int

func (c Cons) VerifyBlock(block *Block) error {
	return nil
}

func TestBlockPool(t *testing.T) {
	bc := NewBlockChain(0)
	var cons Cons
	bc.SetConsensusHandler(cons)
	pool := bc.bkPool
	assert.Equal(t, pool.blockCache.Len(), 0)

	from, _ := Parse("91da63ba4ec9f6a08636d9abd443f64b4855be7fa9e44aa2")
	to, _ := Parse("c7a9cbc5d69126fa4354ac05e16d5e781c8ab4dc61850cd8")
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

	tx1 := NewTransaction(0, *from, *to, 0, 1, []byte("nas"))
	tx1.Sign()
	tx2 := NewTransaction(0, *from, *to, 0, 2, []byte("nas"))
	tx2.Sign()
	tx3 := NewTransaction(0, *from, *to, 0, 3, []byte("nas"))
	tx3.Sign()
	bc.txPool.Push(tx1)
	bc.txPool.Push(tx2)
	bc.txPool.Push(tx3)

	block1 := NewBlock(0, &Address{[]byte("coinbase")}, bc.tailBlock, bc.txPool)
	block1.CollectTransactions(1)
	block1.Seal()

	block2 := NewBlock(0, &Address{[]byte("coinbase")}, block1, bc.txPool)
	block2.CollectTransactions(1)
	block2.Seal()

	block3 := NewBlock(0, &Address{[]byte("coinbase")}, block2, bc.txPool)
	block3.CollectTransactions(1)
	block3.Seal()

	block4 := NewBlock(0, &Address{[]byte("coinbase")}, block3, bc.txPool)
	block4.CollectTransactions(1)
	block4.Seal()

	pb1, _ := block1.ToProto()
	ir1, _ := proto.Marshal(pb1)
	nb1 := new(Block)
	proto.Unmarshal(ir1, pb1)
	nb1.FromProto(pb1)

	pb2, _ := block2.ToProto()
	ir2, _ := proto.Marshal(pb2)
	nb2 := new(Block)
	proto.Unmarshal(ir2, pb2)
	nb2.FromProto(pb2)

	pb3, _ := block3.ToProto()
	ir3, _ := proto.Marshal(pb3)
	nb3 := new(Block)
	proto.Unmarshal(ir3, pb3)
	nb3.FromProto(pb3)

	pb4, _ := block4.ToProto()
	ir4, _ := proto.Marshal(pb4)
	nb4 := new(Block)
	proto.Unmarshal(ir4, pb4)
	nb4.FromProto(pb4)

	pool.addBlock(nb3)
	assert.Equal(t, pool.blockCache.Len(), 1)
	pool.addBlock(nb4)
	assert.Equal(t, pool.blockCache.Len(), 2)
	pool.addBlock(nb2)
	assert.Equal(t, pool.blockCache.Len(), 3)
	pool.addBlock(nb1)
	assert.Equal(t, pool.blockCache.Len(), 0)
}
