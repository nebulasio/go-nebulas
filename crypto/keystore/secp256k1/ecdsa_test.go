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

	"github.com/nebulasio/go-nebulas/util/byteutils"
)

func TestFromECDSAPri(t *testing.T) {
	priv, _ := NewECDSAPrivateKey()
	_, err := FromECDSAPrivateKey(priv)
	if err != nil {
		t.Errorf("FromECDSAPrivateKey err:%s", err)
	}
}

func TestFromECDSAPub(t *testing.T) {
	priv, _ := NewECDSAPrivateKey()
	_, err := FromECDSAPublicKey(&priv.PublicKey)
	if err != nil {
		t.Errorf("FromECDSAPublicKey err:%s", err)
	}
}

func TestToECDSAPrivate(t *testing.T) {
	priv, _ := NewECDSAPrivateKey()
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
	priv, _ := NewECDSAPrivateKey()
	pubByte, _ := FromECDSAPublicKey(&priv.PublicKey)
	_, err := ToECDSAPublicKey(pubByte)
	if err != nil {
		t.Errorf("FromECDSAPublicKey err:%s", err)
	}
}

func TestSign(t *testing.T) {
	priv, _ := NewECDSAPrivateKey()
	hash1, _ := byteutils.FromHex("0eb3be2db3a534c192be5570c6c42f59")
	hash2, _ := byteutils.FromHex("5e6d587f26121f96a07cf4b8b569aac1AAAAAAAA") //5e6d587f26121f96a07cf4b8b569aac1
	hash3, _ := byteutils.FromHex("c7174759e86c59dcb7df87def82f61eb")         //c7174759e86c59dcb7df87def82f61eb
	type args struct {
		s []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"sample hash1",
			args{hash1},
			false,
		},
		{
			"sample hash2",
			args{hash2},
			false,
		},
		{
			"sample hash3",
			args{hash3},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Sign(tt.args.s, priv) //NewAddress(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("Sign() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("Sign() = %v, want %v", got, tt.want)
			//}
			if !Verify(tt.args.s, got, &priv.PublicKey) {
				t.Errorf("Verify() false hash = %v", tt.args.s)
				return
			}
			gpub, err := RecoverECDSAPublicKey(tt.args.s, got)
			if err != nil {
				t.Errorf("recover failed:%s", err)
				return
			}
			if gpub == nil {
				t.Errorf("recover failed: pub nil")
				return
			}
			//t.Logf("orgpub pubX:%d, pubY:%d", priv.PublicKey.X, priv.PublicKey.Y)
			//t.Logf("newpub pubX:%d, pubY:%d", gpub.X, gpub.Y)
			if !Verify(tt.args.s, got, gpub) {
				t.Errorf("recover Verify() false hash = %v", tt.args.s)
				return
			}
		})
	}
}
