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
	"crypto/elliptic"
	"github.com/nebulasio/go-nebulas/crypto/keystore/ecdsa/bitelliptic"
	"io"

	"crypto/rand"
	"encoding/hex"
	"errors"
	"math/big"
)

// S256 returns an instance of the secp256k1 curve.
func S256() elliptic.Curve {
	return bitelliptic.S256()
}

// NewPrivateKey generate a ecdsa private key
func NewPrivateKey(rand io.Reader) (*ecdsa.PrivateKey, error) {
	privateKeyECDSA, err := ecdsa.GenerateKey(S256(), rand)
	if err != nil {
		return nil, err
	}
	return privateKeyECDSA, nil
}

// FromPrivateKey exports a private key into a binary dump.
func FromPrivateKey(pri *ecdsa.PrivateKey) ([]byte, error) {
	if pri == nil {
		return nil, errors.New("ecdsa: please input private key")
	}
	return pri.D.Bytes(), nil
}

// FromPublicKey exports a public key into a binary dump.
func FromPublicKey(pub *ecdsa.PublicKey) ([]byte, error) {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil, errors.New("ecdsa: please input public key")
	}
	return elliptic.Marshal(S256(), pub.X, pub.Y), nil
}

// HexToPrivateKey parses a secp256k1 private key.
func HexToPrivateKey(hexkey string) (*ecdsa.PrivateKey, error) {
	b, err := hex.DecodeString(hexkey)
	if err != nil {
		return nil, errors.New("invalid hex string")
	}
	return ToPrivateKey(b)
}

// ToPrivateKey creates a private key with the given data value.
func ToPrivateKey(d []byte) (*ecdsa.PrivateKey, error) {
	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = S256()
	//if 8*len(d) != priv.Params().BitSize {
	//	return nil, errors.New("invalid length")
	//}
	priv.D = new(big.Int).SetBytes(d)
	priv.PublicKey.X, priv.PublicKey.Y = priv.PublicKey.Curve.ScalarBaseMult(d)
	return priv, nil
}

// ToPublicKey creates a public key with the given data value.
func ToPublicKey(pub []byte) (*ecdsa.PublicKey, error) {
	if len(pub) == 0 {
		return nil, errors.New("ecdsa: please input public key bytes")
	}
	x, y := elliptic.Unmarshal(S256(), pub)
	return &ecdsa.PublicKey{Curve: S256(), X: x, Y: y}, nil
}

// Sign sign hash with private key
func Sign(hash []byte, prv *ecdsa.PrivateKey) (sig []byte, err error) {
	r, s, err := ecdsa.Sign(rand.Reader, prv, hash)
	if err != nil {
		return nil, errors.New("ecdsa: sign err")
	}

	sign := r.Bytes()
	sign = append(sign, s.Bytes()...)
	return sign, nil
}

// Verify verify with public key
func Verify(hash []byte, rs []byte, pub *ecdsa.PublicKey) bool {
	r := new(big.Int)
	r.SetBytes(rs[:32])
	s := new(big.Int)
	s.SetBytes(rs[32:])

	return ecdsa.Verify(pub, hash, r, s)
}
