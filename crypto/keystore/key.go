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

package keystore

// Algorithm type alias
type Algorithm uint8

const (
	// SECP256K1 a type of signer
	SECP256K1 Algorithm = 1

	// SCRYPT a type of encrypt
	SCRYPT Algorithm = 1 << 4
)

// Key interface
type Key interface {

	// Algorithm returns the standard algorithm for this key. For
	// example, "ECDSA" would indicate that this key is a ECDSA key.
	Algorithm() Algorithm

	// Encoded returns the key in its primary encoding format, or null
	// if this key does not support encoding.
	Encoded() ([]byte, error)

	// Decode decode data to key
	Decode(data []byte) error

	// Clear clear key content
	Clear()
}

// PrivateKey privatekey interface
type PrivateKey interface {

	// Algorithm returns the standard algorithm for this key. For
	// example, "ECDSA" would indicate that this key is a ECDSA key.
	Algorithm() Algorithm

	// Encoded returns the key in its primary encoding format, or null
	// if this key does not support encoding.
	Encoded() ([]byte, error)

	// Decode decode data to key
	Decode(data []byte) error

	// Clear clear key content
	Clear()

	// PublicKey returns publickey
	PublicKey() PublicKey
}

// PublicKey publickey interface
type PublicKey interface {

	// Algorithm returns the standard algorithm for this key. For
	// example, "ECDSA" would indicate that this key is a ECDSA key.
	Algorithm() Algorithm

	// Encoded returns the key in its primary encoding format, or null
	// if this key does not support encoding.
	Encoded() ([]byte, error)

	// Decode decode data to key
	Decode(data []byte) error

	// Clear clear key content
	Clear()
}
