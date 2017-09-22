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
	// ErrNeedAlias need alias
	ErrNeedAlias = errors.New("need alias")

	// ErrNeedFilePath need file path
	ErrNeedFilePath = errors.New("need file path")

	// ErrNeedPassphrase need passphrase
	ErrNeedPassphrase = errors.New("need passphrase")
)

// MemoryProvider handle keystore with ecdsa
type MemoryProvider struct {

	// name of ecdsa provider
	name string

	// version of ecdsa provider
	version float32

	// a map storage entry
	entries map[string]Key
}

// NewMemoryProvider generate a provider with version
func NewMemoryProvider(v float32) *MemoryProvider {
	p := &MemoryProvider{"memory", v, make(map[string]Key)}
	return p
}

// Aliases all entry in provider save
func (p *MemoryProvider) Aliases() []string {
	aliases := make([]string, len(p.entries))
	for k := range p.entries {
		aliases = append(aliases, k)
	}
	return aliases
}

// SetKey assigns the given key (that has already been protected) to the given alias.
func (p *MemoryProvider) SetKey(a string, key Key) error {
	if &a == nil {
		return ErrNeedAlias
	}
	p.entries[a] = key
	return nil
}

// GetKey returns the key associated with the given alias, using the given
// password to recover it.
func (p *MemoryProvider) GetKey(a string) (Key, error) {
	if &a == nil {
		return nil, ErrNeedAlias
	}

	key := p.entries[a]
	if key == nil {
		return nil, errors.New("not find in provider")
	}
	return key, nil
}

// Delete remove key
func (p *MemoryProvider) Delete(a string) error {
	if &a == nil {
		return ErrNeedAlias
	}
	delete(p.entries, a)
	return nil
}

// ContainsAlias check provider contains key
func (p *MemoryProvider) ContainsAlias(a string) (bool, error) {
	if &a == nil {
		return false, ErrNeedAlias
	}
	for k := range p.entries {
		if k == a {
			return true, nil
		}
	}
	return false, errors.New("not contains alias")
}

// Clear clear all entries in provider
func (p *MemoryProvider) Clear() error {
	if p.entries == nil {
		return errors.New("need entries map")
	}
	p.entries = make(map[string]Key)
	return nil
}
