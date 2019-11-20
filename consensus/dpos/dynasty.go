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

package dpos

import (
	"io/ioutil"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/common/trie"
	"github.com/nebulasio/go-nebulas/core"
	corepb "github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

// Dynasty dpos dynasty
type Dynasty struct {
	chain *core.BlockChain

	genesisTimestamp int64

	dynasty *corepb.Dynasty
	tries   map[int64]*trie.Trie
}

// NewDynasty create dynasty
func NewDynasty(neb core.Neblet) (*Dynasty, error) {
	dynasty, err := loadDynastyConf(neb.Genesis(), neb.Config().Chain.Dynasty)
	if err != nil {
		return nil, err
	}

	tries := make(map[int64]*trie.Trie)
	for _, v := range dynasty.Candidate {
		dynastyTrie, err := trie.NewTrie(nil, neb.Storage(), false)
		if err != nil {
			return nil, err
		}
		for _, miner := range v.Dynasty {
			addr, err := core.AddressParse(miner)
			if err != nil {
				return nil, err
			}
			v := addr.Bytes()
			if _, err = dynastyTrie.Put(v, v); err != nil {
				return nil, err
			}
		}
		tries[int64(v.Serial)] = dynastyTrie
	}

	return &Dynasty{
		chain:   neb.BlockChain(),
		dynasty: dynasty,
		tries:   tries,
	}, nil
}

// loadDynastyConf ..
func loadDynastyConf(genesis *corepb.Genesis, filePath string) (*corepb.Dynasty, error) {

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
			return nil, ErrFailedToLoadDynasty
		}
		content := string(conf)

		dynastyConf := new(corepb.Dynasty)
		if err = proto.UnmarshalText(content, dynastyConf); err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err":      err,
				"filePath": filePath,
			}).Error("Failed to parse dynasty.conf.")
			return nil, ErrFailedToParseDynasty
		}

		if dynastyConf.Meta.ChainId != genesis.Meta.ChainId {
			logging.VLog().WithFields(logrus.Fields{
				"GenesisChainId": genesis.Meta.ChainId,
				"DynastyChainId": dynastyConf.Meta.ChainId,
			}).Error("ChainId in dynasty.conf differs from that in genesis.conf.")
			return nil, ErrCheckDynastyChainID
		}

		for _, v := range dynastyConf.Candidate {
			if len(v.Dynasty) != len(genesis.Consensus.Dpos.Dynasty) {
				logging.VLog().WithFields(logrus.Fields{
					"Serial":      v.Serial,
					"dynasty":     v.Dynasty,
					"DynastySize": len(v.Dynasty),
				}).Error("Miners count in dynasty.conf differs from that in genesis.conf.")
				return nil, ErrCheckDynastyMinersCount
			}
		}
		dynasty.Candidate = append(dynasty.Candidate, dynastyConf.Candidate...)
	}
	return dynasty, nil
}

// getDynastyTrie query dynasty trie
func (d *Dynasty) getDynasty(timestamp int64) (*trie.Trie, error) {

	var (
		interval   int64
		curDynasty int64
		tmpDynasty int64
		dt         *trie.Trie
	)

	if d.genesisTimestamp == 0 {
		curDynasty = GenesisDynasty
		secondBlock := d.chain.GetBlockOnCanonicalChainByHeight(2)
		if secondBlock != nil {
			d.genesisTimestamp = secondBlock.Timestamp() - BlockIntervalInMs/SecondInMs
		} else {
			interval = BlockIntervalInMs
		}
	}
	if d.genesisTimestamp > 0 {
		// for the real genesis block, the timestamp is 0.
		if timestamp < d.genesisTimestamp {
			timestamp += d.genesisTimestamp
		}
		interval = (timestamp - d.genesisTimestamp) * SecondInMs
		curDynasty = interval/DynastyIntervalInMs + 1
	}

	tmpDynasty = 0
	// give a default dynasty trie
	dt = d.tries[GenesisDynastySerial]

	// eg: dynasty is: 1----3-----6, if serial={1,2}  dynasty=1, serial={3,4,5}, dynasty=3
	for k, v := range d.tries {
		start := int64(k)

		if start < curDynasty && start >= tmpDynasty && interval > start*DynastyIntervalInMs {
			tmpDynasty = start
			dt = v
		}
	}

	if dt == nil {
		logging.CLog().WithFields(logrus.Fields{
			"genesis":    d.genesisTimestamp,
			"interval":   interval,
			"timestamp":  timestamp,
			"tmpDynasty": tmpDynasty,
			"curDynasty": curDynasty,
		}).Fatal("Failed to get dynasty with current genesis and dynasty.")
	}

	logging.VLog().WithFields(logrus.Fields{
		"interval":   interval,
		"tmpDynasty": tmpDynasty,
		"timestamp":  timestamp,
		"curDynasty": curDynasty,
		"dt":         dt,
	}).Debug("dynasty info.")

	dynastyTrie, err := dt.Clone()
	if err != nil {
		return nil, err
	}

	return dynastyTrie, nil
}
