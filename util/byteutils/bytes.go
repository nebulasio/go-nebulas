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

package byteutils

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"hash/fnv"

	"github.com/btcsuite/btcutil/base58"
)

// Hash by Sha3-256
type Hash []byte

// HexHash is the hex string of a hash
type HexHash string

// Hex return hex encoded hash.
func (h Hash) Hex() HexHash {
	return HexHash(Hex(h))
}

// Base58 return base58 encodes string
func (h Hash) Base58() string {
	return base58.Encode(h)
}

// Equals compare two Hash. True is equal, otherwise false.
func (h Hash) Equals(b Hash) bool {
	return bytes.Compare(h, b) == 0
}

func (h Hash) String() string {
	return string(h.Hex())
}

// Hash return hex decoded hash.
func (hh HexHash) Hash() (Hash, error) {
	v, err := FromHex(string(hh))
	if err != nil {
		return nil, err
	}
	return Hash(v), nil
}

// Encode encodes object to Encoder.
func Encode(s interface{}, enc Encoder) ([]byte, error) {
	return enc.EncodeToBytes(s)
}

// Decode decodes []byte from Decoder.
func Decode(data []byte, dec Decoder) (interface{}, error) {
	return dec.DecodeFromBytes(data)
}

// Hex encodes []byte to Hex.
func Hex(data []byte) string {
	return hex.EncodeToString(data)
}

// FromHex decodes string from Hex.
func FromHex(data string) ([]byte, error) {
	return hex.DecodeString(data)
}

// Uint64 encodes []byte.
func Uint64(data []byte) uint64 {
	return binary.BigEndian.Uint64(data)
}

// FromUint64 decodes unit64 value.
func FromUint64(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}

// Uint32 encodes []byte.
func Uint32(data []byte) uint32 {
	return binary.BigEndian.Uint32(data)
}

// FromUint32 decodes uint32.
func FromUint32(v uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, v)
	return b
}

// Uint16 encodes []byte.
func Uint16(data []byte) uint16 {
	return binary.BigEndian.Uint16(data)
}

// FromUint16 decodes uint16.
func FromUint16(v uint16) []byte {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, v)
	return b
}

// Int64 encodes []byte.
func Int64(data []byte) int64 {
	return int64(binary.BigEndian.Uint64(data))
}

// FromInt64 decodes int64 v.
func FromInt64(v int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

// Int32 encodes []byte.
func Int32(data []byte) int32 {
	return int32(binary.BigEndian.Uint32(data))
}

// FromInt32 decodes int32 v.
func FromInt32(v int32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(v))
	return b
}

// Int16 encode []byte.
func Int16(data []byte) int16 {
	return int16(binary.BigEndian.Uint16(data))
}

// FromInt16 decodes int16 v.
func FromInt16(v int16) []byte {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, uint16(v))
	return b
}

// Equal checks whether byte slice a and b are equal.
func Equal(a []byte, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// HashBytes return bytes hash
func HashBytes(a []byte) uint32 {
	hasherA := fnv.New32a()
	hasherA.Write(a)
	return hasherA.Sum32()
}

// Less return if a < b
func Less(a []byte, b []byte) bool {
	return HashBytes(a) < HashBytes(b)
}
