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

package ecdsa

import (
	"errors"
	"github.com/nebulasio/go-nebulas/crypto/keystore/key"
	"io/ioutil"
	"os"
)

var (
	// ErrNeedAlias need alias
	ErrNeedAlias = errors.New("Provider: need alias")

	// ErrNeedFilePath need file path
	ErrNeedFilePath = errors.New("Provider: need file path")

	// ErrNeedPassphrase need passphrase
	ErrNeedPassphrase = errors.New("Provider: need passphrase")
)

// EntryType entry type
type EntryType int

// enum EntryType
const (
	_ EntryType = iota
	EntryPrivate
	EntryPublic
)

// Entry ecdsa entry
type Entry struct {

	// enum of key,like EntryPrivate/EntryPublic
	etype EntryType

	// encrypt key
	data []byte
}

// Provider handle keystore with ecdsa
type Provider struct {

	// name of ecdsa provider
	name string

	// version of ecdsa provider
	version float32

	// a map storage entry
	entries map[key.Alias]*Entry
}

// NewProvider generate a provider with version
func NewProvider(v float32) *Provider {
	p := &Provider{"ecdsa", v, make(map[key.Alias]*Entry)}
	return p
}

// Aliases all entry in provider save
func (p *Provider) Aliases() []key.Alias {
	aliases := make([]key.Alias, len(p.entries))
	for k := range p.entries {
		aliases = append(aliases, k)
	}
	return aliases
}

// SetKeyPassphrase assigns the given key to the given alias, protecting it with the given passphrase.
func (p *Provider) SetKeyPassphrase(a key.Alias, k key.Key, passphrase []byte) error {
	if &a == nil {
		return ErrNeedAlias
	}
	if len(passphrase) == 0 {
		return ErrNeedPassphrase
	}
	data, err := k.Encoded()
	if err != nil {
		return err
	}
	//TODO(larry.wang): impl the encrypt the data with passphrase
	var entry = &Entry{}
	switch k.(type) {
	case *PrivateStoreKey:
		entry = &Entry{EntryPrivate, data}
	case *PublicStoreKey:
		entry = &Entry{EntryPublic, data}
	default:
		return errors.New("Provider: key is incorrect for ecdsa")
	}
	p.entries[a] = entry
	return nil
}

// SetKey assigns the given key (that has already been protected) to the given alias.
func (p *Provider) SetKey(a key.Alias, val []byte) error {
	return errors.New("Provider: use  SetKeyPassphrase to setKey")
}

// GetKey returns the key associated with the given alias, using the given
// password to recover it.
func (p *Provider) GetKey(a key.Alias, param key.ProtectionParameter) (key.Key, error) {
	if &a == nil {
		return nil, ErrNeedAlias
	}
	if param == nil {
		return nil, ErrNeedPassphrase
	}

	//if ph,ok:=param.(key.PassphraseProtection);ok{
	//}
	_, err := p.ContainsAlias(a)
	if err != nil {
		return nil, err
	}
	var key key.Key
	entry := p.entries[a]
	//TODO(larry.wang): impl the decrypt the data with passphrase
	data := entry.data
	switch entry.etype {
	case EntryPrivate:
		priv, err := ToPrivateKey(data)
		if err != nil {
			return nil, err
		}
		key = &PrivateStoreKey{priv}
	case EntryPublic:
		pub, err := ToPublicKey(data)
		if err != nil {
			return nil, err
		}
		key = &PublicStoreKey{pub}
	default:
		return nil, errors.New("Provider: entry type incorrect")
	}

	return key, nil
}

// Delete remove key
func (p *Provider) Delete(a key.Alias) error {
	if &a == nil {
		return ErrNeedAlias
	}
	delete(p.entries, a)
	return nil
}

// ContainsAlias check provider contains key
func (p *Provider) ContainsAlias(a key.Alias) (bool, error) {
	if &a == nil {
		return false, ErrNeedAlias
	}
	for k := range p.entries {
		//if byteutils.Equal(k.Alias, a.Alias) {
		//	return true, nil
		//}
		if k == a {
			return true, nil
		}
	}
	return false, errors.New("Provider: not contains alias")
}

// Clear clear all entries in provider
func (p *Provider) Clear() error {
	if p.entries == nil {
		return errors.New("Provider: need entries map")
	}
	p.entries = make(map[key.Alias]*Entry)
	return nil
}

// Load loads this KeyStore from the given input stream.
func (p *Provider) Load(d []byte, passphrase []byte) error {
	if len(d) == 0 {
		return ErrNeedFilePath
	}
	if len(passphrase) == 0 {
		return ErrNeedPassphrase
	}
	//TODO(larry.wang):impl keystore unserialize with passphrase
	return nil
}

// LoadFile Load this KeyStore from the given file path
func (p *Provider) LoadFile(f string, passphrase []byte) error {
	if len(f) == 0 {
		return ErrNeedFilePath
	}
	if len(passphrase) == 0 {
		return ErrNeedPassphrase
	}
	data, err := ioutil.ReadFile(f)
	if err != nil {
		return errors.New("Provider: load file failed")
	}
	return p.Load(data, passphrase)
}

// Store stores this keystore to the output stream, and protects its
// integrity with the given password.
func (p *Provider) Store(passphrase []byte) (out []byte, err error) {
	if len(passphrase) == 0 {
		return nil, ErrNeedPassphrase
	}
	//TODO(larry.wang):impl keystore serialize with passphrase
	return nil, nil
}

// StoreFile stores this keystore to the given file, and protects its
// integrity with the given password.
func (p *Provider) StoreFile(f string, passphrase []byte) error {
	if len(f) == 0 {
		return ErrNeedFilePath
	}
	data, err := p.Store(passphrase)
	if err != nil {
		return errors.New("Provider: store to stream failed")
	}
	return ioutil.WriteFile(f, data, os.ModeAppend)
}
