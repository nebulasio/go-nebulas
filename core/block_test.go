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
)

func TestBlockHeader(t *testing.T) {
	type fields struct {
		hash       Hash
		parentHash Hash
		stateRoot  Hash
		nonce      uint64
		coinbase   *Address
		timestamp  int64
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
			}
			ir, _ := b.Serialize()
			nb := new(BlockHeader)
			nb.Deserialize(ir)
			b.timestamp = nb.timestamp
			if !reflect.DeepEqual(*b, *nb) {
				t.Errorf("Transaction.Serialize() = %v, want %v", *b, *nb)
			}
		})
	}
}

func TestBlock_Deserialize(t *testing.T) {
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
					3546456,
					&Address{[]byte("hello")},
					time.Now().Unix(),
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
			ir, _ := b.Serialize()
			nb := new(Block)
			nb.Deserialize(ir)
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
