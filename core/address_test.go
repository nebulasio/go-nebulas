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
			args{"nXgV6DnAkx4JeKnuWzNYMgaxvqZbmhNL38n2"},
			&Address{
				address: []byte{1, 0, 71, 121, 81, 68, 192, 123, 227, 7, 81, 138, 233, 187, 218, 247, 117, 213, 225, 64, 121, 53, 183, 214, 82, 76},
			},
			false,
		},
		{
			"sample contranct address",
			args{"n38NmMaShXKZ64SCskdbjQAGD22ZqzZMMen2"},
			&Address{
				address: []byte{1, 1, 147, 245, 147, 89, 227, 222, 141, 219, 123, 78, 142, 159, 229, 26, 252, 242, 124, 89, 164, 193, 56, 214, 45, 140},
			},
			false,
		},
		{
			"insufficient length",
			args{"n38NmMaShXKZ64SCskdbjQAGD22ZqzZMMen"},
			nil,
			true,
		},
		{
			"over length",
			args{"n38NmMaShXKZ64SCskdbjQAGD22ZqzZMMen22"},
			nil,
			true,
		},
		{
			"not start with n",
			args{"38NmMaShXKZ64SCskdbjQAGD22ZqzZMMen22"},
			nil,
			true,
		},
		{
			"invalid type",
			// [1 2 147 245 147 89 227 222 141 219 123 78 142 159 229 26 252 242 124 89 164 193 55 142 245 117]
			args{"nNhHTZHPx2xLNBvWCBfgsqy8gCjq87Zxg3o2"},
			nil,
			true,
		},
		{
			"invalid checksum",
			// []byte{1, 0, 71, 121, 81, 68, 192, 123, 227, 7, 81, 138, 233, 187, 218, 247, 117, 213, 225, 64, 121, 53, 183, 214, 82, 75}
			// just last byte is different from "sample account address"
			args{"nWgV6DnAkx4JeKnuWzNYMgaxvqZbmhNL38n2"},
			nil,
			true,
		},
		{
			"beyond base58 alphabet",
			// letter O
			args{"n38NmMaShXKZ64SCskdbjQAGD22ZqzZMOen2"},
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
		s [][]byte
	}
	tests := []struct {
		name    string
		args    args
		want    *Address
		wantErr bool
	}{
		{
			"sample address",
			args{[][]byte{
				{223, 77, 34, 97, 20, 18, 19, 45, 62, 155, 211, 34, 248, 46, 41, 64, 103, 78, 193, 188},
			}},
			&Address{
				// nXgV6DnAkx4JeKnuWzNYMgaxvqZbmhNL38n2
				address: []byte{1, 0, 71, 121, 81, 68, 192, 123, 227, 7, 81, 138, 233, 187, 218, 247, 117, 213, 225, 64, 121, 53, 183, 214, 82, 76},
			},
			false,
		},
		{
			"empty args",
			args{},
			nil,
			true,
		},
		{
			"empty partial of args",
			args{[][]byte{
				[]byte{},
			}},
			nil,
			true,
		},
		{
			"genesis address",
			args{[][]byte{make([]byte, AddressDataLength)}},
			&Address{
				// nNz6tuj2eEKyGEgk9SCHAXxpQavZbw3hk8n2
				address: []byte{1, 0, 79, 75, 161, 188, 115, 200, 95, 60, 154, 183, 83, 215, 19, 238, 252, 236, 110, 206, 46, 65, 135, 29, 203, 235},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAddress(AccountAddress, tt.args.s...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
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
			"sample single arg",
			args{[][]byte{
				{12, 23, 24, 109, 223, 77, 34, 97, 20, 18, 19, 45, 62, 155, 211, 34, 248, 46, 41, 64, 103, 78, 193, 188},
			}},
			&Address{
				// n38NmMaShXKZ64SCskdbjQAGD22ZqzZMMen2
				address: []byte{1, 1, 147, 245, 147, 89, 227, 222, 141, 219, 123, 78, 142, 159, 229, 26, 252, 242, 124, 89, 164, 193, 56, 214, 45, 140},
			},
			false,
		},
		{
			"sample multi args",
			args{[][]byte{
				{12, 23, 24, 109, 223, 77, 34, 97, 20, 18, 19, 45, 62, 155, 211, 34, 248, 46, 41, 64, 103, 78, 193, 188},
				{41, 64, 103, 78, 193, 188},
			}},
			&Address{
				// nqiH9Ye23MSNdqZnsWrthH42imQ9MGfJAnn2
				address: []byte{1, 1, 233, 159, 147, 235, 220, 36, 136, 197, 149, 72, 61, 48, 57, 228, 47, 72, 224, 45, 204, 179, 27, 254, 175, 74},
			},
			false,
		},
		{
			"empty args",
			args{},
			nil,
			true,
		},
		{
			"empty partial of args",
			args{[][]byte{
				[]byte{},
			}},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewContractAddressFromData(tt.args.s...)
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

/*
func TestNewGenesisCoinbaseAddress(t *testing.T) {
	addr, err := NewAddress(GenesisAddress, make([]byte, AddressDataLength))
	fmt.Println([]byte(addr.address), addr, err)
} */

/* func TestNewAddress2(t *testing.T) {

	fmt.Println("====", AccountAddress)
	addr, err := NewAddress(AccountAddress, make([]byte, AddressDataLength))
	fmt.Printf("type<AccountAddress> address = %v\n%v\n%d\n\n", addr, []byte(addr.address), len(addr.String()))
	assert.Nil(t, err)

	addr, err = NewAddress(AccountAddress, []byte{0})
	fmt.Printf("type<AccountAddress> address = %v\n%v\n%d\n\n", addr, addr.address, len(addr.String()))
	assert.Nil(t, err)
	addr, err = NewAddress(AccountAddress, []byte{223, 77, 34, 97, 20, 18, 19, 45, 62, 155, 211, 34, 248, 46, 41, 64, 103, 78, 193, 188})
	fmt.Printf("type<AccountAddress> address = %v\n%v\n%d\n\n", addr, []byte(addr.address), len(addr.String()))
	assert.Nil(t, err)

	addr, err = NewAddress(ContractAddress, []byte{12, 23, 24, 109, 223, 77, 34, 97, 20, 18, 19, 45, 62, 155, 211, 34, 248, 46, 41, 64, 103, 78, 193, 188})
	fmt.Printf("type<ContractAddress> address = %v\n%v\n%d\n\n", addr, []byte(addr.address), len(addr.String()))
	assert.Nil(t, err)

	// addr, err = NewAddress(InvalidContractAddress, []byte{12, 23, 24, 109, 223, 77, 34, 97, 20, 18, 19, 45, 62, 155, 211, 34, 248, 46, 41, 64, 103, 78, 193, 188})
	// fmt.Printf("type<InvalidContractAddress> address = %v\n%v\n%d\n\n", addr, []byte(addr.address), len(addr.String()))
	// assert.Nil(t, err)

	addr, err = NewAddress(ContractAddress, []byte{12, 23, 24, 109, 223, 77, 34, 97, 20, 18, 19, 45, 62, 155, 211, 34, 248, 46, 41, 64, 103, 78, 193, 188}, []byte{41, 64, 103, 78, 193, 188})
	fmt.Printf("type<ContractAddress> address = %v\n%v\n%d\n\n", addr, []byte(addr.address), len(addr.String()))
	assert.Nil(t, err)

	data := [][]byte{

		// len = 26
		{1, 0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		{1, 0x01, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		{1, 0x17, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		{1, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	}
	for _, bytes := range data {
		ret := btcbase58.Encode(bytes)
		fmt.Println(ret, len(ret))
	}

	a := &Address{}
	var b *Address
	fmt.Println("a", a)
	fmt.Println("b", b)
	fmt.Println(a.Equals(b))

	bs := []byte{1, 0, 71, 121, 81, 68, 192, 123, 227, 7, 81, 138, 233, 187, 218, 247, 117, 213, 225, 64, 121, 53, 183, 214, 82, 75}
	n := new(big.Int)
	n.SetBytes(bs)
	n.Mul(n, hash.Base58BigRadix)
	n.Add(n, hash.Base58Nebulas)

	fmt.Println("invalid checksum", hash.EncodeBase58(n.Bytes()))
} */
