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

	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/crypto/keystore/secp256k1"
	"github.com/stretchr/testify/assert"
)

func mockAddress() *Address {
	ks := keystore.DefaultKS
	priv1 := secp256k1.GeneratePrivateKey()
	pubdata1, _ := priv1.PublicKey().Encoded()
	addr, _ := NewAddressFromPublicKey(pubdata1)
	ks.SetKey(addr.String(), priv1, []byte("passphrase"))
	ks.Unlock(addr.String(), []byte("passphrase"), time.Second*60*60*24*365)
	return addr
}

func TestParse(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    *Address
		wantErr bool
	}{
		{
			"sample account address",
			args{"n1TV3sU6jyzR4rJ1D7jCAmtVGSntJagXZHC"},
			&Address{
				address: []byte{25, 87, 142, 66, 118, 31, 219, 67, 233, 197, 173, 156, 169, 58, 207, 29, 218, 107, 233, 1, 248, 72, 171, 56, 100, 171},
			},
			false,
		},
		{
			"sample contranct address",
			args{"n1sLnoc7j57YfzAVP8tJ3yK5a2i56QrTDdK"},
			&Address{
				address: []byte{25, 88, 147, 245, 147, 89, 227, 222, 141, 219, 123, 78, 142, 159, 229, 26, 252, 242, 124, 89, 164, 193, 65, 149, 143, 170},
			},
			false,
		},
		{
			"insufficient length",
			args{"n38NmMaShXKZ64SCskdbjQAGD22ZqzZMMe"},
			nil,
			true,
		},
		{
			"over length",
			args{"n38NmMaShXKZ64SCskdbjQAGD22ZqzZMMen2"},
			nil,
			true,
		},
		{
			"not start with n",
			args{"38NmMaShXKZ64SCskdbjQAGD22ZqzZMMen2"},
			nil,
			true,
		},
		{
			"invalid type",
			// [25, 23, 233, 159, 147, 235, 220, 36, 136, 197, 149, 72, 61, 48, 57, 228, 47, 72, 224, 45, 204, 179, 194, 171, 165, 35]
			args{"mZrBXtpFbK8vUisQ3mNevawvFEjfNLr3rHY"},
			nil,
			true,
		},
		{
			"invalid checksum",
			// []byte{25, 87, 71, 121, 81, 68, 192, 123, 227, 7, 81, 138, 233, 187, 218, 247, 117, 213, 225, 64, 121, 53, 109, 153, 204, 124}
			// just last byte is different from "sample account address"
			args{"n1M2mcK3mcwGNQS7Kt7wmKadJn97paakkZ8"},
			nil,
			true,
		},
		{
			"beyond base58 alphabet",
			// letter O
			args{"n38NmMaShXKZ64SCskdbjQAGD22ZqzZMOen"},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AddressParse(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddressParse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddressParse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewAddress(t *testing.T) {
	type args struct {
		s []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *Address
		wantErr bool
	}{
		{
			"sample address",
			args{[]byte{223, 77, 34, 97, 20, 18, 19, 45, 62, 155, 211, 34, 248, 46, 41, 64, 103, 78, 193, 188}},
			nil,
			true,
		},
		{
			"empty args",
			args{},
			nil,
			true,
		},
		{
			"empty partial of args",
			args{[]byte{}},
			nil,
			true,
		},
		{
			"genesis address",
			args{make([]byte, PublicKeyDataLength)},
			&Address{
				// n1Mk8HYsjyfkEnGv4hZacDVFz2R1gYEXNej
				address: []byte{25, 87, 31, 185, 52, 219, 123, 90, 141, 18, 150, 2, 238, 90, 246, 63, 157, 200, 210, 75, 140, 233, 151, 17, 227, 22},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAddressFromPublicKey(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// fmt.Println(got.Bytes())
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewContractAddress(t *testing.T) {
	type args struct {
		s [][]byte
	}
	tests := []struct {
		name    string
		args    args
		want    *Address
		wantErr bool
	}{
		{
			"sample multi args",
			args{[][]byte{
				{12, 23, 24, 109, 223, 77, 34, 97, 20, 18, 19, 45, 62, 155, 211, 34, 248, 46, 41, 64, 103, 78, 193, 188},
				{41, 64, 103, 78, 193, 188},
			}},
			&Address{
				// n219jtsdKTs3Rt3TgMeJe76cP9XQgYLdJXu
				address: []byte{25, 88, 233, 159, 147, 235, 220, 36, 136, 197, 149, 72, 61, 48, 57, 228, 47, 72, 224, 45, 204, 179, 166, 28, 157, 212},
			},
			false,
		},
		{
			"empty args",
			args{[][]byte{
				[]byte{},
				[]byte{},
			}},
			nil,
			true,
		},
		{
			"empty partial of args",
			args{[][]byte{
				[]byte{},
				[]byte{01},
			}},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewContractAddressFromData(tt.args.s[0], tt.args.s[1])
			if (err != nil) != tt.wantErr {
				t.Errorf("NewContractAddressFromData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddressGetType(t *testing.T) {
	tests := []struct {
		address string
		want    AddressType
	}{
		{"n1TV3sU6jyzR4rJ1D7jCAmtVGSntJagXZHC", AccountAddress},
		{"n1sLnoc7j57YfzAVP8tJ3yK5a2i56QrTDdK", ContractAddress},
	}

	for _, tt := range tests {
		t.Run(tt.address, func(t *testing.T) {
			addr, err := AddressParse(tt.address)
			assert.Nil(t, err)
			assert.NotNil(t, addr)
			assert.Equal(t, tt.want, addr.Type())
		})
	}
}
