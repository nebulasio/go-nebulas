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

package secp256k1

import (
	"crypto/ecdsa"

	"github.com/nebulasio/go-nebulas/crypto/keystore"
)

// PublicKey ecdsa publickey
type PublicKey struct {
	publickey ecdsa.PublicKey
}

// NewPublicKey generate PublicKey
func NewPublicKey(pub ecdsa.PublicKey) *PublicKey {
	ecdsaPub := &PublicKey{pub}
	return ecdsaPub
}

// Algorithm algorithm name
func (k *PublicKey) Algorithm() keystore.Algorithm {
	return keystore.SECP256K1
}

// Encoded encoded to byte
func (k *PublicKey) Encoded() ([]byte, error) {
	return FromECDSAPublicKey(&k.publickey)
}

// Decode decode data to key
func (k *PublicKey) Decode(data []byte) error {
	pub, err := ToECDSAPublicKey(data)
	if err != nil {
		return err
	}
	k.publickey = *pub
	return nil
}

// Clear clear key content
func (k *PublicKey) Clear() {
	k.publickey = ecdsa.PublicKey{}
}

// Verify verify ecdsa publickey
func (k *PublicKey) Verify(hash []byte, signature []byte) (bool, error) {
	return Verify(hash, signature, &k.publickey)
}
