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
	unlocked []unlocked

	mu sync.RWMutex
}

// NewKeystore new
func NewKeystore() *Keystore {
	ks := &Keystore{}
	ks.unlocked = []unlocked{}
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
	key, err := ks.p.GetKey(alias, passphrase)
	if err != nil {
		return err
	}
	ks.mu.Lock()
	defer ks.mu.Unlock()

	hasUnlocked := false
	for _, u := range ks.unlocked {
		if u.alias == alias {
			u.key = key
			u.timer.Reset(timeout)
			hasUnlocked = true
			break
		}
	}
	if !hasUnlocked {
		u := unlocked{alias, key, time.NewTimer(timeout)}
		ks.unlocked = append(ks.unlocked, u)
		go ks.expire(alias)
	}
	return nil
}

// Lock lock key
func (ks *Keystore) Lock(alias string) error {

	ks.mu.Lock()
	defer ks.mu.Unlock()
	for _, u := range ks.unlocked {
		if u.alias == alias {
			u.timer.Reset(time.Duration(0) * time.Nanosecond)
			return nil
		}
	}

	return ErrNotUnlocked
}

func (ks *Keystore) expire(alias string) {
	var (
		u   *unlocked
		idx int
	)
	for idx = 0; idx < len(ks.unlocked); idx++ {
		if ks.unlocked[idx].alias == alias {
			u = &ks.unlocked[idx]
			break
		}
	}
	if u == nil {
		return
	}
	defer u.timer.Stop()
	select {
	case <-u.timer.C:
		ks.mu.Lock()
		u.key.Clear()
		if idx < len(ks.unlocked)-1 {
			ks.unlocked = append(ks.unlocked[:idx], ks.unlocked[idx+1:]...)
		} else {
			ks.unlocked = ks.unlocked[:idx]
		}
		ks.mu.Unlock()
	}
}

// GetUnlocked returns a unlocked key
func (ks *Keystore) GetUnlocked(alias string) (Key, error) {
	if len(alias) == 0 {
		return nil, ErrNeedAlias
	}
	for _, u := range ks.unlocked {
		if u.alias == alias {
			return u.key, nil
		}
	}
	return nil, ErrNotUnlocked
}

// SetKey assigns the given key to the given alias, protecting it with the given passphrase.
func (ks *Keystore) SetKey(a string, k Key, passphrase []byte) error {
	if ks.p == nil {
		return ErrUninitialized
	}
	ks.mu.RLock()
	defer ks.mu.RUnlock()
	return ks.p.SetKey(a, k, passphrase)
}

// GetKeyByIndex returns the key associated with the given index in unlocked
func (ks *Keystore) GetKeyByIndex(idx int) (string, Key, error) {
	if idx < 0 || idx >= len(ks.unlocked) {
		return "", nil, ErrNotUnlocked
	}
	u := ks.unlocked[idx]
	return u.alias, u.key, nil
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
	return ks.p.Delete(a)
}
