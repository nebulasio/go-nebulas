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

package crypto

import (
	"errors"

	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/crypto/keystore/secp256k1"
)

var (
	// ErrAlgorithmInvalid invalid Algorithm for sign.
	ErrAlgorithmInvalid = errors.New("invalid Algorithm")
)

// NewPrivateKey generate a privatekey with Algorithm
func NewPrivateKey(alg keystore.Algorithm, data []byte) (keystore.PrivateKey, error) {
	switch alg {
	case keystore.SECP256K1:
		var (
			priv *secp256k1.PrivateKey
			err  error
		)
		if len(data) == 0 {
			priv = secp256k1.GeneratePrivateKey()
		} else {
			priv = new(secp256k1.PrivateKey)
			err = priv.Decode(data)
		}
		if err != nil {
			return nil, err
		}
		return priv, nil
	default:
		return nil, ErrAlgorithmInvalid
	}
}

// NewSignature returns a specific signature with the algorithm
func NewSignature(alg keystore.Algorithm) (keystore.Signature, error) {
	switch alg {
	case keystore.SECP256K1:
		return new(secp256k1.Signature), nil
	default:
		return nil, ErrAlgorithmInvalid
	}
}
