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
	"errors"
)

type ECDSASignature struct {
	privateKey *ecdsa.PrivateKey

	publicKey *ecdsa.PublicKey
}

func (s *ECDSASignature) InitSign(priByte []byte) error {
	pri, err := ToECDSAPrivate(priByte)
	if err != nil {
		return err
	}
	s.privateKey = pri
	return nil
}

func (s *ECDSASignature) Sign(data []byte) (out []byte, err error) {
	if s.privateKey == nil {
		return nil, errors.New("please call InitSign to set private key")
	}
	signature, err := Sign(data, s.privateKey)
	if err != nil {
		return nil, err
	}
	return signature, nil
}

func (s *ECDSASignature) InitVerify(pubByte []byte) error {
	pub, _ := ToECDSAPublic(pubByte)
	s.publicKey = pub
	return nil
}

func (s *ECDSASignature) Verify(data []byte, signature []byte) (bool, error) {
	if s.publicKey == nil {
		return false, errors.New("please call InitVerify to set public key")
	}
	result := Verify(data, signature, s.publicKey)
	return result, nil
}
