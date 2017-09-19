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
)

var (
	// ErrUninitialized uninitialized provider error.
	ErrUninitialized = errors.New("keystore: uninitialized the provider")
)

/*
 * This class represents a storage facility for cryptographic keys
 */
type Keystore struct {

	// keystore provider
	p Provider
}

// Lists all the alias names of this keystore.
func (ks *Keystore) Aliases() ([]Alias, error) {
	if ks.p == nil {
		return nil, ErrUninitialized
	}
	keys := make([]Alias, len(ks.p.Entries()))
	for k := range ks.p.Entries() {
		keys = append(keys, k)
	}
	return keys, nil
}

// Checks if the given alias exists in this keystore.
func (ks *Keystore) ContainsAlias(a Alias) (bool, error) {
	if ks.p == nil {
		return false, ErrUninitialized
	}
	return ks.p.ContainsAlias(a)
}

/*
* Assigns the given key to the given alias, protecting it with the given
* passphrase.
 */
func (ks *Keystore) SetKeyPassphrase(a Alias, k Key, passphrase []byte) error {
	if ks.p == nil {
		return ErrUninitialized
	}
	return ks.p.SetKeyPassphrase(a, k, passphrase)
}

/*
* Assigns the given key (that has already been protected) to the given
* alias.
 */
func (ks *Keystore) SetKey(a Alias, d []byte) error {
	if ks.p == nil {
		return ErrUninitialized
	}
	return ks.p.SetKey(a, d)
}

/*
* Returns the key associated with the given alias, using the given
* password to recover it.
 */
func (ks *Keystore) GetKey(a Alias, p ProtectionParameter) (Key, error) {
	if ks.p == nil {
		return nil, ErrUninitialized
	}
	return ks.p.GetKey(a, p)
}

/*
Deletes the entry identified by the given alias from this keystore.
*/
func (ks *Keystore) Delete(a Alias) error {
	if ks.p == nil {
		return ErrUninitialized
	}
	return ks.p.Delete(a)
}

/*
Loads this KeyStore from the given input stream.
*/
func (ks *Keystore) Load(d []byte, passphrase []byte) error {
	if ks.p == nil {
		return ErrUninitialized
	}
	ks.p.Load(d, passphrase)
	return nil
}

/*
Loads this KeyStore from the given file path
*/
func (ks *Keystore) LoadFile(f string, passphrase []byte) error {
	if ks.p == nil {
		return ErrUninitialized
	}
	ks.p.LoadFile(f, passphrase)
	return nil
}

/*
* Stores this keystore to the output stream, and protects its
* integrity with the given password.
 */
func (ks *Keystore) Store(passphrase []byte) (out []byte, err error) {
	if ks.p == nil {
		return nil, ErrUninitialized
	}
	return ks.p.Store(passphrase)
}

/*
* Stores this keystore to the given file, and protects its
* integrity with the given password.
 */
func (ks *Keystore) StoreFile(f string, passphrase []byte) error {
	if ks.p == nil {
		return ErrUninitialized
	}
	ks.p.StoreFile(f, passphrase)
	return nil
}
