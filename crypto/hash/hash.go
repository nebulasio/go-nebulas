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

package hash

import (
	"crypto/sha256"
	"encoding/base64"

	keccak "github.com/nebulasio/go-nebulas/crypto/sha3"
	"golang.org/x/crypto/ripemd160"
	"golang.org/x/crypto/sha3"
)

// const alphabet = "./ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

var bcEncoding = base64.NewEncoding(alphabet)

// Sha256 returns the SHA-256 digest of the data.
func Sha256(args ...[]byte) []byte {
	hasher := sha256.New()
	for _, bytes := range args {
		hasher.Write(bytes)
	}
	return hasher.Sum(nil)
}

// Sha3256 returns the SHA3-256 digest of the data.
func Sha3256(args ...[]byte) []byte {
	hasher := sha3.New256()
	for _, bytes := range args {
		hasher.Write(bytes)
	}
	return hasher.Sum(nil)
}

// Keccak256 returns the Keccak-256 digest of the data.
func Keccak256(data ...[]byte) []byte {
	d := keccak.NewKeccak256()
	for _, b := range data {
		d.Write(b)
	}
	return d.Sum(nil)
}

// Ripemd160 return the RIPEMD160 digest of the data.
func Ripemd160(args ...[]byte) []byte {
	hasher := ripemd160.New()
	for _, bytes := range args {
		hasher.Write(bytes)
	}
	return hasher.Sum(nil)
}

// Base64Encode encode to base64
func Base64Encode(src []byte) []byte {
	n := bcEncoding.EncodedLen(len(src))
	dst := make([]byte, n)
	bcEncoding.Encode(dst, src)
	// for dst[n-1] == '=' {
	// 	n--
	// }
	return dst[:n]
}

// Base64Decode decode base64
func Base64Decode(src []byte) ([]byte, error) {
	numOfEquals := 4 - (len(src) % 4)
	for i := 0; i < numOfEquals; i++ {
		src = append(src, '=')
	}

	dst := make([]byte, bcEncoding.DecodedLen(len(src)))
	n, err := bcEncoding.Decode(dst, src)
	if err != nil {
		return nil, err
	}
	return dst[:n], nil
}
