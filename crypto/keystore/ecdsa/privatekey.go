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
	"crypto/ecdsa"
)

// PrivateStoreKey ecdsa privatekey
type PrivateStoreKey struct {
	privateKey *ecdsa.PrivateKey
}

// NewPrivateStoreKey generate PrivateStoreKey
func NewPrivateStoreKey(pri *ecdsa.PrivateKey) *PrivateStoreKey {
	ecdsaPri := &PrivateStoreKey{pri}
	return ecdsaPri
}

// Algorithm algorithm name
func (k *PrivateStoreKey) Algorithm() string {
	return "ecdsa"
}

// Format formate
func (k *PrivateStoreKey) Format() string {
	return "byte"
}

// Encoded encoded to byte
func (k *PrivateStoreKey) Encoded() ([]byte, error) {
	return FromPrivateKey(k.privateKey)
}

// EncodedPub encode publickey
func (k *PrivateStoreKey) EncodedPub() ([]byte, error) {
	return FromPublicKey(&k.privateKey.PublicKey)
}
