// Copyright (C) 2017-2019 go-nebulas authors
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

package pod

import (
	"github.com/gogo/protobuf/proto"
	consensuspb "github.com/nebulasio/go-nebulas/consensus/pb"
	"github.com/nebulasio/go-nebulas/crypto"
	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/crypto/sha3"
	"github.com/nebulasio/go-nebulas/util/byteutils"
)

// Witness obj
type Witness struct {
	witness    byteutils.Hash
	blockHashs []byteutils.Hash

	// sign
	alg  keystore.Algorithm
	sign byteutils.Hash
}

func (w *Witness) Hash() byteutils.Hash {
	hasher := sha3.New256()
	for _, v := range w.blockHashs {
		hasher.Write(v)
	}
	return hasher.Sum(nil)
}

// ToProto converts domain BlockHeader to proto BlockHeader
func (w *Witness) ToProto() (proto.Message, error) {
	hashs := make([][]byte, len(w.blockHashs))
	for k, v := range w.blockHashs {
		hashs[k] = v
	}
	return &consensuspb.Witness{
		Witness: w.witness,
		Hash:    hashs,
		Alg:     uint32(w.alg),
		Sign:    w.sign,
	}, nil
}

// FromProto converts proto BlockHeader to domain BlockHeader
func (w *Witness) FromProto(msg proto.Message) error {
	if msg, ok := msg.(*consensuspb.Witness); ok {
		if msg != nil {
			w.witness = msg.Witness
			hashs := make([]byteutils.Hash, len(msg.Hash))
			for k, v := range msg.Hash {
				hashs[k] = v
			}
			w.blockHashs = hashs

			alg := keystore.Algorithm(msg.Alg)
			if err := crypto.CheckAlgorithm(alg); err != nil {
				return err
			}
			w.alg = alg
			w.sign = msg.Sign
			return nil
		}
		return ErrInvalidProtoToWitness
	}
	return ErrInvalidProtoToWitness
}
