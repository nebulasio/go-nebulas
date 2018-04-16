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

	"crypto/ecdsa"
	"crypto/rand"
	"io"

	"github.com/nebulasio/go-nebulas/crypto/hash"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/stretchr/testify/assert"
)

func TestFromECDSAPri(t *testing.T) {
	priv, err := NewECDSAPrivateKey()
	assert.Nil(t, err)
	_, err = FromECDSAPrivateKey(priv)
	assert.Nil(t, err)
}

func TestFromECDSAPub(t *testing.T) {
	priv, err := NewECDSAPrivateKey()
	assert.Nil(t, err)
	_, err = FromECDSAPublicKey(&priv.PublicKey)
	assert.Nil(t, err)
}

func TestToECDSAPrivate(t *testing.T) {
	priv, err := NewECDSAPrivateKey()
	assert.Nil(t, err)
	privByte, err := FromECDSAPrivateKey(priv)
	assert.Nil(t, err)
	aPriv, err := ToECDSAPrivateKey(privByte)
	assert.Nil(t, err)
	assert.Equal(t, priv, aPriv)
}

func TestToECDSAPublic(t *testing.T) {
	priv, err := NewECDSAPrivateKey()
	assert.Nil(t, err)
	pubByte, err := FromECDSAPublicKey(&priv.PublicKey)
	assert.Nil(t, err)
	_, err = ToECDSAPublicKey(pubByte)
	assert.Nil(t, err)
}

func TestSign(t *testing.T) {
	type args struct {
		s []byte
	}
	type test struct {
		name  string
		priv  *ecdsa.PrivateKey
		args  args
		count int
	}

	tests := []test{}
	for index := 0; index < 10; index++ {
		mainBuff := make([]byte, 32)
		io.ReadFull(rand.Reader, mainBuff)
		priv, err := NewECDSAPrivateKey()
		assert.Nil(t, err)
		test := test{string(index), priv, args{hash.Sha3256(mainBuff)}, 1}
		tests = append(tests, test)
	}
	for _, tt := range tests {
		for index := 0; index < tt.count; index++ {
			t.Run(tt.name, func(t *testing.T) {
				privData, err := FromECDSAPrivateKey(tt.priv)
				assert.Nil(t, err)
				got, err := Sign(tt.args.s, privData)
				assert.Nil(t, err)
				gpub, err := RecoverECDSAPublicKey(tt.args.s, got)
				assert.Nil(t, err)
				originPub, err := FromECDSAPublicKey(&tt.priv.PublicKey)
				assert.Nil(t, err)
				if !byteutils.Equal(originPub, gpub) {
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
