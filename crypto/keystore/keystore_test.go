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

package keystore

import (
	"crypto/rand"
	"github.com/nebulasio/go-nebulas/crypto/keystore/ecdsa"
	"github.com/nebulasio/go-nebulas/crypto/keystore/key"
	"github.com/nebulasio/go-nebulas/utils/byteutils"
	"testing"
)

func TestKeystore_SetKeyPassphrase(t *testing.T) {
	priv, _ := ecdsa.GenerateECDSAPrivateKey(rand.Reader)

	ks := NewKeystore()

	alias := key.Alias("test_alias")

	kpri := ecdsa.NewPrivateStoreKey(priv)
	err := ks.SetKeyPassphrase(alias, kpri, []byte("paeeword"))
	if err != nil {
		t.Errorf("SetKeyPassphrase failed:%s", err)
	}
}

func TestKeystore_SetKey(t *testing.T) {
	ks := NewKeystore()

	alias := key.Alias("test_alias")

	err := ks.SetKey(alias, []byte{})
	if err != nil {
		t.Errorf("SetKey failed:%s", err)
	}
}

func TestKeystore_GetKey(t *testing.T) {
	priv, _ := ecdsa.GenerateECDSAPrivateKey(rand.Reader)

	ks := NewKeystore()

	alias := key.Alias("test_alias")

	kpri := ecdsa.NewPrivateStoreKey(priv)
	kdata, _ := kpri.Encoded()
	err := ks.SetKeyPassphrase(alias, kpri, []byte("password"))
	if err != nil {
		t.Errorf("SetKeyPassphrase failed:%s", err)
	}
	pass, pErr := key.NewPassphrase([]byte("password"))
	if pErr != nil {
		t.Errorf("NewPassphrase failed:%s", pErr)
	}
	key, err := ks.GetKey(alias, pass)
	data, _ := key.Encoded()
	if !byteutils.Equal(kdata, data) {
		t.Errorf("GetKey err")
	}
}

func TestKeystore_ContainsAlias(t *testing.T) {
	priv, _ := ecdsa.GenerateECDSAPrivateKey(rand.Reader)

	ks := NewKeystore()

	alias := key.Alias("test_alias")

	kpri := ecdsa.NewPrivateStoreKey(priv)
	err := ks.SetKeyPassphrase(alias, kpri, []byte("password"))
	if err != nil {
		t.Errorf("SetKeyPassphrase failed:%s", err)
	}
	_, err = ks.ContainsAlias(alias)
	if err != nil {
		t.Errorf("ContainsAlias failed:%s", err)
	}
}

func TestKeystore_Delete(t *testing.T) {
	priv, _ := ecdsa.GenerateECDSAPrivateKey(rand.Reader)

	alias := key.Alias("test_alias")

	ks := NewKeystore()

	kpri := ecdsa.NewPrivateStoreKey(priv)
	err := ks.SetKeyPassphrase(alias, kpri, []byte("password"))
	if err != nil {
		t.Errorf("SetKeyPassphrase failed:%s", err)
	}

	ks.Delete(alias)
	_, err = ks.ContainsAlias(alias)
	if err != nil {
		t.Errorf("ContainsAlias failed:%s", err)
	}
}
