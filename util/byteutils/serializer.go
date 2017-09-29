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
	"github.com/gogo/protobuf/proto"
	json "github.com/pquerna/ffjson/ffjson"
)

// JSONSerializer implements conversion between bytes and json string.
type JSONSerializer struct{}

// Serialize converts struct or array into json string.
func (s *JSONSerializer) Serialize(val interface{}) ([]byte, error) {
	return json.Marshal(val)
}

// Deserialize converts json string into struct or array.
func (s *JSONSerializer) Deserialize(val []byte, res interface{}) error {
	return json.Unmarshal(val, res)
}

// ProtoSerializer implements conversion between bytes and proto message.
type ProtoSerializer struct{}

// Serialize converts proto message to bytes.
func (s *ProtoSerializer) Serialize(val interface{}) ([]byte, error) {
	return proto.Marshal(val.(proto.Message))
}

// Deserialize converts byte into proto message.
func (s *ProtoSerializer) Deserialize(val []byte, res interface{}) error {
	return proto.Unmarshal(val, res.(proto.Message))
}
