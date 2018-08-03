// Copyright (C) 2018 go-nebulas authors
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
	"io/ioutil"
	"os"
	"path"
	"sync"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/common/trie"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

var (
	// DynastyConf define dynasty addresses
	DynastyConf *corepb.Dynasty
	DynastyTrie *trie.Trie
	once        sync.Once
)

func loadDynastyConf(genesisConfPath string, genesis *corepb.Genesis) {
	dir := path.Dir(genesisConfPath)
	fp := path.Join(dir, "dynasty.conf")
	b, err := ioutil.ReadFile(fp)
	if os.IsNotExist(err) {
		logging.VLog().WithFields(logrus.Fields{
			"filepath": fp,
		}).Fatal("File doesn't exist.")
		return
	}

	DynastyConf = new(corepb.Dynasty)
	if err = proto.Unmarshal(b, DynastyConf); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err":      err,
			"filepath": fp,
		}).Fatal("Failed to parse dynasty file.")
	}

	// only 1 candidate allowed
	if genesis.Meta.ChainId != genesis.Meta.ChainId || len(DynastyConf.Candidate) != 1 {
		logging.VLog().WithFields(logrus.Fields{
			"GenesisChainId":      genesis.Meta.ChainId,
			"DynastyChainId":      DynastyConf.Meta.ChainId,
			"DynastyCandidateLen": len(DynastyConf.Candidate),
		}).Fatal("Dynasty conf is invalid.")
	}

	if len(DynastyConf.Candidate[0].Dynasty) != len(genesis.Consensus.Dpos.Dynasty) {
		logging.VLog().WithFields(logrus.Fields{
			"DynastySize": len(DynastyConf.Candidate[0].Dynasty),
		}).Fatal("Dynasty conf is invalid.")
	}
}

// InitDynastyFromConf ...
func InitDynastyFromConf(chain *BlockChain) {
	once.Do(func() {
		DynastyTrie, err := trie.NewTrie(nil, chain.Storage(), false)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err": err,
			}).Fatal("Failed to new trie.")
		}

		candidate := DynastyConf.Candidate[0]
		for i := 0; i < len(candidate.Dynasty); i++ {
			addr := candidate.Dynasty[i]
			member, err := AddressParse(addr)
			if err != nil {
				logging.VLog().WithFields(logrus.Fields{
					"err": err,
				}).Fatal("Failed to parse address.")
			}
			v := member.Bytes()
			if _, err = DynastyTrie.Put(v, v); err != nil {
				logging.VLog().WithFields(logrus.Fields{
					"err": err,
				}).Fatal("Failed to put value.")
			}
		}

		logging.VLog().WithFields(logrus.Fields{
			"chainId": DynastyConf.Meta.ChainId,
			"serial":  DynastyConf.Candidate[0].Serial,
		}).Debug("Init dynasty from conf done.")
	})

}
