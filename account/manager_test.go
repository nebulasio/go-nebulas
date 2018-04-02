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

	"os"

	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/stretchr/testify/assert"
)

func TestManager_NewAccount(t *testing.T) {
	manager, _ := NewManager(nil)
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
			acc, err := manager.getAccount(got)
			assert.Nil(t, err, "new acc err")
			err = manager.Remove(got, tt.passphrase)
			assert.Nil(t, err)
			err = os.Remove(acc.path)
			assert.Nil(t, err)
		})
	}
}

func TestManager_Unlock(t *testing.T) {
	manager, _ := NewManager(nil)
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
			acc, err := manager.getAccount(got)
			assert.Nil(t, err, "new acc err")
			err = manager.Remove(got, tt.passphrase)
			assert.Nil(t, err)
			err = os.Remove(acc.path)
			assert.Nil(t, err)
		})
	}
}

func TestManager_Load(t *testing.T) {
	manager, _ := NewManager(nil)
	passphrase := []byte("b84c54af84672b5ae814")
	key := `{"address":"n1NaY2ywi1J6ENA1htPa4FdeTRMo2hjpD8f","crypto":{"cipher":"aes-128-ctr","ciphertext":"2fc58f9135b318a98e8dd00d7da19aed2108c6fde033f4f7c28e710c06cdc740","cipherparams":{"iv":"4359b97562f55aa774af5dbb83b0b378"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":4096,"p":1,"r":8,"salt":"ef2940317217226d7d9a11c4464b89cdc97edd2520f556858c8b566612a38ec3"},"mac":"dbbaf40ae7f7c5428b9e907e539c407f670fc11f56423b71fabe6bdc05c8d191","machash":"sha3256"},"id":"1eb90282-5e82-41be-8dd0-f3a6ff86eda0","version":3}`
	_, err := manager.Load([]byte(key), passphrase)
	assert.Nil(t, err, "import address err")
}

func TestManager_Export(t *testing.T) {
	manager, _ := NewManager(nil)
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
			acc, err := manager.getAccount(got)
			assert.Nil(t, err, "new acc err")
			err = manager.Remove(got, tt.passphrase)
			assert.Nil(t, err)
			err = os.Remove(acc.path)
			assert.Nil(t, err)
		})
	}
}

func TestManager_SignTransaction(t *testing.T) {
	manager, _ := NewManager(nil)
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
			value, _ := util.NewUint128FromInt(5)
			gasLimit, _ := util.NewUint128FromInt(5)
			gasPrice, _ := util.NewUint128FromInt(1)
			tx, _ := core.NewTransaction(0, got, got, value, 0, core.TxPayloadBinaryType, nil, gasPrice, gasLimit)
			err = manager.SignTransaction(got, tx)
			assert.Nil(t, err, "sign err")
			acc, err := manager.getAccount(got)
			assert.Nil(t, err, "new acc err")
			err = manager.Remove(got, tt.passphrase)
			assert.Nil(t, err)
			err = os.Remove(acc.path)
			assert.Nil(t, err)
		})
	}
}

func TestManager_SignTransactionWithPassphrase(t *testing.T) {
	manager, _ := NewManager(nil)
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
			value, _ := util.NewUint128FromInt(5)
			gasLimit, _ := util.NewUint128FromInt(5)
			gasPrice, _ := util.NewUint128FromInt(1)
			tx, _ := core.NewTransaction(0, got, got, value, 0, core.TxPayloadBinaryType, nil, gasPrice, gasLimit)
			err = manager.SignTransactionWithPassphrase(got, tx, tt.passphrase)
			assert.Nil(t, err, "sign with passphrase err")
			acc, err := manager.getAccount(got)
			assert.Nil(t, err, "new acc err")
			err = manager.Remove(got, tt.passphrase)
			assert.Nil(t, err)
			err = os.Remove(acc.path)
			assert.Nil(t, err)
		})
	}
}
