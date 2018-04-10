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
	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/crypto/utils"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

// PrivateKey ecdsa privatekey
type PrivateKey struct {
	seckey []byte
}

// GeneratePrivateKey generate a new private key
func GeneratePrivateKey() *PrivateKey {
	priv := new(PrivateKey)
	seckey := NewSeckey()
	priv.seckey = seckey
	return priv
}

// Algorithm algorithm name
func (k *PrivateKey) Algorithm() keystore.Algorithm {
	return keystore.SECP256K1
}

// Encoded encoded to byte
func (k *PrivateKey) Encoded() ([]byte, error) {
	return k.seckey, nil
}

// Decode decode data to key
func (k *PrivateKey) Decode(data []byte) error {
	if SeckeyVerify(data) == false {
		return ErrInvalidPrivateKey
	}
	k.seckey = data
	return nil
}

// Clear clear key content
func (k *PrivateKey) Clear() {
	utils.ZeroBytes(k.seckey)
}

// PublicKey returns publickey
func (k *PrivateKey) PublicKey() keystore.PublicKey {
	pub, err := GetPublicKey(k.seckey)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to get public key.")
		return nil
	}
	return NewPublicKey(pub)
}

// Sign sign hash with privatekey
func (k *PrivateKey) Sign(hash []byte) ([]byte, error) {
	return Sign(hash, k.seckey)
}
