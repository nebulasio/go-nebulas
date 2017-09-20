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

package key

import (
	"errors"
)

var (
	// ErrNeedPassphrase need passphrase error.
	ErrNeedPassphrase = errors.New("keystore: PassphraseProtection need passphrase")

	// ErrNeedCallback need callback func
	ErrNeedCallback = errors.New("keystore: CallbackHandlerProtection need callback func")
)

// Alias for keystore
type Alias string

// Alias for keystore
//type Alias struct {
//	Alias []byte
//}
//
//
//// generate a Alias for provider use
//func NewAlias(a []byte) *Alias {
//	alias := &Alias{Alias: a}
//	return alias
//}

// ProtectionParameter A marker interface for keystore protection parameters.
//
// The information stored in a ProtectionParameter
// object protects the contents of a keystore.
// For example, protection parameters may be used to check
// the integrity of keystore data, or to protect the
// confidentiality of sensitive keystore data
// (such as a PrivateKey).
type ProtectionParameter interface {
	Protection(p []byte) error
}

// PassphraseProtection a password-based implementation of ProtectionParameter
type PassphraseProtection struct {
	passphrase []byte
}

// Protection construct
func (pro *PassphraseProtection) Protection(p []byte) error {
	if len(p) == 0 {
		return ErrNeedPassphrase
	}
	pro.passphrase = p
	return nil
}

// NewPassphrase generate PassphraseProtection
func NewPassphrase(p []byte) (*PassphraseProtection, error) {
	if len(p) == 0 {
		return nil, ErrNeedPassphrase
	}
	pro := &PassphraseProtection{passphrase: p}
	return pro, nil
}

// CallbackHandlerProtection a ProtectionParameter encapsulating a CallbackHandler,hardware encryption may use it
type CallbackHandlerProtection struct {
	handler func()
}

// Protection construct
func (pro *CallbackHandlerProtection) Protection(p []byte) error {
	return nil
}

// NewCallback generate CallbackHandlerProtection
func NewCallback(f func()) (*CallbackHandlerProtection, error) {
	if f == nil {
		return nil, ErrNeedCallback
	}
	pro := &CallbackHandlerProtection{handler: f}
	return pro, nil
}
