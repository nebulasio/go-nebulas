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
	"errors"
	"github.com/nebulasio/go-nebulas/crypto/keystore/ecdsa"
	"github.com/nebulasio/go-nebulas/crypto/keystore/key"
)

const (
	// KeystoreTypeDefault default keystore
	KeystoreTypeDefault = "default"
)

var (
	// ErrUninitialized uninitialized provider error.
	ErrUninitialized = errors.New("keystore: uninitialized the provider")
)

// Keystore class represents a storage facility for cryptographic keys
type Keystore struct {

	// keystore provider
	p Provider
}

// NewKeystore new
func NewKeystore() *Keystore {
	return NewKeystoreType(KeystoreTypeDefault)
}

// NewKeystoreType new keystore with type
func NewKeystoreType(t string) *Keystore {
	ks := &Keystore{}
	switch t {
	case KeystoreTypeDefault:
		ks.p = ecdsa.NewProvider(1.0)
	default:
		ks.p = ecdsa.NewProvider(1.0)
	}
	return ks
}

// Aliases lists all the alias names of this keystore.
func (ks *Keystore) Aliases() ([]key.Alias, error) {
	if ks.p == nil {
		return nil, ErrUninitialized
	}
	return ks.p.Aliases(), nil
}

// ContainsAlias checks if the given alias exists in this keystore.
func (ks *Keystore) ContainsAlias(a key.Alias) (bool, error) {
	if ks.p == nil {
		return false, ErrUninitialized
	}
	return ks.p.ContainsAlias(a)
}

// SetKeyPassphrase assigns the given key to the given alias, protecting it with the given passphrase.
func (ks *Keystore) SetKeyPassphrase(a key.Alias, k key.Key, passphrase []byte) error {
	if ks.p == nil {
		return ErrUninitialized
	}
	return ks.p.SetKeyPassphrase(a, k, passphrase)
}

// SetKey assigns the given key (that has already been protected) to the given alias.
func (ks *Keystore) SetKey(a key.Alias, d []byte) error {
	if ks.p == nil {
		return ErrUninitialized
	}
	return ks.p.SetKey(a, d)
}

// GetKey returns the key associated with the given alias, using the given
// password to recover it.
func (ks *Keystore) GetKey(a key.Alias, p key.ProtectionParameter) (key.Key, error) {
	if ks.p == nil {
		return nil, ErrUninitialized
	}
	return ks.p.GetKey(a, p)
}

// Delete the entry identified by the given alias from this keystore.
func (ks *Keystore) Delete(a key.Alias) error {
	if ks.p == nil {
		return ErrUninitialized
	}
	return ks.p.Delete(a)
}

// Load this KeyStore from the given input stream.
func (ks *Keystore) Load(d []byte, passphrase []byte) error {
	if ks.p == nil {
		return ErrUninitialized
	}
	ks.p.Load(d, passphrase)
	return nil
}

// LoadFile load this KeyStore from the given file path
func (ks *Keystore) LoadFile(f string, passphrase []byte) error {
	if ks.p == nil {
		return ErrUninitialized
	}
	ks.p.LoadFile(f, passphrase)
	return nil
}

// Store this keystore to the output stream, and protects its
// integrity with the given password.
func (ks *Keystore) Store(passphrase []byte) (out []byte, err error) {
	if ks.p == nil {
		return nil, ErrUninitialized
	}
	return ks.p.Store(passphrase)
}

// StoreFile this keystore to the given file, and protects its
// integrity with the given password.
func (ks *Keystore) StoreFile(f string, passphrase []byte) error {
	if ks.p == nil {
		return ErrUninitialized
	}
	ks.p.StoreFile(f, passphrase)
	return nil
}
