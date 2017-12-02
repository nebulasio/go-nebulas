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

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/crypto"
	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/crypto/keystore/secp256k1"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/nebulasio/go-nebulas/util/byteutils"
)

func TestTransaction(t *testing.T) {
	type fields struct {
		hash      byteutils.Hash
		from      *Address
		to        *Address
		value     *util.Uint128
		nonce     uint64
		timestamp int64
		alg       uint8
		data      *corepb.Data
		gasPrice  *util.Uint128
		gasLimit  *util.Uint128
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			"full struct",
			fields(fields{
				[]byte("123455"),
				&Address{[]byte("1235")},
				&Address{[]byte("1245")},
				util.NewUint128(),
				456,
				time.Now().Unix(),
				12,
				&corepb.Data{Type: TxPayloadBinaryType, Payload: []byte("hwllo")},
				util.NewUint128(),
				util.NewUint128(),
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
				alg:       tt.fields.alg,
				data:      tt.fields.data,
				gasPrice:  tt.fields.gasPrice,
				gasLimit:  tt.fields.gasLimit,
			}
			msg, _ := tx.ToProto()
			ir, _ := proto.Marshal(msg)
			ntx := new(Transaction)
			nMsg := new(corepb.Transaction)
			proto.Unmarshal(ir, nMsg)
			ntx.FromProto(nMsg)
			ntx.timestamp = tx.timestamp
			if !reflect.DeepEqual(tx, ntx) {
				t.Errorf("Transaction.Serialize() = %v, want %v", *tx, *ntx)
			}
		})
	}
}

func TestTransactionVerify(t *testing.T) {
	testCount := 3
	type testTx struct {
		name   string
		tx     *Transaction
		signer keystore.Signature
		count  int
	}

	tests := []testTx{}
	ks := keystore.DefaultKS

	for index := 0; index < testCount; index++ {
		priv1 := secp256k1.GeneratePrivateKey()
		pubdata1, _ := priv1.PublicKey().Encoded()
		from, _ := NewAddressFromPublicKey(pubdata1)
		ks.SetKey(from.ToHex(), priv1, []byte("passphrase"))
		ks.Unlock(from.ToHex(), []byte("passphrase"), time.Second*60*60*24*365)
		key1, _ := ks.GetUnlocked(from.ToHex())
		signature1, _ := crypto.NewSignature(keystore.SECP256K1)
		signature1.InitSign(key1.(keystore.PrivateKey))

		toPriv := secp256k1.GeneratePrivateKey()
		toPub, _ := toPriv.PublicKey().Encoded()
		toAddr, _ := NewAddressFromPublicKey(toPub)
		tx := NewTransaction(1, from, toAddr, util.NewUint128(), 10, TxPayloadBinaryType, []byte("datadata"), TransactionGasPrice, util.NewUint128FromInt(200000))

		test := testTx{string(index), tx, signature1, 1}
		tests = append(tests, test)
	}
	for _, tt := range tests {
		for index := 0; index < tt.count; index++ {
			t.Run(tt.name, func(t *testing.T) {
				err := tt.tx.Sign(tt.signer)
				if err != nil {
					t.Errorf("Sign() error = %v", err)
					return
				}
				err = tt.tx.Verify(tt.tx.chainID)
				if err != nil {
					t.Errorf("verify failed:%s", err)
					return
				}
			})
		}
	}
}
