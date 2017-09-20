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
	"fmt"
	"reflect"
	"testing"
	"time"
	_ "time"
)

func TestTransaction(t *testing.T) {
	type fields struct {
		hash      Hash
		from      Address
		to        Address
		value     uint64
		nonce     uint64
		timestamp time.Time
		data      []byte
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			"full struct",
			fields(fields{
				[]byte("123455"),
				Address{[]byte("1235")},
				Address{[]byte("1245")},
				123,
				456,
				time.Now(),
				[]byte("hwllo"),
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &Transaction{
				hash:      tt.fields.hash,
				from:      tt.fields.from,
				to:        tt.fields.to,
				value:     tt.fields.value,
				nonce:     tt.fields.nonce,
				timestamp: tt.fields.timestamp,
				data:      tt.fields.data,
			}
			ir, err := tx.Serialize()
			fmt.Println(err)
			fmt.Printf("%v\n", ir)
			ntx := new(Transaction)
			ntx.Deserialize(ir)
			tx.timestamp = ntx.timestamp
			if !reflect.DeepEqual(tx, ntx) {
				t.Errorf("Transaction.Serialize() = %v, want %v", *tx, *ntx)
			}
		})
	}
}

func TestTransactions(t *testing.T) {
	tests := []struct {
		name    string
		txs     Transactions
		wantErr bool
	}{
		{
			"full struct",
			Transactions{
				&Transaction{
					[]byte("123455"),
					Address{[]byte("1235")},
					Address{[]byte("1245")},
					123,
					456,
					time.Now(),
					[]byte("hwllo"),
				},
				&Transaction{
					[]byte("123455"),
					Address{[]byte("1235")},
					Address{[]byte("1245")},
					123,
					456,
					time.Now(),
					[]byte("hwllo"),
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := tt.txs.Serialize()
			fmt.Println(string(got))
			var txs Transactions
			(&txs).Deserialize(got)
			tt.txs[0].timestamp = txs[0].timestamp
			if !reflect.DeepEqual(tt.txs[0], txs[0]) {
				t.Errorf("Transactions.Serialize() = %v, want %v", tt.txs[0], txs[0])
			}
			tt.txs[1].timestamp = txs[1].timestamp
			if !reflect.DeepEqual(tt.txs[1], txs[1]) {
				t.Errorf("Transactions.Serialize() = %v, want %v", tt.txs[1], txs[1])
			}
		})
	}
}
