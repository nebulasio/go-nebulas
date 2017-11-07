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
	"errors"

	"github.com/nebulasio/go-nebulas/crypto/keystore"
)

// Signature signature ecdsa
type Signature struct {
	privateKey *PrivateKey

	publicKey *PublicKey
}

// Algorithm secp256k1 algorithm
func (s *Signature) Algorithm() keystore.Algorithm {
	return keystore.SECP256K1
}

// InitSign ecdsa init sign
func (s *Signature) InitSign(priv keystore.PrivateKey) error {
	s.privateKey = priv.(*PrivateKey)
	return nil
}

// Sign ecdsa sign
func (s *Signature) Sign(data []byte) (out []byte, err error) {
	if s.privateKey == nil {
		return nil, errors.New("please get private key first")
	}
	signature, err := s.privateKey.Sign(data)
	if err != nil {
		return nil, err
	}
	return signature, nil
}

// RecoverPublic returns a public key
func (s *Signature) RecoverPublic(data []byte, signature []byte) (keystore.PublicKey, error) {
	pub, err := RecoverECDSAPublicKey(data, signature)
	if err != nil {
		return nil, err
	}
	s.publicKey = NewPublicKey(*pub)
	return s.publicKey, nil
}

// InitVerify ecdsa verify init
func (s *Signature) InitVerify(pub keystore.PublicKey) error {
	s.publicKey = pub.(*PublicKey)
	return nil
}

// Verify ecdsa verify
func (s *Signature) Verify(data []byte, signature []byte) (bool, error) {
	if s.publicKey == nil {
		return false, errors.New("please give public key first")
	}
	return s.publicKey.Verify(data, signature)
}
