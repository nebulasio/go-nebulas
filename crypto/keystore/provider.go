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

// Provider class represents a "provider" for the
// Security API, where a provider implements some or all parts of
// Security. Services that a provider may implement include:
// Algorithms
// Key generation, conversion, and management facilities (such as for
// algorithm-specific keys).
// Each provider has a name and a version number, and is configured
// in each runtime it is installed in.
type Provider interface {

	// Aliases all alias in provider save
	Aliases() []string

	// SetKey assigns the given key (that has already been protected) to the given alias.
	SetKey(a string, key Key, passphrase []byte) error

	// GetKey returns the key associated with the given alias, using the given
	// password to recover it.
	GetKey(a string, passphrase []byte) (Key, error)

	// Delete remove key
	Delete(a string) error

	// ContainsAlias check provider contains key
	ContainsAlias(a string) (bool, error)

	// Clear all entries in provider
	Clear() error
}
