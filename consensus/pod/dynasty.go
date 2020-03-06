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
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/gogo/protobuf/proto"
	lru "github.com/hashicorp/golang-lru"
	"github.com/nebulasio/go-nebulas/common/trie"
	"github.com/nebulasio/go-nebulas/core"
	corepb "github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

// Dynasty pod dynasty
type Dynasty struct {
	chain *core.BlockChain

	genesisTimestamp int64

	//tries map[int64]*trie.Trie
	tries *lru.Cache
}

// NewDynasty create dynasty
func NewDynasty(neb core.Neblet) (*Dynasty, error) {
	tries, err := lru.New(128)
	if err != nil {
		return nil, err
	}
	dynasty := &Dynasty{
		chain: neb.BlockChain(),
		//tries: make(map[int64]*trie.Trie),
		tries: tries,
	}
	if err := dynasty.loadFromConfig(neb.Genesis(), neb.Config().Chain.Dynasty); err != nil {
		return nil, err
	}

	return dynasty, nil
}

func (d Dynasty) updateDynasty(dynasty *corepb.Dynasty) error {
	for _, v := range dynasty.Candidate {
		if len(v.Dynasty) > 0 {
			if len(v.Dynasty) != DynastySize {
				return ErrInvalidDynasty
			}
			dynastyTrie, err := DynastyTire(v.Dynasty, d.chain.Storage())
			if err != nil {
				return err
			}
			//d.tries[int64(v.Serial)] = dynastyTrie
			d.tries.Add(int64(v.Serial), dynastyTrie)

			logging.VLog().WithFields(logrus.Fields{
				"serial":  v.Serial,
				"dynasty": v.Dynasty,
			}).Info("Update dynasty.")
		}
	}
	return nil
}

// loadFromConfig ..
func (d *Dynasty) loadFromConfig(genesis *corepb.Genesis, filePath string) error {

	dynasty := &corepb.Dynasty{
		Meta: &corepb.DynastyMeta{
			ChainId: genesis.Meta.ChainId,
		},
		Candidate: []*corepb.DynastyCandidate{
			{
				Serial:  GenesisDynastySerial,
				Dynasty: genesis.Consensus.Dpos.Dynasty,
			},
		},
	}

	if len(filePath) > 0 {
		conf, err := ioutil.ReadFile(filePath)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err":      err,
				"filePath": filePath,
			}).Error("Failed to load dynasty file.")
			return ErrFailedToLoadDynasty
		}
		content := string(conf)

		dynastyConf := new(corepb.Dynasty)
		if err = proto.UnmarshalText(content, dynastyConf); err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err":      err,
				"filePath": filePath,
			}).Error("Failed to parse dynasty.conf.")
			return ErrFailedToParseDynasty
		}

		if dynastyConf.Meta.ChainId != genesis.Meta.ChainId {
			logging.VLog().WithFields(logrus.Fields{
				"GenesisChainId": genesis.Meta.ChainId,
				"DynastyChainId": dynastyConf.Meta.ChainId,
			}).Error("ChainId in dynasty.conf differs from that in genesis.conf.")
			return ErrCheckDynastyChainID
		}

		for _, v := range dynastyConf.Candidate {
			if len(v.Dynasty) != len(genesis.Consensus.Dpos.Dynasty) {
				logging.VLog().WithFields(logrus.Fields{
					"Serial":      v.Serial,
					"dynasty":     v.Dynasty,
					"DynastySize": len(v.Dynasty),
				}).Error("Miners count in dynasty.conf differs from that in genesis.conf.")
				return ErrCheckDynastyMinersCount
			}
		}
		dynasty.Candidate = append(dynasty.Candidate, dynastyConf.Candidate...)
	}
	return d.updateDynasty(dynasty)
}

func (d *Dynasty) loadFromContract(serial int64) error {
	args := fmt.Sprintf("[%d]", serial)
	result, err := d.chain.SimulateCallContract(core.NodePodContract(), core.PoDMiners, args)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"serial": serial,
			"result": result,
			"err":    err,
		}).Error("Failed to load dynasty from contract.")
		return err
	}

	data := &corepb.Dynasty{}
	if err := json.Unmarshal([]byte(result.Msg), data); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"serial": serial,
			"result": result,
		}).Error("Failed to parse dynasty from contract.")
		return err
	}

	return d.updateDynasty(data)
}

// serial dynasty serial number
func (d *Dynasty) serial(timestamp int64) int64 {
	if d.genesisTimestamp == 0 {
		if second := d.chain.GetBlockOnCanonicalChainByHeight(2); second != nil {
			d.genesisTimestamp = second.Timestamp() - BlockIntervalInMs/SecondInMs
		} else {
			return GenesisDynastySerial
		}
	}
	if timestamp < d.genesisTimestamp {
		return GenesisDynastySerial
	}
	interval := (timestamp - d.genesisTimestamp) * SecondInMs
	return interval / DynastyIntervalInMs
}

