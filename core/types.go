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

package core

import (
	"bytes"

	"github.com/nebulasio/go-nebulas/util/byteutils"
	log "github.com/sirupsen/logrus"
)

type Hash []byte
type HexHash string

type Consensus interface {
	VerifyBlock(*Block) error
}

// Hex return hex encoded hash.
func (h Hash) Hex() HexHash {
	return HexHash(byteutils.Hex(h))
}

// Equals compare two Hash. True is equal, otherwise false.
func (h Hash) Equals(b Hash) bool {
	return bytes.Compare(h, b) == 0
}

func (h Hash) String() string {
	return string(h.Hex())
}

// Hash return hex decoded hash.
func (hh HexHash) Hash() Hash {
	v, err := byteutils.FromHex(string(hh))
	if err != nil {
		log.Errorf("HexHash.Hash: hex decode %s failed, err is %s", hh, err)
		return nil
	}
	return Hash(v)
}
