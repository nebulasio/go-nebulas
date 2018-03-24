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

var (
	// DefaultKS generate a default keystore
	DefaultKS = NewKeystore()

	// DefaultUnlockDuration default lock 300s
	DefaultUnlockDuration = time.Duration(300 * time.Second)

	// YearUnlockDuration lock 1 year time
	YearUnlockDuration = time.Duration(365 * 24 * 60 * 60 * time.Second)
)

var (
	// ErrUninitialized uninitialized provider error.
	ErrUninitialized = errors.New("uninitialized the provider")

	// ErrNotUnlocked key not unlocked
	ErrNotUnlocked = errors.New("key not unlocked")

	// ErrInvalidPassphrase invalid passphrase
	ErrInvalidPassphrase = errors.New("passphrase is invalid")
)

// unlock item
type unlocked struct {
	alias string

	key Key

	timer *time.Timer
}

// Keystore class represents a storage facility for cryptographic keys
type Keystore struct {

	// keystore provider
	p Provider

	// unlocked items
	unlocked map[string]*unlocked

	mu sync.RWMutex
}

// NewKeystore new
func NewKeystore() *Keystore {
	ks := &Keystore{}

	ks.unlocked = make(map[string]*unlocked)
	ks.p = NewMemoryProvider(1.0, SCRYPT)
	return ks
}

// Aliases lists all the alias names of this keystore.
func (ks *Keystore) Aliases() []string {
	return ks.p.Aliases()
}

// ContainsAlias checks if the given alias exists in this keystore.
func (ks *Keystore) ContainsAlias(a string) (bool, error) {
	if ks.p == nil {
		return false, ErrUninitialized
	}

	return ks.p.ContainsAlias(a)
}

// Unlock unlock key with ProtectionParameter
func (ks *Keystore) Unlock(alias string, passphrase []byte, timeout time.Duration) error {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	key, err := ks.p.GetKey(alias, passphrase)
	if err != nil {
		return err
	}

	unlockedKey, ok := ks.unlocked[alias]
	if ok == true {
		unlockedKey.key = key
		unlockedKey.timer.Reset(timeout)
	} else {
		u := &unlocked{alias, key, time.NewTimer(timeout)}
		ks.unlocked[alias] = u
		go ks.expire(alias)
	}
	return nil
}

// Lock lock key
func (ks *Keystore) Lock(alias string) error {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	if u, ok := ks.unlocked[alias]; ok == true {
		u.timer.Reset(time.Duration(0) * time.Nanosecond)
		return nil
	}

	return ErrNotUnlocked
}

func (ks *Keystore) expire(alias string) {
	if u, ok := ks.unlocked[alias]; ok == true {
		defer u.timer.Stop()
		select {
		case <-u.timer.C:
			ks.mu.Lock()
			u.key.Clear()
			delete(ks.unlocked, alias)
			ks.mu.Unlock()
		}
	}
}

// GetUnlocked returns a unlocked key
func (ks *Keystore) GetUnlocked(alias string) (Key, error) {
	if len(alias) == 0 {
		return nil, ErrNeedAlias
	}

	ks.mu.RLock()
	defer ks.mu.RUnlock()

	key, ok := ks.unlocked[alias]
	if ok == false {
		return nil, ErrNotUnlocked
	}

	return key.key, nil
}

// SetKey assigns the given key to the given alias, protecting it with the given passphrase.
func (ks *Keystore) SetKey(a string, k Key, passphrase []byte) error {
	if ks.p == nil {
		return ErrUninitialized
	}

	return ks.p.SetKey(a, k, passphrase)
}

// GetKey returns the key associated with the given alias, using the given
// password to recover it.
func (ks *Keystore) GetKey(a string, passphrase []byte) (Key, error) {
	if ks.p == nil {
		return nil, ErrUninitialized
	}

	key, err := ks.p.GetKey(a, passphrase)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// Delete the entry identified by the given alias from this keystore.
func (ks *Keystore) Delete(a string, passphrase []byte) error {
	if ks.p == nil {
		return ErrUninitialized
	}

	key, err := ks.p.GetKey(a, passphrase)
	if err != nil {
		return err
	}
	key.Clear()

	if _, ok := ks.unlocked[a]; ok == true {
		ks.Lock(a)
	}

	return ks.p.Delete(a)
}
