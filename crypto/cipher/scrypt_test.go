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

package cipher

import (
	"reflect"
	"testing"

	"github.com/nebulasio/go-nebulas/util/byteutils"
)

func TestScrypt_Encrypt(t *testing.T) {
	passphrase := []byte("passphrase")
	hash1, _ := byteutils.FromHex("0eb3be2db3a534c192be5570c6c42f59")
	hash2, _ := byteutils.FromHex("5e6d587f26121f96a07cf4b8b569aac1")
	hash3, _ := byteutils.FromHex("c7174759e86c59dcb7df87def82f61eb")

	scrypt := new(Scrypt)
	tests := []struct {
		name string
		data []byte
	}{
		{
			"test1",
			hash1,
		},
		{
			"test2",
			hash2,
		},
		{
			"test3",
			hash3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := scrypt.Encrypt(tt.data, passphrase)
			if err != nil {
				t.Errorf("Encrypt() error = %v", err)
				return
			}
			want, err := scrypt.Decrypt(got, passphrase)
			if err != nil {
				t.Errorf("Decrypt() error = %v", err)
				return
			}
			if !reflect.DeepEqual(tt.data, want) {
				t.Errorf("Decrypt() = %v, data %v", want, tt.data)
			}
		})
	}
}

func TestScrypt_DecryptKey(t *testing.T) {
	passphrase := []byte("qwertyuiop")
	key := `{
    "version":3,
    "id":"3913ded3-2707-4a25-996a-807265dc0cdf",
    "address":"70e30fcae5e7f4b2460faaa9e5b1bd912332ebb5",
    "Crypto":{
        "ciphertext":"30c9606797a6e4fd5bb8e91694184ecdb9ab0230c453fe1922732a1e3212301c",
        "cipherparams":{
            "iv":"65d14cb11d6bb6e57dff0d12346637cc"
        },
        "cipher":"aes-128-ctr",
        "kdf":"scrypt",
        "kdfparams":{
            "dklen":32,
            "salt":"8728c5a28888692acb5e28ee46bdc7935b8306dfece2c6d0cd003b2cbc259af2",
            "n":1024,
            "r":8,
            "p":1
        },
        "mac":"a22874c9c35e365e305b1defe6663bde930d2efbcc6c3d0db192ff44bd9dfa7c"
    }
	}`
	scrypt := new(Scrypt)
	_, err := scrypt.DecryptKey([]byte(key), passphrase)
	if err != nil {
		t.Errorf("DecryptKey() error = %v", err)
		return
	}
	//t.Logf("decrypt key :%d", d)
}
