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

package test

import (
	"testing"

	"time"

	"github.com/nebulasio/go-nebulas/crypto"
	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/stretchr/testify/assert"
)

func TestKeystore_SetKeyPassphrase(t *testing.T) {
	priv1, _ := crypto.NewPrivateKey(keystore.SECP256K1, nil)
	priv2, _ := crypto.NewPrivateKey(keystore.SECP256K1, nil)
	priv3, _ := crypto.NewPrivateKey(keystore.SECP256K1, nil)

	ks := keystore.NewKeystore()

	tests := []struct {
		name       string
		passphrase []byte
		alias      string
		key        keystore.PrivateKey
	}{
		{
			"address1",
			[]byte("passphrase"),
			"alias1",
			priv1,
		},
		{
			"address2",
			[]byte("passphrase"),
			"alias2",
			priv2,
		},
		{
			"address3",
			[]byte("passphrase"),
			"alias3",
			priv3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ks.SetKey(tt.alias, tt.key, tt.passphrase)
			assert.Nil(t, err, "set key err")
			got, err := ks.GetKey(tt.alias, tt.passphrase)
			assert.Equal(t, tt.key, got, "keystore get err")
		})
	}
}

func TestKeystore_ContainsAlias(t *testing.T) {
	priv1, _ := crypto.NewPrivateKey(keystore.SECP256K1, nil)
	priv2, _ := crypto.NewPrivateKey(keystore.SECP256K1, nil)
	priv3, _ := crypto.NewPrivateKey(keystore.SECP256K1, nil)

	ks := keystore.NewKeystore()

	tests := []struct {
		name       string
		passphrase []byte
		alias      string
		key        keystore.PrivateKey
	}{
		{
			"address1",
			[]byte("passphrase"),
			"alias1",
			priv1,
		},
		{
			"address2",
			[]byte("passphrase"),
			"alias2",
			priv2,
		},
		{
			"address3",
			[]byte("passphrase"),
			"alias3",
			priv3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ks.SetKey(tt.alias, tt.key, tt.passphrase)
			assert.Nil(t, err, "set key err")
			got, err := ks.ContainsAlias(tt.alias)
			assert.Nil(t, err, "contains err")
			assert.True(t, got, "not contains")
		})
	}
}

func TestKeystore_Unlock(t *testing.T) {
	priv1, _ := crypto.NewPrivateKey(keystore.SECP256K1, nil)
	priv2, _ := crypto.NewPrivateKey(keystore.SECP256K1, nil)
	priv3, _ := crypto.NewPrivateKey(keystore.SECP256K1, nil)

	ks := keystore.NewKeystore()

	tests := []struct {
		name       string
		passphrase []byte
		alias      string
		key        keystore.PrivateKey
		duration   time.Duration
		want       bool
	}{
		{
			"address1",
			[]byte("passphrase"),
			"alias1",
			priv1,
			time.Second * 1,
			false,
		},
		{
			"address2",
			[]byte("passphrase"),
			"alias2",
			priv2,
			time.Second * 3,
			true,
		},
		{
			"address3",
			[]byte("passphrase"),
			"alias3",
			priv3,
			time.Second * 5,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ks.SetKey(tt.alias, tt.key, tt.passphrase)
			assert.Nil(t, err, "set key err")
			err = ks.Unlock(tt.alias, tt.passphrase, tt.duration)
			assert.Nil(t, err, "unlock err")
		})
	}
	time.Sleep(time.Second * 2)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := ks.GetUnlocked(tt.alias)
			assert.Equal(t, tt.want, tt.key == got, "get unlock err:%s", tt.alias)
		})
	}
}

func TestKeystore_Delete(t *testing.T) {
	priv1, _ := crypto.NewPrivateKey(keystore.SECP256K1, nil)
	priv2, _ := crypto.NewPrivateKey(keystore.SECP256K1, nil)
	priv3, _ := crypto.NewPrivateKey(keystore.SECP256K1, nil)

	ks := keystore.NewKeystore()

	tests := []struct {
		name       string
		passphrase []byte
		alias      string
		key        keystore.PrivateKey
	}{
		{
			"address1",
			[]byte("passphrase"),
			"alias1",
			priv1,
		},
		{
			"address2",
			[]byte("passphrase"),
			"alias2",
			priv2,
		},
		{
			"address3",
			[]byte("passphrase"),
			"alias3",
			priv3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ks.SetKey(tt.alias, tt.key, tt.passphrase)
			assert.Nil(t, err, "set key err")
			err = ks.Delete(tt.alias, tt.passphrase)
			assert.Nil(t, err, "delete err")
			got, err := ks.ContainsAlias(tt.alias)
			assert.NotNil(t, err, "contains err")
			assert.False(t, got, "not contains")
		})
	}
}
