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

package core

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/satori/go.uuid"
)

type Account struct {
	// uuid is random for unique id
	id string

	// nick for user account
	nick string

	address Address

	// ecdsa private key ,public & address can be derived from it
	privateKey ecdsa.PrivateKey
}

// create new account with nick and private key
func newAccountFromECDSA(nick string, privateKeyECDSA *ecdsa.PrivateKey) *Account {
	id := fmt.Sprintf("%s", uuid.NewV4())
	addr := NewAddressWithPrivateKey(privateKeyECDSA)
	account := &Account{
		id:         id,
		nick:       nick,
		address:    *addr,
		privateKey: *privateKeyECDSA,
	}
	return account
}
