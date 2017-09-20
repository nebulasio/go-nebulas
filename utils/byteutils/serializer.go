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
	"compress/gzip"
	"io/ioutil"

	"github.com/nebulasio/go-nebulas/common/rlp"
	json "github.com/pquerna/ffjson/ffjson"
	"github.com/golang/protobuf/proto"
)

// RLPSerializer implements ethereum rlp algorithm
// not support map
// not support int, use uint instead
type RLPSerializer struct{}

// Serialize convert struct or array into rlp encoding bytes
func (s *RLPSerializer) Serialize(val interface{}) ([]byte, error) {
	return rlp.EncodeToBytes(val)
}

// Deserialize convert rlp encoding bytes into struct or array
func (s *RLPSerializer) Deserialize(val []byte, res interface{}) error {
	return rlp.DecodeBytes(val, res)
}

// JSONSerializer implements conversion between bytes and json string
type JSONSerializer struct{}

// Serialize convert struct or array into json string
func (s *JSONSerializer) Serialize(val interface{}) ([]byte, error) {
	return json.Marshal(val)
}

// Deserialize convert json string into struct or array
func (s *JSONSerializer) Deserialize(val []byte, res interface{}) error {
	return json.Unmarshal(val, res)
}

// ProtoSerializer implements conversion between bytes and proto message.
type ProtoSerializer struct{}

// Serialize converts proto message to bytes.
func (s *ProtoSerializer) Serialize(val proto.Message) ([]byte, error) {
	return proto.Marshal(val)
}

// Deserialize converts byte into proto message.
func (s *ProtoSerializer) Deserialize(val []byte, res proto.Message) error {
	return proto.Unmarshal(val, res)
}

func compress(val []byte) ([]byte, error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write(val); err != nil {
		return nil, err
	}
	if err := gz.Flush(); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func uncompress(val []byte) ([]byte, error) {
	source := bytes.NewReader(val)
	reader, err := gzip.NewReader(source)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(reader)
}
