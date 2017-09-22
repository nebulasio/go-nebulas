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
	"sync"
	"time"
)

const (
	// KeystoreTypeDefault default keystore
	KeystoreTypeDefault = "default"
)

var (
	// DefaultKS generate a default keystore
	DefaultKS = NewKeystore()
)

var (
	// ErrUninitialized uninitialized provider error.
	ErrUninitialized = errors.New("uninitialized the provider")
)

// Keystore class represents a storage facility for cryptographic keys
type Keystore struct {

	// keystore provider
	p Provider

	// unlocked aliases，as map range returns unordered
	unlockedalias []string

	// unlocked map
	unlocked map[string]Key

	mu sync.RWMutex
}

// NewKeystore new
func NewKeystore() *Keystore {
	ks := &Keystore{unlockedalias: []string{}, unlocked: make(map[string]Key)}
	ks.p = NewMemoryProvider(1.0)
	return ks
}

// Aliases lists all the alias names of this keystore.
func (ks *Keystore) Aliases() ([]string, error) {
	if ks.p == nil {
		return nil, ErrUninitialized
	}
	return ks.p.Aliases(), nil
}

// ContainsAlias checks if the given alias exists in this keystore.
func (ks *Keystore) ContainsAlias(a string) (bool, error) {
	if ks.p == nil {
		return false, ErrUninitialized
	}
	return ks.p.ContainsAlias(a)
}

// SetKey assigns the given key to the given alias, protecting it with the given passphrase.
func (ks *Keystore) SetKey(a string, k Key, passphrase []byte) error {
	if ks.p == nil {
		return ErrUninitialized
	}
	//TODO(larry.wang):lock k before provider SetKey call
	return ks.p.SetKey(a, k)
}

// Unlock unlock key with ProtectionParameter
func (ks *Keystore) Unlock(alias string, passphrase []byte, timeout time.Duration) error {
	key, err := ks.p.GetKey(alias)
	if err != nil {
		return err
	}
	ks.mu.Lock()
	defer ks.mu.Unlock()

	unlocked := false
	for _, a := range ks.unlockedalias {
		if a == alias {
			unlocked = true
			break
		}
	}
	// if alias has not unlocked
	if !unlocked {
		ks.unlockedalias = append(ks.unlockedalias, alias)
	}
	//TODO(larry.wang):unlock k
	ks.unlocked[alias] = key
	return nil
}

// GetUnlocked returns a unlocked key
func (ks *Keystore) GetUnlocked(alias string) (Key, error) {
	if len(alias) == 0 {
		return nil, errors.New("need alias")
	}
	for a := range ks.unlocked {
		if a == alias {
			return ks.unlocked[a], nil
		}
	}
	return nil, errors.New("not find key")
}

// GetKeyByIndex returns the key associated with the given index in unlocked
func (ks *Keystore) GetKeyByIndex(idx int) (string, Key, error) {
	if idx < 0 || idx > len(ks.unlocked) {
		return "", nil, errors.New("index out of range")
	}
	alias := ks.unlockedalias[idx]
	return alias, ks.unlocked[alias], nil
}

// GetKey returns the key associated with the given alias, using the given
// password to recover it.
func (ks *Keystore) GetKey(a string, passphrase []byte) (Key, error) {
	if ks.p == nil {
		return nil, ErrUninitialized
	}
	key, err := ks.p.GetKey(a)
	if err != nil {
		return nil, err
	}
	//TODO(larry.wang):unlock k after get from provider
	return key, nil
}

// Delete the entry identified by the given alias from this keystore.
func (ks *Keystore) Delete(a string) error {
	if ks.p == nil {
		return ErrUninitialized
	}
	return ks.p.Delete(a)
}
