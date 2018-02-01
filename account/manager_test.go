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

package account

import (
	"testing"

	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/stretchr/testify/assert"
)

func TestManager_NewAccount(t *testing.T) {
	manager := NewManager(nil)
	tests := []struct {
		name       string
		passphrase []byte
	}{
		{
			"address1",
			[]byte("passphrase"),
		},
		{
			"address2",
			[]byte("passphrase"),
		},
		{
			"address3",
			[]byte("passphrase"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := manager.NewAccount(tt.passphrase)
			assert.Nil(t, err, "new address err")
			addrs := manager.Accounts()
			assert.Contains(t, addrs, got, "new account not in keystore")
			err = manager.Delete(got, tt.passphrase)
			assert.Nil(t, err)
		})
	}
}

func TestManager_Unlock(t *testing.T) {
	manager := NewManager(nil)
	tests := []struct {
		name       string
		passphrase []byte
	}{
		{
			"address1",
			[]byte("passphrase"),
		},
		{
			"address2",
			[]byte("passphrase"),
		},
		{
			"address3",
			[]byte("passphrase"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := manager.NewAccount(tt.passphrase)
			assert.Nil(t, err, "new address err")
			err = manager.Unlock(got, tt.passphrase, keystore.DefaultUnlockDuration)
			assert.Nil(t, err, "unlock err")
			err = manager.Lock(got)
			assert.Nil(t, err, "lock err")
			err = manager.Delete(got, tt.passphrase)
			assert.Nil(t, err)
		})
	}
}

func TestManager_Load(t *testing.T) {
	manager := NewManager(nil)
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
	_, err := manager.Load([]byte(key), passphrase)
	assert.Nil(t, err, "import address err")
}

func TestManager_Export(t *testing.T) {
	manager := NewManager(nil)
	tests := []struct {
		name       string
		passphrase []byte
	}{
		{
			"address1",
			[]byte("passphrase"),
		},
		{
			"address2",
			[]byte("passphrase"),
		},
		{
			"address3",
			[]byte("passphrase"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := manager.NewAccount(tt.passphrase)
			assert.Nil(t, err, "new address err")
			_, err = manager.Export(got, tt.passphrase)
			assert.Nil(t, err, "export err")
			err = manager.Delete(got, tt.passphrase)
			assert.Nil(t, err)
		})
	}
}

func TestManager_SignTransaction(t *testing.T) {
	manager := NewManager(nil)
	tests := []struct {
		name       string
		passphrase []byte
	}{
		{
			"address1",
			[]byte("passphrase"),
		},
		{
			"address2",
			[]byte("passphrase"),
		},
		{
			"address3",
			[]byte("passphrase"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := manager.NewAccount(tt.passphrase)
			assert.Nil(t, err, "new address err")
			err = manager.Unlock(got, tt.passphrase, keystore.DefaultUnlockDuration)
			assert.Nil(t, err, "unlock err")
			tx := core.NewTransaction(0, got, got, util.NewUint128FromInt(5), 0, core.TxPayloadBinaryType, nil, util.NewUint128FromInt(1), util.NewUint128FromInt(5))
			err = manager.SignTransaction(got, tx)
			assert.Nil(t, err, "sign err")
			err = manager.Delete(got, tt.passphrase)
			assert.Nil(t, err)
		})
	}
}

func TestManager_SignTransactionWithPassphrase(t *testing.T) {
	manager := NewManager(nil)
	tests := []struct {
		name       string
		passphrase []byte
		unlock     bool
		want       bool
	}{
		{
			"address1",
			[]byte("passphrase"),
			true,
			true,
		},
		{
			"address2",
			[]byte("passphrase"),
			false,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := manager.NewAccount(tt.passphrase)
			assert.Nil(t, err, "new address err")
			tx := core.NewTransaction(0, got, got, util.NewUint128FromInt(5), 0, core.TxPayloadBinaryType, nil, util.NewUint128FromInt(1), util.NewUint128FromInt(5))
			err = manager.SignTransactionWithPassphrase(got, tx, tt.passphrase)
			assert.Nil(t, err, "sign with passphrase err")
			err = manager.Delete(got, tt.passphrase)
			assert.Nil(t, err)
		})
	}
}