// isProposer return if the miner is propser in that dynasty
func (d *Dynasty) isProposer(now int64, miner byteutils.Hash) (bool, error) {
	tire, err := d.getDynasty(now)
	if err != nil {
		return false, err
	}

	//logging.VLog().WithFields(logrus.Fields{
	//	"timestamp":  now,
	//	"serial":     d.serial(now),
	//	"lastHeight": d.chain.TailBlock().Height(),
	//	"miner":      miner,
	//}).Debug("isProposer.")

	iter, err := tire.Iterator(nil)
	if err != nil {
		return false, err
	}
	exist, err := iter.Next()
	for exist {
		if byteutils.Equal(iter.Value(), miner) {
			return true, nil
		}
		exist, err = iter.Next()
	}
	return false, nil
}

// getDynastyTrie query dynasty trie
func (d *Dynasty) getDynasty(timestamp int64) (*trie.Trie, error) {
	// give a default dynasty trie
	dt, _ := d.tries.Get(GenesisDynastySerial)

	serial := d.serial(timestamp)
	interval := (timestamp - d.genesisTimestamp) * SecondInMs

	// after the height, miner is selected by consensus contract
	if !core.NodeUpdateAtHeight(d.chain.TailBlock().Height()) {
		tmpDynasty := int64(0)
		// eg: dynasty is: 1----3-----6, if serial={1,2}  dynasty=1, serial={3,4,5}, dynasty=3
		for _, v := range d.tries.Keys() {
			start := v.(int64)

			if start < serial+1 && start >= tmpDynasty && interval > start*DynastyIntervalInMs {
				tmpDynasty = start
				dt, _ = d.tries.Get(v)
			}
		}
	} else {
		if !d.tries.Contains(serial) {
			d.loadFromContract(serial)
		}

		temp, _ := d.tries.Get(serial)
		// if dynasty not found in contract, use last dynasty.
		if temp == nil {
			dynastyRoot, err := d.chain.TailBlock().DynastyRoot()
			if err != nil {
				return nil, err
			}
			tail, err := trie.NewTrie(dynastyRoot, d.chain.Storage(), false)
			if err != nil {
				return nil, err
			}
			tailSerial := d.serial(d.chain.TailBlock().Timestamp())
			if tailSerial == serial {
				d.tries.Add(serial, tail)
			}

			temp = tail

			logging.CLog().WithFields(logrus.Fields{
				"timestamp":  timestamp,
				"serial":     serial,
				"lastSerial": d.serial(d.chain.TailBlock().Timestamp()),
				"lastHeight": d.chain.TailBlock().Height(),
			}).Info("Use tail dynasty until the latest block dynasty is obtained.")
		}

		dt = temp
	}

	//logging.VLog().WithFields(logrus.Fields{
	//	"timestamp": timestamp,
	//	"serial":    serial,
	//	"dt":        dt,
	//}).Debug("Dynasty info.")

	dynastyTrie, err := dt.(*trie.Trie).Clone()
	if err != nil {
		return nil, err
	}

	return dynastyTrie, nil
}

func (d *Dynasty) getNodeInfo(miner *core.Address) (*core.NodeInfo, error) {
	args := fmt.Sprintf("[\"%s\"]", miner.String())
	result, err := d.chain.SimulateCallContract(core.NodePodContract(), core.PoDNodeInfo, args)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"result": result,
			"error":  err,
		}).Error("Failed to get node from contract.")
		return nil, err
	}
	node := &core.NodeInfo{}
	if err := json.Unmarshal([]byte(result.Msg), node); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"result": result,
		}).Debug("Failed to parse node from contract.")
		return nil, err
	}
	return node, nil
}

func (d *Dynasty) getParticipants() ([]*core.NodeInfo, error) {
	result, err := d.chain.SimulateCallContract(core.NodePodContract(), core.PoDParticipants, "")
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"result": result,
			"error":  err,
		}).Error("Failed to get participants from contract.")
		return nil, err
	}
	participants := []*core.NodeInfo{}
	if err := json.Unmarshal([]byte(result.Msg), &participants); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"result": result,
		}).Debug("Failed to parse Participants from contract.")
		return nil, err
	}
	return participants, nil
}

// TraverseDynasty return all members in the dynasty
func TraverseDynasty(dynasty *trie.Trie) ([]byteutils.Hash, error) {
	members := []byteutils.Hash{}
	iter, err := dynasty.Iterator(nil)
	if err != nil && err != storage.ErrKeyNotFound {
		return nil, err
	}
	if err != nil {
		return members, nil
	}
	exist, err := iter.Next()
	for exist {
		members = append(members, iter.Value())
		exist, err = iter.Next()
	}
	return members, nil
}

// DynastyTire return dynasty tire
func DynastyTire(dynasty []string, storage storage.Storage) (*trie.Trie, error) {
	dynastyTrie, err := trie.NewTrie(nil, storage, false)
	if err != nil {
		return nil, err
	}
	for _, miner := range dynasty {
		addr, err := core.AddressParse(miner)
		if err != nil {
			return nil, err
		}
		v := addr.Bytes()
		if _, err = dynastyTrie.Put(v, v); err != nil {
			return nil, err
		}
	}
	return dynastyTrie, nil
}
