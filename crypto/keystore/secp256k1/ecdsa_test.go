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

package secp256k1

import (
	"testing"

	"reflect"

	"crypto/ecdsa"
	"crypto/rand"
	"io"

	"github.com/nebulasio/go-nebulas/crypto/hash"
	"github.com/nebulasio/go-nebulas/util/byteutils"
)

func TestFromECDSAPri(t *testing.T) {
	priv := NewECDSAPrivateKey()
	_, err := FromECDSAPrivateKey(priv)
	if err != nil {
		t.Errorf("FromECDSAPrivateKey err:%s", err)
	}
}

func TestFromECDSAPub(t *testing.T) {
	priv := NewECDSAPrivateKey()
	_, err := FromECDSAPublicKey(&priv.PublicKey)
	if err != nil {
		t.Errorf("FromECDSAPublicKey err:%s", err)
	}
}

func TestToECDSAPrivate(t *testing.T) {
	priv := NewECDSAPrivateKey()
	privByte, _ := FromECDSAPrivateKey(priv)
	aPriv, err := ToECDSAPrivateKey(privByte)
	if err != nil {
		t.Errorf("ToECDSAPrivateKey err:%s", err)
	}
	if !reflect.DeepEqual(priv, aPriv) {
		t.Errorf("ToECDSAPrivateKey err")
	}
}

func TestToECDSAPublic(t *testing.T) {
	priv := NewECDSAPrivateKey()
	pubByte, _ := FromECDSAPublicKey(&priv.PublicKey)
	_, err := ToECDSAPublicKey(pubByte)
	if err != nil {
		t.Errorf("FromECDSAPublicKey err:%s", err)
	}
}

func TestSign(t *testing.T) {
	type args struct {
		s []byte
	}
	type test struct {
		name    string
		priv    *ecdsa.PrivateKey
		args    args
		wantErr bool
		count   int
	}

	tests := []test{}
	for index := 0; index < 10; index++ {
		mainBuff := make([]byte, 32)
		io.ReadFull(rand.Reader, mainBuff)
		priv := NewECDSAPrivateKey()
		test := test{string(index), priv, args{hash.Sha3256(mainBuff)}, false, 1}
		tests = append(tests, test)
	}
	for _, tt := range tests {
		for index := 0; index < tt.count; index++ {
			t.Run(tt.name, func(t *testing.T) {
				got, err := Sign(tt.args.s, tt.priv) //NewAddress(tt.args.s)
				if (err != nil) != tt.wantErr {
					t.Errorf("Sign() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				gpub, err := RecoverECDSAPublicKey(tt.args.s, got)
				if err != nil {
					t.Errorf("recover failed:%s", err)
					return
				}
				originPub, _ := FromECDSAPublicKey(&tt.priv.PublicKey)
				gotPub, _ := FromECDSAPublicKey(gpub)
				if !byteutils.Equal(originPub, gotPub) {
					t.Errorf("recover failed: pub not equal")
					seckey, _ := FromECDSAPrivateKey(tt.priv)
					t.Log("private:", byteutils.Hex(seckey))
					t.Log("public:", byteutils.Hex(originPub))
					return
				}
			})
		}
	}
}
