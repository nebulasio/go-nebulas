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
			args{"n1M2mcK3mcwGNQS7Kt7wmKadJn97paakkZ9"},
			&Address{
				address: []byte{25, 87, 71, 121, 81, 68, 192, 123, 227, 7, 81, 138, 233, 187, 218, 247, 117, 213, 225, 64, 121, 53, 109, 153, 204, 124},
			},
			false,
		},
		{
			"sample contranct address",
			args{"n1sLnoc7j57YfzAVP8tJ3yK5a2i56Uka4v6"},
			&Address{
				address: []byte{25, 88, 147, 245, 147, 89, 227, 222, 141, 219, 123, 78, 142, 159, 229, 26, 252, 242, 124, 89, 164, 193, 218, 26, 188, 107},
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
				// n1M2mcK3mcwGNQS7Kt7wmKadJn97paakkZ9
				address: []byte{25, 87, 71, 121, 81, 68, 192, 123, 227, 7, 81, 138, 233, 187, 218, 247, 117, 213, 225, 64, 121, 53, 109, 153, 204, 124},
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
				// n1Mk8HYsjyfkEnGv4hZacDVFz2R1gYEXNej
				address: []byte{25, 87, 79, 75, 161, 188, 115, 200, 95, 60, 154, 183, 83, 215, 19, 238, 252, 236, 110, 206, 46, 65, 58, 216, 79, 32},
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
				// n1sLnoc7j57YfzAVP8tJ3yK5a2i56Uka4v6
				address: []byte{25, 88, 147, 245, 147, 89, 227, 222, 141, 219, 123, 78, 142, 159, 229, 26, 252, 242, 124, 89, 164, 193, 218, 26, 188, 107},
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
				// n219jtsdKTs3Rt3TgMeJe76cP9XQgXvQZ1n
				address: []byte{25, 88, 233, 159, 147, 235, 220, 36, 136, 197, 149, 72, 61, 48, 57, 228, 47, 72, 224, 45, 204, 179, 149, 198, 110, 149},
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

/* func TestBigINt(t *testing.T) {
	// low = 58^33 * 45
	// high = 58^33 * 46 - 1
	low := new(big.Int)
	high := new(big.Int)
	low.Exp(big.NewInt(58), big.NewInt(33), nil)
	fmt.Println(low.Bytes())
	high.Mul(low, big.NewInt(46))
	high.Sub(high, big.NewInt(1))
	low.Mul(low, big.NewInt(45))
	fmt.Println(low.Bytes())  // [111 213 184 75 176 88 186 42 63 104 237 127 10 128 7 147 66 24 23 39 50 0 0 0 0]
	fmt.Println(high.Bytes()) // [114 81 239 151 83 141 230 31 207 10 140 95 187 22 201 42 113 18 239 216 107 255 255 255 255]
} */

/* func TestNewGenesisCoinbaseAddress(t *testing.T) {

	data := [][]byte{
		// len =26
		{0x19, 0x70, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		{0x19, 0x70, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	}
	for i := 0; i <= 255; i++ {
		data[0][1] = byte(i)
		data[1][1] = byte(i)

		// ret0 := hash.EncodeBase58(data[0])
		// ret1 := hash.EncodeBase58(data[0])
		// if ret0[len(ret0)-1] == 'n' && ret1[len(ret1)-1] == 'n' {
		// 112 113
		// fmt.Println(i, ret0, len(ret0), ret1, len(ret1))
		// }
		// fmt.Println(i, data[0], ret1, len(ret1))
	}

} */

/* func TestNewAddress2(t *testing.T) {

	fmt.Println("AccountAddress = ", AccountAddress, ", ContractAddress = ", ContractAddress)

	addr, err := NewAddress(AccountAddress, make([]byte, AddressDataLength))
	fmt.Printf("type<AccountAddress> address = %v\n%v\n%d\n\n", addr, []byte(addr.address), len(addr.String()))
	assert.Nil(t, err)

	addr, err = NewAddress(ContractAddress, [][]byte{
		{12, 23, 24, 109, 223, 77, 34, 97, 20, 18, 19, 45, 62, 155, 211, 34, 248, 46, 41, 64, 103, 78, 193, 188},
		{41, 64, 103, 78, 193, 188}}...)
	fmt.Printf("type<ContractAddress> address = %v\n%v\n%d\n\n", addr, []byte(addr.address), len(addr.String()))
	assert.Nil(t, err)

	addr, err = NewAddress(ContractAddress, []byte{12, 23, 24, 109, 223, 77, 34, 97, 20, 18, 19, 45, 62, 155, 211, 34, 248, 46, 41, 64, 103, 78, 193, 188})
	fmt.Printf("type<ContractAddress> address = %v\n%v\n%d\n\n", addr, []byte(addr.address), len(addr.String()))
	assert.Nil(t, err)

	s := base58.Encode([]byte{25, 87, 71, 121, 81, 68, 192, 123, 227, 7, 81, 138, 233, 187, 218, 247, 117, 213, 225, 64, 121, 53, 109, 153, 204, 123})
	fmt.Println(s)
} */
