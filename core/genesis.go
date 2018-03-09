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
	"io/ioutil"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/common/trie"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/core/state"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

// Genesis Block Hash
var (
	GenesisHash      = make([]byte, BlockHashLength)
	GenesisTimestamp = int64(0)
	GenesisCoinbase  = &Address{make([]byte, AddressLength)} // ToFix: make checksum valid
)

// LoadGenesisConf load genesis conf for file
func LoadGenesisConf(filePath string) (*corepb.Genesis, error) { // ToCheck: filepath shouldn't be nil.
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		//logging.CLog().Error("Failed to read the config file : %s. error: %s", filePath, err)
		logging.CLog().WithFields(logrus.Fields{
			"err": err,
		}).Info("Failed to read the genesis config file.")
		return nil, err
	}
	content := string(b)

	genesis := new(corepb.Genesis)
	if err := proto.UnmarshalText(content, genesis); err != nil {
		logging.CLog().Fatalf("genesis.conf parse failed. err:%v", err)
		return nil, err
	}
	return genesis, nil
}

// NewGenesisBlock create genesis @Block from file.
func NewGenesisBlock(conf *corepb.Genesis, chain *BlockChain) (*Block, error) {
	accState, err := state.NewAccountState(nil, chain.storage)
	if err != nil {
		return nil, err
	}
	txsState, err := trie.NewBatchTrie(nil, chain.storage)
	if err != nil {
		return nil, err
	}
	eventsState, err := trie.NewBatchTrie(nil, chain.storage)
	if err != nil {
		return nil, err
	}
	consensusState, err := chain.consensusHandler.GenesisState(chain, conf)
	if err != nil {
		return nil, err
	}
	genesisBlock := &Block{
		header: &BlockHeader{
			hash:       GenesisHash,
			chainID:    conf.Meta.ChainId,
			parentHash: GenesisHash,
			coinbase:   GenesisCoinbase,
			timestamp:  GenesisTimestamp,
		},
		accState:       accState,
		txsState:       txsState,
		eventsState:    eventsState,
		consensusState: consensusState,
		txPool:         chain.txPool,
		storage:        chain.storage,
		eventEmitter:   chain.eventEmitter,
		height:         1,
		sealed:         false,
	}

	genesisBlock.begin()

	for _, v := range conf.TokenDistribution {
		addr, err := AddressParse(v.Address)
		if err != nil {
			logging.CLog().WithFields(logrus.Fields{
				"address": v.Address,
				"err":     err,
			}).Error("Found invalid address in genesis token distribution.")
			genesisBlock.rollback()
			return nil, err
		}
		acc, err := genesisBlock.accState.GetOrCreateUserAccount(addr.address)
		if err != nil {
			genesisBlock.rollback()
			return nil, err
		}
		txsBalance, err := util.NewUint128FromString(v.Value)
		if err != nil {
			genesisBlock.rollback()
			return nil, err
		}
		err = acc.AddBalance(txsBalance)
		if err != nil {
			genesisBlock.rollback()
			return nil, err
		}
	}

	genesisBlock.header.stateRoot, err = genesisBlock.accState.RootHash()
	if err != nil {
		return nil, err
	}
	genesisBlock.header.txsRoot = genesisBlock.txsState.RootHash()
	genesisBlock.header.eventsRoot = genesisBlock.eventsState.RootHash()
	if genesisBlock.header.consensusRoot, err = genesisBlock.consensusState.RootHash(); err != nil {
		return nil, err
	}
	genesisBlock.sealed = true

	genesisBlock.commit()

	return genesisBlock, nil
}

// CheckGenesisBlock if a block is a genesis block
func CheckGenesisBlock(block *Block) bool {
	if block == nil {
		return false
	}
	if block.Hash().Equals(GenesisHash) {
		return true
	}
	return false
}

// DumpGenesis return the configuration of the genesis block in the storage
func DumpGenesis(chain *BlockChain) (*corepb.Genesis, error) {
	genesis, err := LoadBlockFromStorage(GenesisHash, chain)
	if err != nil {
		return nil, err
	}
	dynasty, err := genesis.consensusState.Dynasty()
	if err != nil {
		return nil, err
	}
	bootstrap := []string{}
	for _, v := range dynasty {
		bootstrap = append(bootstrap, v.String())
	}
	distribution := []*corepb.GenesisTokenDistribution{}
	accounts, err := genesis.accState.Accounts() // ToConfirm: Accounts interface is risky
	for _, v := range accounts {
		balance := v.Balance()
		if v.Address().Equals(genesis.Coinbase().Bytes()) {
			continue
		}
		distribution = append(distribution, &corepb.GenesisTokenDistribution{
			Address: string(v.Address().Hex()),
			Value:   balance.String(),
		})
	}
	return &corepb.Genesis{
		Meta: &corepb.GenesisMeta{ChainId: genesis.ChainID()},
		Consensus: &corepb.GenesisConsensus{
			Dpos: &corepb.GenesisConsensusDpos{Dynasty: bootstrap},
		},
		TokenDistribution: distribution,
	}, nil
}

func checkGenesisConfByDB(chain *BlockChain, pGenesis *corepb.Genesis) error {
	//private function [Empty parameters are checked by the caller]
	if genesis, _ := DumpGenesis(chain); genesis != nil {
		if pGenesis.Meta.ChainId != genesis.Meta.ChainId {
			return ErrGenesisNotEqualChainIDInDB
		}

		if len(pGenesis.Consensus.Dpos.Dynasty) != len(genesis.Consensus.Dpos.Dynasty) {
			return ErrGenesisNotEqualDynastyLenInDB
		}

		if len(pGenesis.TokenDistribution) != len(genesis.TokenDistribution) {
			return ErrGenesisNotEqualTokenLenInDB
		}

		// check dpos equal
		for _, confDposAddr := range pGenesis.Consensus.Dpos.Dynasty {
			contains := false
			for _, dposAddr := range genesis.Consensus.Dpos.Dynasty {
				if dposAddr == confDposAddr {
					contains = true
					break
				}
			}
			if !contains {
				return ErrGenesisNotEqualDynastyInDB
			}

		}

		// check distribution equal
		for _, confDistribution := range pGenesis.TokenDistribution {
			contains := false
			for _, distribution := range genesis.TokenDistribution {
				if distribution.Address == confDistribution.Address &&
					distribution.Value == confDistribution.Value {
					contains = true
					break
				}
			}
			if !contains {
				return ErrGenesisNotEqualTokenInDB
			}
		}
	}
	return nil
}
