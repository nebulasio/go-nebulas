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

package cipher

import (
	"errors"

	"crypto/rand"

	"github.com/nebulasio/go-nebulas/crypto/encrypt"
	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/crypto/keystore/ecdsa"

	goecdsa "crypto/ecdsa"
)

// Algorithm type alias
type Algorithm uint8

const (
	// SECP256K1 a type of signer
	SECP256K1 Algorithm = 1

	// SCRYPT a type of encrypt
	SCRYPT Algorithm = 1 << 4
)

var (
	// ErrAlgorithmInvalid invalid Algorithm for sign.
	ErrAlgorithmInvalid = errors.New("invalid Algorithm")
)

// NewPrivateKey generate a privatekey with Algorithm
func NewPrivateKey(alg Algorithm, data []byte) (keystore.PrivateKey, error) {
	switch alg {
	case SECP256K1:
		var (
			priv *goecdsa.PrivateKey
			err  error
		)
		if len(data) == 0 {
			priv, err = ecdsa.NewPrivateKey(rand.Reader)
		} else {
			priv, err = ecdsa.ToPrivateKey(data)
		}
		if err != nil {
			return nil, err
		}
		key := ecdsa.NewPrivateStoreKey(priv)
		return key, nil
	default:
		return nil, ErrAlgorithmInvalid
	}
}

// GetSignature returns the specified algorithm Signature
func GetSignature(alg Algorithm) (keystore.Signature, error) {
	switch alg {
	case SECP256K1:
		secp256k1 := &ecdsa.Signature{}
		return secp256k1, nil
	default:
		return nil, ErrAlgorithmInvalid
	}
}

// GetEncrypt returns the specified algorithm Encrpt
func GetEncrypt(alg Algorithm) (encrypt.Encrypt, error) {
	switch alg {
	case SCRYPT:
		encrypt := &encrypt.Scrypt{}
		return encrypt, nil
	default:
		return nil, ErrAlgorithmInvalid
	}
}
