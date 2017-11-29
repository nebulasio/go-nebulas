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

// use btcec https://godoc.org/github.com/btcsuite/btcd/btcec#example-package--VerifySignature

package secp256k1

import (
	"crypto/ecdsa"
	"crypto/elliptic"

	"encoding/hex"
	"errors"
	"math/big"

	"crypto/rand"

	"github.com/nebulasio/go-nebulas/crypto/keystore/secp256k1/bitelliptic"
)

// S256 returns an instance of the secp256k1 curve.
func S256() elliptic.Curve {
	return bitelliptic.S256()
}

// NewECDSAPrivateKey generate a ecdsa private key
func NewECDSAPrivateKey() *ecdsa.PrivateKey {
	var priv *ecdsa.PrivateKey
	for {
		priv, _ = ecdsa.GenerateKey(S256(), rand.Reader)
		if SeckeyVerify(priv) {
			break
		}
	}
	return priv
}

// FromECDSAPrivateKey exports a private key into a binary dump.
func FromECDSAPrivateKey(priv *ecdsa.PrivateKey) ([]byte, error) {
	if priv == nil {
		return nil, errors.New("ecdsa: please input private key")
	}
	// as private key len cannot guarantee greater than Params bytes len, padding big bytes.
	return paddedBigBytes(priv.D, priv.Params().BitSize/8), nil
}

// FromECDSAPublicKey exports a public key into a binary dump.
func FromECDSAPublicKey(pub *ecdsa.PublicKey) ([]byte, error) {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil, errors.New("ecdsa: please input public key")
	}
	return elliptic.Marshal(S256(), pub.X, pub.Y), nil
}

// HexToECDSAPrivateKey parses a secp256k1 private key.
func HexToECDSAPrivateKey(hexkey string) (*ecdsa.PrivateKey, error) {
	b, err := hex.DecodeString(hexkey)
	if err != nil {
		return nil, err
	}
	return ToECDSAPrivateKey(b)
}

// ToECDSAPrivateKey creates a private key with the given data value.
func ToECDSAPrivateKey(d []byte) (*ecdsa.PrivateKey, error) {
	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = S256()
	priv.D = new(big.Int).SetBytes(d)
	priv.PublicKey.X, priv.PublicKey.Y = priv.PublicKey.Curve.ScalarBaseMult(d)
	return priv, nil
}

// ToECDSAPublicKey creates a public key with the given data value.
func ToECDSAPublicKey(pub []byte) (*ecdsa.PublicKey, error) {
	if len(pub) == 0 {
		return nil, errors.New("ecdsa: please input public key bytes")
	}
	x, y := elliptic.Unmarshal(S256(), pub)
	return &ecdsa.PublicKey{Curve: S256(), X: x, Y: y}, nil
}

// zeroKey zeroes the private key
func zeroKey(k *ecdsa.PrivateKey) {
	b := k.D.Bits()
	for i := range b {
		b[i] = 0
	}
}

// ZeroBytes clears byte slice.
func ZeroBytes(bytes []byte) {
	for i := range bytes {
		bytes[i] = 0
	}
}

// paddedBigBytes encodes a big integer as a big-endian byte slice.
func paddedBigBytes(bigint *big.Int, n int) []byte {
	if bigint.BitLen()/8 >= n {
		return bigint.Bytes()
	}
	ret := make([]byte, n)
	readBits(bigint, ret)
	return ret
}

const (
	// number of bits in a big.Word
	wordBits = 32 << (uint64(^big.Word(0)) >> 63)
	// number of bytes in a big.Word
	wordBytes = wordBits / 8
)

// readBits encodes the absolute value of bigint as big-endian bytes.
func readBits(bigint *big.Int, buf []byte) {
	i := len(buf)
	for _, d := range bigint.Bits() {
		for j := 0; j < wordBytes && i > 0; j++ {
			i--
			buf[i] = byte(d)
			d >>= 8
		}
	}
}
