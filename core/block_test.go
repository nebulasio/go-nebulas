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
	"reflect"
	"testing"
	"time"

	pb "github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/crypto/cipher"
	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/crypto/keystore/ecdsa"
	"github.com/stretchr/testify/assert"
)

func TestBlockHeader(t *testing.T) {
	type fields struct {
		hash       Hash
		parentHash Hash
		stateRoot  Hash
		nonce      uint64
		coinbase   *Address
		timestamp  int64
		chainID    uint32
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			"full struct",
			fields{
				[]byte("124546"),
				[]byte("344543"),
				[]byte("43656"),
				3546456,
				&Address{[]byte("hello")},
				time.Now().Unix(),
				1,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BlockHeader{
				hash:       tt.fields.hash,
				parentHash: tt.fields.parentHash,
				stateRoot:  tt.fields.stateRoot,
				nonce:      tt.fields.nonce,
				coinbase:   tt.fields.coinbase,
				timestamp:  tt.fields.timestamp,
				chainID:    tt.fields.chainID,
			}
			proto, _ := b.ToProto()
			ir, _ := pb.Marshal(proto)
			nb := new(BlockHeader)
			pb.Unmarshal(ir, proto)
			nb.FromProto(proto)
			b.timestamp = nb.timestamp
			if !reflect.DeepEqual(*b, *nb) {
				t.Errorf("Transaction.Serialize() = %v, want %v", *b, *nb)
			}
		})
	}
}

func TestBlock(t *testing.T) {
	type fields struct {
		header       *BlockHeader
		transactions Transactions
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			"full struct",
			fields{
				&BlockHeader{
					[]byte("124546"),
					[]byte("344543"),
					[]byte("43656"),
					[]byte("43656"),
					3546456,
					&Address{[]byte("hello")},
					time.Now().Unix(),
					1,
				},
				Transactions{
					&Transaction{
						[]byte("123455"),
						Address{[]byte("1235")},
						Address{[]byte("1245")},
						123,
						456,
						time.Now(),
						[]byte("hwllo"),
						1,
						uint8(cipher.SECP256K1),
						nil,
					},
					&Transaction{
						[]byte("123455"),
						Address{[]byte("1235")},
						Address{[]byte("1245")},
						123,
						456,
						time.Now(),
						[]byte("hwllo"),
						1,
						uint8(cipher.SECP256K1),
						nil,
					},
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Block{
				header:       tt.fields.header,
				transactions: tt.fields.transactions,
			}
			proto, _ := b.ToProto()
			ir, _ := pb.Marshal(proto)
			nb := new(Block)
			pb.Unmarshal(ir, proto)
			nb.FromProto(proto)
			b.header.timestamp = nb.header.timestamp
			b.transactions[0].timestamp = nb.transactions[0].timestamp
			b.transactions[1].timestamp = nb.transactions[1].timestamp
			if !reflect.DeepEqual(*b.header, *nb.header) {
				t.Errorf("Transaction.Serialize() = %v, want %v", *b.header, *nb.header)
			}
			if !reflect.DeepEqual(*b.transactions[0], *nb.transactions[0]) {
				t.Errorf("Transaction.Serialize() = %v, want %v", *b.transactions[0], *nb.transactions[0])
			}
			if !reflect.DeepEqual(*b.transactions[1], *nb.transactions[1]) {
				t.Errorf("Transaction.Serialize() = %v, want %v", *b.transactions[1], *nb.transactions[1])
			}
		})
	}
}

func TestBlock_LinkParentBlock(t *testing.T) {
	genesis := NewGenesisBlock(0)
	assert.Equal(t, genesis.Height(), uint64(1))
	block1 := &Block{
		header: &BlockHeader{
			[]byte("124546"),
			genesis.Hash(),
			[]byte("43656"),
			[]byte("43656"),
			3546456,
			&Address{[]byte("hello")},
			time.Now().Unix(),
			1,
		},
		transactions: []*Transaction{},
	}
	assert.Equal(t, block1.Height(), uint64(0))
	assert.Equal(t, block1.LinkParentBlock(genesis), true)
	assert.Equal(t, block1.Height(), uint64(2))
	assert.Equal(t, block1.ParentBlock(), genesis)
	block2 := &Block{
		header: &BlockHeader{
			[]byte("124546"),
			[]byte("1234"),
			[]byte("43656"),
			[]byte("43656"),
			3546456,
			&Address{[]byte("hello")},
			time.Now().Unix(),
			1,
		},
		transactions: []*Transaction{},
	}
	assert.Equal(t, block2.LinkParentBlock(genesis), false)
	assert.Equal(t, block2.Height(), uint64(0))
}

func TestBlock_CollectTransactions(t *testing.T) {
	bc := NewBlockChain(0)
	tail := bc.tailBlock
	assert.Panics(t, func() { tail.CollectTransactions(1) })
	block := NewBlock(0, &Address{[]byte("coinbase")}, tail, bc.txPool)

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
	tx3 := NewTransaction(0, *from, *to, 0, 0, []byte("nas"))
	tx3.Sign()
	tx4 := NewTransaction(0, *from, *to, 0, 4, []byte("nas"))
	tx4.Sign()
	tx5 := NewTransaction(0, *from, *to, 1, 3, []byte("nas"))
	tx5.Sign()
	tx6 := NewTransaction(1, *from, *to, 0, 1, []byte("nas"))
	tx6.Sign()

	bc.txPool.Push(tx1)
	bc.txPool.Push(tx2)
	bc.txPool.Push(tx3)
	bc.txPool.Push(tx4)
	bc.txPool.Push(tx5)
	bc.txPool.Push(tx6)

	assert.Equal(t, len(block.transactions), 0)
	block.CollectTransactions(bc.txPool.cache.Len())
	assert.Equal(t, len(block.transactions), 2)
	assert.Equal(t, block.txPool.cache.Len(), 1)

	assert.Equal(t, block.Sealed(), false)
	fromAcc := block.FindAccount(block.header.coinbase)
	assert.Equal(t, fromAcc.Balance, uint64(0))
	block.Seal()
	assert.Equal(t, block.Sealed(), true)
	assert.Equal(t, block.transactions[0], tx1)
	assert.Equal(t, block.transactions[1], tx2)
	assert.Equal(t, block.StateRoot().Equals(block.stateTrie.RootHash()), true)
	assert.Equal(t, block.TxsRoot().Equals(block.txsTrie.RootHash()), true)
	fromAcc = block.FindAccount(block.header.coinbase)
	assert.Equal(t, fromAcc.Balance, uint64(BlockReward))
	// mock net message
	proto, _ := block.ToProto()
	ir, _ := pb.Marshal(proto)
	nb := new(Block)
	pb.Unmarshal(ir, proto)
	nb.FromProto(proto)
	nb.LinkParentBlock(bc.tailBlock)
	assert.Nil(t, nb.Verify())
}
