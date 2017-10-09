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
	"crypto/rand"
	"testing"

	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/crypto/keystore/ecdsa"
)

func TestKeystore_SetKeyPassphrase(t *testing.T) {
	priv, _ := ecdsa.NewPrivateKey(rand.Reader)

	ks := keystore.NewKeystore()

	alias := "test_alias"

	kpri := ecdsa.NewPrivateStoreKey(priv)
	err := ks.SetKey(alias, kpri, []byte("paeeword"))
	if err != nil {
		t.Errorf("SetKeyPassphrase failed:%s", err)
	}
}

func TestKeystore_GetKey(t *testing.T) {
	priv, _ := ecdsa.NewPrivateKey(rand.Reader)

	ks := keystore.NewKeystore()

	alias := "test_alias"

	passphrase := []byte("password")

	kpri := ecdsa.NewPrivateStoreKey(priv)
	err := ks.SetKey(alias, kpri, passphrase)
	if err != nil {
		t.Errorf("SetKeyPassphrase failed:%s", err)
	}
	key, err := ks.GetKey(alias, passphrase)
	if err != nil {
		t.Errorf("getkey failed:%s", err)
	}
	if key == nil {
		t.Errorf("getkey failed")
	}

}

func TestKeystore_ContainsAlias(t *testing.T) {
	priv, _ := ecdsa.NewPrivateKey(rand.Reader)

	ks := keystore.NewKeystore()

	alias := "test_alias"
	passphrase := []byte("password")

	kpri := ecdsa.NewPrivateStoreKey(priv)
	err := ks.SetKey(alias, kpri, passphrase)
	if err != nil {
		t.Errorf("SetKeyPassphrase failed:%s", err)
	}
	_, err = ks.ContainsAlias(alias)
	if err != nil {
		t.Errorf("ContainsAlias failed:%s", err)
	}
}

func TestKeystore_Unlock(t *testing.T) {
	priv, _ := ecdsa.NewPrivateKey(rand.Reader)

	ks := keystore.NewKeystore()

	alias := "test_alias"
	passphrase := []byte("password")

	kpri := ecdsa.NewPrivateStoreKey(priv)
	err := ks.SetKey(alias, kpri, passphrase)
	if err != nil {
		t.Errorf("SetKeyPassphrase failed:%s", err)
	}
	err = ks.Unlock(alias, passphrase, keystore.DefaultUnlockDuration)
	if err != nil {
		t.Errorf("Unlock failed:%s", err)
	}
	_, err = ks.GetUnlocked(alias)
	if err != nil {
		t.Errorf("GetUnlocked failed:%s", err)
	}
}

func TestKeystore_Delete(t *testing.T) {
	priv, _ := ecdsa.NewPrivateKey(rand.Reader)

	alias := "test_alias"
	passphrase := []byte("password")

	ks := keystore.NewKeystore()

	kpri := ecdsa.NewPrivateStoreKey(priv)
	err := ks.SetKey(alias, kpri, passphrase)
	if err != nil {
		t.Errorf("SetKeyPassphrase failed:%s", err)
	}

	ks.Delete(alias)
	_, err = ks.ContainsAlias(alias)
	if err == nil {
		t.Errorf("ContainsAlias failed:%s", err)
	}
}
