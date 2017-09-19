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

import "errors"

var (
	// ErrNeedPassphrase need passphrase error.
	ErrNeedPassphrase = errors.New("keystore: PassphraseProtection need passphrase")

	ErrNeedCallback = errors.New("keystore: CallbackHandlerProtection need callback func")
)

type Alias struct {
	alias []byte
}

// generate a Alias for provider use
func NewAlias(a []byte) *Alias {
	alias := &Alias{alias: a}
	return alias
}

/*
* A marker interface for keystore protection parameters.
*
* <p> The information stored in a <code>ProtectionParameter</code>
* object protects the contents of a keystore.
* For example, protection parameters may be used to check
* the integrity of keystore data, or to protect the
* confidentiality of sensitive keystore data
* (such as a <code>PrivateKey</code>).
 */
type ProtectionParameter interface {
	Protection(p []byte) error
}

/*
A password-based implementation of ProtectionParameter
*/
type PassphraseProtection struct {
	passphrase []byte
}

func (pro *PassphraseProtection) Protection(p []byte) error {
	if len(p) == 0 {
		return ErrNeedPassphrase
	}
	pro.passphrase = p
	return nil
}

func NewPassphraseParameter(p []byte) (*PassphraseProtection, error) {
	if len(p) == 0 {
		return nil, ErrNeedPassphrase
	}
	pro := &PassphraseProtection{passphrase: p}
	return pro, nil
}

/*
A ProtectionParameter encapsulating a CallbackHandler,hardware encryption may use it
*/
type CallbackHandlerProtection struct {
	handler func()
}

func (pro *CallbackHandlerProtection) Protection(p []byte) error {
	return nil
}

func NewCallbackParameter(f func()) (*CallbackHandlerProtection, error) {
	if f == nil {
		return nil, ErrNeedCallback
	}
	pro := &CallbackHandlerProtection{handler: f}
	return pro, nil
}

/*
* This class represents a "provider" for the
 * Security API, where a provider implements some or all parts of
 * Security. Services that a provider may implement include:
 *
 * <ul>
 *
 * <li>Algorithms
 *
 * <li>Key generation, conversion, and management facilities (such as for
 * algorithm-specific keys).
 *
 *</ul>
 *
 * <p>Each provider has a name and a version number, and is configured
 * in each runtime it is installed in.
*/
type Provider interface {

	/*
			 * Constructs a provider with the specified name, version number,
		     * and information.
	*/
	NewProvider(name string, version float32, desc string)

	// all entry in provider save
	Entries() map[Alias]*Key

	SetKeyPassphrase(a Alias, k Key, passphrase []byte) error

	SetKey(a Alias, val []byte) error

	// get key
	GetKey(a Alias, p ProtectionParameter) (Key, error)

	// remove key
	Delete(a Alias) error

	// check provider contains key
	ContainsAlias(a Alias) (bool, error)

	// clear all entries in provider
	Clear() error

	Load(d []byte, passphrase []byte) error

	LoadFile(f string, passphrase []byte) error

	Store(passphrase []byte) (out []byte, err error)

	StoreFile(f string, passphrase []byte) error
}
