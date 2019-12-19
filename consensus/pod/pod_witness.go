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
	mapset "github.com/deckarep/golang-set"
	"github.com/gogo/protobuf/proto"
	consensuspb "github.com/nebulasio/go-nebulas/consensus/pb"
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/net"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

func (pod *PoD) broadcastWitness(hashs []byteutils.Hash) error {
	witness := &Witness{
		witness:    pod.miner.Bytes(),
		blockHashs: hashs,
	}
	if err := pod.signWitness(witness); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"miner": pod.miner,
			"hash":  hashs,
			"err":   err,
		}).Error("Failed to sign witness.")
		return err
	}
	pod.ns.Broadcast(MessageTypeWitness, witness, net.MessagePriorityNormal)
	//logging.VLog().WithFields(logrus.Fields{
	//	"miner": pod.miner,
	//	"hash":  hashs,
	//}).Debug("Broadcast witness to peers.")
	return nil
}

// signWitness sign witness
func (pod *PoD) signWitness(witness *Witness) (err error) {
	hash := witness.Hash()
	alg := keystore.SECP256K1
	var sign byteutils.Hash
	if pod.enableRemoteSignServer {
		sign, err = pod.remoteSign(alg, hash)
		if err != nil {
			return err
		}
	} else {
		sign, err = pod.am.SignHash(pod.miner, hash, alg)
		if err != nil {
			return err
		}
	}
	witness.alg = alg
	witness.sign = sign
	return nil
}

func (pod *PoD) onWitnessReceived(msg net.Message) error {
	witness := new(Witness)
	pbWitness := new(consensuspb.Witness)
	if err := proto.Unmarshal(msg.Data(), pbWitness); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"msgType": msg.MessageType(),
			"msg":     msg,
			"err":     err,
		}).Debug("Failed to unmarshal data.")
		return err
	}
	if err := witness.FromProto(pbWitness); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"msgType": msg.MessageType(),
			"msg":     msg,
			"err":     err,
		}).Debug("Failed to recover a witness from proto data.")
		return err
	}

	if err := verifyWitnessSign(witness); err != nil {
		return err
	}

	for _, v := range witness.blockHashs {
		block := pod.chain.GetBlock(v)
		if block != nil {
			found, err := pod.dynasty.isProposer(block.Timestamp(), witness.witness)
			//logging.VLog().WithFields(logrus.Fields{
			//	"msgType": msg.MessageType(),
			//	"found": found,
			//	"err":   err,
			//}).Debug("Check witness proposer in dynasty.")
			// if witness is not the miner in block's dynasty, don't relay the message.
			if !found || err != nil {
				return err
			}

			// find local reversible blocks and update to the lib
			reversibleSet, _ := pod.reversible.Get(v.Hex())
			if reversibleSet == nil {
				reversibleSet = mapset.NewSet()
			}
			reversibleSet.(mapset.Set).Add(witness.witness.Hex())
			if reversibleSet.(mapset.Set).Cardinality() >= ConsensusSize {
				logging.VLog().WithFields(logrus.Fields{
					"hash": v.Hex(),
					"set":  reversibleSet,
					"len":  reversibleSet.(mapset.Set).Cardinality(),
				}).Debug("Update lib by bft witness.")
				pod.setLib(block, reversibleSet.(mapset.Set).Cardinality())
			} else {
				pod.reversible.Add(block.Hash().Hex(), reversibleSet)
			}
		}
	}

	pod.ns.Relay(MessageTypeWitness, witness, net.MessagePriorityNormal)

	//logging.VLog().WithFields(logrus.Fields{
	//	"msgType": msg.MessageType(),
	//	"witness": witness,
	//}).Debug("Receive witness from peers.")
	return nil
}

func verifyWitnessSign(witness *Witness) error {
	signer, err := core.RecoverSignerFromSignature(witness.alg, witness.Hash(), witness.sign)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"signer": signer,
			"err":    err,
			"block":  witness,
		}).Debug("Failed to recover witness.")
		return err
	}

	from, err := core.AddressParseFromBytes(witness.witness)
	if err != nil {
		return err
	}
	if !from.Equals(signer) {
		logging.VLog().WithFields(logrus.Fields{
			"signer":  signer,
			"from":    from,
			"witness": witness,
		}).Debug("Failed to verify witness's sign.")
		return ErrInvalidWitnessSign
	}
	return nil
}
