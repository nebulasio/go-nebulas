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

package pbopt

import "fmt"

type CustomizedProtobufMessage interface {
	Reset()
	String() string
	ProtoMessage()
}

// Marshaler is the interface representing objects that can marshal themselves.
type Marshaler interface {
	Marshal() ([]byte, error)
}

// Unmarshaler is the interface representing objects that can
// unmarshal themselves.  The argument points to data that may be
// overwritten, so implementations should not keep references to the
// buffer.
// Unmarshal implementations should not clear the receiver.
// Any unmarshaled data should be merged into the receiver.
// Callers of Unmarshal that do not want to retain existing data
// should Reset the receiver before calling Unmarshal.
type Unmarshaler interface {
	Unmarshal([]byte) error
}

var (
	ErrMarshaler   = fmt.Errorf("customized protobuf: can not marshal")
	ErrUnmarshaler = fmt.Errorf("customized protobuf: can not unmarshal")
)

// simplify func Marshal from protobuf
func Marshal(pb CustomizedProtobufMessage) ([]byte, error) {
	if m, ok := pb.(Marshaler); ok {
		return m.Marshal()
	}
	return nil, ErrMarshaler
}

// simplify func Unmarshal from protobuf
func Unmarshal(buf []byte, pb CustomizedProtobufMessage) error {
	pb.Reset()
	if u, ok := pb.(Unmarshaler); ok {
		return u.Unmarshal(buf)
	}
	return ErrUnmarshaler
}
