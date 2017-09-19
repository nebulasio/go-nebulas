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

	"crypto/rand"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"github.com/nebulasio/go-nebulas/utils/byteutils"
	"math/big"
	"sync"
)

var (
	once        sync.Once
	curveParams *elliptic.CurveParams
)

// S256 returns an instance of the secp256k1 curve.
func Curve() elliptic.Curve {
	once.Do(func() {
		curveParams := new(elliptic.CurveParams)
		curveParams.Name = "nebulas"
		curveParams.P, _ = new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEFFFFFC2F", 16)
		curveParams.N, _ = new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141", 16)
		curveParams.B, _ = new(big.Int).SetString("0000000000000000000000000000000000000000000000000000000000000007", 16)
		curveParams.Gx, _ = new(big.Int).SetString("79BE667EF9DCBBAC55A06295CE870B07029BFCDB2DCE28D959F2815B16F81798", 16)
		curveParams.Gy, _ = new(big.Int).SetString("483ADA7726A3C4655DA4FBFC0E1108A8FD17B448A68554199C47D08FFB10D4B8", 16)
		curveParams.BitSize = 256
	})
	return curveParams
}

// generate a ecdsa private key
func GenerateECDSAPrivateKey(rand io.Reader) (*ecdsa.PrivateKey, error) {
	privateKeyECDSA, err := ecdsa.GenerateKey(Curve(), rand)
	if err != nil {
		return nil, err
	}
	return privateKeyECDSA, nil
}

// FromECDSAPri exports a private key into a binary dump.
func FromECDSAPri(pri *ecdsa.PrivateKey) ([]byte, error) {
	if pri == nil {
		return nil, errors.New("ecdsa: please input private key")
	}
	return x509.MarshalECPrivateKey(pri)
}

// FromECDSAPub exports a public key into a binary dump.
func FromECDSAPub(pub *ecdsa.PublicKey) ([]byte, error) {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil, errors.New("ecdsa: please input public key")
	}
	return x509.MarshalPKIXPublicKey(pub)
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
	return x509.ParseECPrivateKey(d)
}

// ToECDSAPrivate creates a public key with the given data value.
func ToECDSAPublic(pub []byte) (*ecdsa.PublicKey, error) {
	if len(pub) == 0 {
		return nil, errors.New("ecdsa: please input public key bytes")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(pub)
	return pubInterface.(*ecdsa.PublicKey), err
}

func Sign(hash []byte, prv *ecdsa.PrivateKey) (sig []byte, err error) {
	r, s, err := ecdsa.Sign(rand.Reader, prv, hash)
	if err != nil {
		return nil, errors.New("ecdsa: sign err")
	}
	sign := r.Bytes()
	sign = append(sign, s.Bytes()...)
	return sign, nil
}

func Verify(hash []byte, rs []byte, pub *ecdsa.PublicKey) bool {
	r := big.NewInt(byteutils.Int64(rs[:32]))
	s := big.NewInt(byteutils.Int64(rs[32:]))
	return ecdsa.Verify(pub, hash, r, s)
}
