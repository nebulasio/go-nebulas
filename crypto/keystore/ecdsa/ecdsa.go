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
	"io"

	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"math/big"
)

// S256 returns an instance of the secp256k1 curve.
func S256() elliptic.Curve {
	return secp256k1.S256()
}

// generate a ecdsa private key
func GenerateECDSAPrivateKey(rand io.Reader) (*ecdsa.PrivateKey, error) {
	privateKeyECDSA, err := ecdsa.GenerateKey(S256(), rand)
	if err != nil {
		return nil, err
	}
	return privateKeyECDSA, nil
}

// FromECDSAPri exports a private key into a binary dump.
func FromECDSAPri(pri *ecdsa.PrivateKey) []byte {
	if pri == nil {
		return nil
	}
	return pri.D.Bytes()
}

// FromECDSAPub exports a public key into a binary dump.
func FromECDSAPub(pub *ecdsa.PublicKey) []byte {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil
	}
	return elliptic.Marshal(S256(), pub.X, pub.Y)
}

// HexToECDSAPrivate parses a secp256k1 private key.
func HexToECDSAPrivate(hexkey string) (*ecdsa.PrivateKey, error) {
	b, err := hex.DecodeString(hexkey)
	if err != nil {
		return nil, errors.New("invalid hex string")
	}
	return ToECDSAPrivate(b)
}

// ToECDSAPrivate creates a private key with the given data value.
func ToECDSAPrivate(d []byte) (*ecdsa.PrivateKey, error) {
	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = S256()
	if 8*len(d) != priv.Params().BitSize {
		return nil, fmt.Errorf("invalid length, need %d bits", priv.Params().BitSize)
	}
	priv.D = new(big.Int).SetBytes(d)
	priv.PublicKey.X, priv.PublicKey.Y = priv.PublicKey.Curve.ScalarBaseMult(d)
	return priv, nil
}

// ToECDSAPrivate creates a public key with the given data value.
func ToECDSAPublic(pub []byte) *ecdsa.PublicKey {
	if len(pub) == 0 {
		return nil
	}
	x, y := elliptic.Unmarshal(S256(), pub)
	return &ecdsa.PublicKey{Curve: S256(), X: x, Y: y}
}

func Sign(hash []byte, prv *ecdsa.PrivateKey) (sig []byte, err error) {
	// TODO(larry.wang) implement it later
	return nil, nil
}

func Verify(data []byte, pub *ecdsa.PublicKey) (bool, error) {
	// TODO(larry.wang) implement it later
	return false, nil
}
