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

package bytes

import (
	"encoding/binary"
	"encoding/hex"
)

// Encode encode object to Encoder.
func Encode(s interface{}, enc Encoder) ([]byte, error) {
	return enc.EncodeToBytes(s)
}

// Decode decode []byte from Decoder.
func Decode(data []byte, dec Decoder) (interface{}, error) {
	return dec.DecodeFromBytes(data)
}

// Hex encode []byte to Hex.
func Hex(data []byte) string {
	return hex.EncodeToString(data)
}

// FromHex decode string from Hex.
func FromHex(data string) ([]byte, error) {
	return hex.DecodeString(data)
}

// Uint64 encode []byte.
func Uint64(data []byte) uint64 {
	return binary.BigEndian.Uint64(data)
}

// FromUint64 decode v.
func FromUint64(v uint64) (b [8]byte) {
	binary.BigEndian.PutUint64(b[:8], v)
	return
}

// Uint32 encode []byte.
func Uint32(data []byte) uint32 {
	return binary.BigEndian.Uint32(data)
}

// FromUint32 decode v.
func FromUint32(v uint32) (b [4]byte) {
	binary.BigEndian.PutUint32(b[:4], v)
	return
}

// Uint16 encode []byte.
func Uint16(data []byte) uint16 {
	return binary.BigEndian.Uint16(data)
}

// FromUint16 decode v.
func FromUint16(v uint16) (b [2]byte) {
	binary.BigEndian.PutUint16(b[:2], v)
	return
}

// Int64 encode []byte.
func Int64(data []byte) int64 {
	return int64(binary.BigEndian.Uint64(data))
}

// FromInt64 decode v.
func FromInt64(v int64) (b [8]byte) {
	binary.BigEndian.PutUint64(b[:8], uint64(v))
	return
}

// Int32 encode []byte.
func Int32(data []byte) int32 {
	return int32(binary.BigEndian.Uint32(data))
}

// FromInt32 decode v.
func FromInt32(v int32) (b [4]byte) {
	binary.BigEndian.PutUint32(b[:4], uint32(v))
	return
}

// Int16 encode []byte.
func Int16(data []byte) int16 {
	return int16(binary.BigEndian.Uint16(data))
}

// FromInt16 decode v.
func FromInt16(v int16) (b [2]byte) {
	binary.BigEndian.PutUint16(b[:2], uint16(v))
	return
}
