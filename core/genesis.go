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
	"fmt"
	"io/ioutil"

	"github.com/nebulasio/go-nebulas/crypto/keystore"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/common/dag"
	consensuspb "github.com/nebulasio/go-nebulas/consensus/pb"
	corepb "github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/core/state"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

// Genesis Block Hash
var (
	GenesisHash        = make([]byte, BlockHashLength)
	GenesisTimestamp   = int64(0)
	GenesisCoinbase, _ = NewAddressFromPublicKey(make([]byte, PublicKeyDataLength))
)

// LoadGenesisConf load genesis conf for file
func LoadGenesisConf(filePath string) (*corepb.Genesis, error) {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to read the genesis config file.")
		return nil, err
	}
	content := string(b)

	genesis := new(corepb.Genesis)
	if err := proto.UnmarshalText(content, genesis); err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to parse genesis file.")
		return nil, err
	}

	return genesis, nil
}

// NewGenesisBlock create genesis @Block from file.
func NewGenesisBlock(conf *corepb.Genesis, chain *BlockChain) (*Block, error) {
	if conf == nil || chain == nil {
		return nil, ErrNilArgument
	}

	worldState, err := state.NewWorldState(chain.ConsensusHandler(), chain.storage)
	if err != nil {
		return nil, err
	}
	genesisBlock := &Block{
		header: &BlockHeader{
			hash:          GenesisHash,
			parentHash:    GenesisHash,
			chainID:       conf.Meta.ChainId,
			coinbase:      GenesisCoinbase,
			timestamp:     GenesisTimestamp,
			consensusRoot: &consensuspb.ConsensusRoot{},
			alg:           keystore.SECP256K1,
		},
		transactions: make(Transactions, 0),
		dependency:   dag.NewDag(),
		worldState:   worldState,
		txPool:       chain.txPool,
		storage:      chain.storage,
		eventEmitter: chain.eventEmitter,
		nvm:          chain.nvm,
		dip:          chain.dip,
		height:       1,
		sealed:       false,
	}

	consensusState, err := chain.ConsensusHandler().GenesisConsensusState(chain, conf)
	if err != nil {
		return nil, err
	}
	genesisBlock.worldState.SetConsensusState(consensusState)

	if err := genesisBlock.Begin(); err != nil {
		return nil, err
	}
	// add token distribution for genesis
	for _, v := range conf.TokenDistribution {
		addr, err := AddressParse(v.Address)
		if err != nil {
			logging.CLog().WithFields(logrus.Fields{
				"address": v.Address,
				"err":     err,
			}).Error("Found invalid address in genesis token distribution.")
			genesisBlock.RollBack()
			return nil, err
		}
		acc, err := genesisBlock.worldState.GetOrCreateUserAccount(addr.address)
		if err != nil {
			genesisBlock.RollBack()
			return nil, err
		}
		txsBalance, err := util.NewUint128FromString(v.Value)
		if err != nil {
			genesisBlock.RollBack()
			return nil, err
		}
		err = acc.AddBalance(txsBalance)
		if err != nil {
			genesisBlock.RollBack()
			return nil, err
		}
	}

	// genesis transaction
	declaration := fmt.Sprintf(
		"%s\n\n%s\n\n%s\n\n%s\n\n%s\n\n%s\n\n%s\n\n\n\n%s\n\n%s\n\n%s\n\n%s\n\n%s\n\n%s\n\n\n\n%s",
		"Nebulas Manifesto",
		"Yes! We Believe",
		"We believe blockchains are a foundational innovation of the new world. At its essence, this innovation is about the decentralization of data. People will be empowered to claim ownership of their data through tokens, which will enable data to be valued and exchanged by everyone on the blockchain.",
		"The blockchain community embodies the values of openness, collaboration, and transparency. An ecosystem created by blockchain believers is a voluntary association anchored by aligned incentives. We believe blockchain represents the social contract of the future, and that will lead to a civilization where cooperation, inclusion, and the interests of society converge.",
		"Blockchains will make life more free, equitable and purposeful. As a nascent digital organism and economic system, blockchain is fertile ground for creative evolution. It will give rise to transformative ideas and breakthrough technologies. Now is a time of great opportunity, challenge and hope.",
		"Do not ask what blockchain can do for you. Ask what you can do for blockchain.",
		"This is the genesis of Nebulas.",

		"星云宣言",
		"Yes! We Believe",
		"我们认为，区块链是奠基新世界的颠覆式创新，其本质是去中心化的数据确权。确权的数据承载于通证之上，对于“链上”数据的交互具有不可或缺的作用。",
		"我们同时看到，真正的区块链社区秉持着开放、共享、透明的精神，逐步建立人类历史上前所未有的大规模协作关系。一个公正并有效的价值发现、激励以及持续进化机制所构成的生态系统，是这场大规模协作关系蓬勃发展的原生推动力，也是星云对于区块链的伟大使命。",
		"我们始终坚信，区块链技术会帮助人们抵达更为自由、平等、美好的生活。区块链作为全新的生命体和经济体，意味着新的思想和技术，也蕴含着新的挑战、机遇和希望。面对无穷可能性的感召，不要问区块链能为你做什么，要问你能为区块链做什么。",
		"星云正是为此而生。",

		"by Nebulas (nebulas.io)",
	)
	declarationTx, err := NewTransaction(
		chain.ChainID(),
		GenesisCoinbase, GenesisCoinbase,
		util.Uint128Zero(), 1,
		TxPayloadBinaryType,
		[]byte(declaration),
		GenesisGasPrice,
		MinGasCountPerTransaction,
	)
	if err != nil {
		return nil, err
	}
	declarationTx.timestamp = 0
	hash, err := declarationTx.HashTransaction()
	if err != nil {
		return nil, err
	}
	declarationTx.hash = hash
	declarationTx.alg = keystore.SECP256K1
	pbTx, err := declarationTx.ToProto()
	if err != nil {
		return nil, err
	}
	txBytes, err := proto.Marshal(pbTx)
	if err != nil {
		return nil, err
	}
	genesisBlock.transactions = append(genesisBlock.transactions, declarationTx)
	if err := genesisBlock.worldState.PutTx(declarationTx.hash, txBytes); err != nil {
		return nil, err
	}

	genesisBlock.Commit()

	genesisBlock.header.stateRoot = genesisBlock.WorldState().AccountsRoot()
	genesisBlock.header.txsRoot = genesisBlock.WorldState().TxsRoot()
	genesisBlock.header.eventsRoot = genesisBlock.WorldState().EventsRoot()
	genesisBlock.header.consensusRoot = genesisBlock.WorldState().ConsensusRoot()

	genesisBlock.sealed = true

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

// CheckGenesisTransaction if a tx is a genesis transaction
func CheckGenesisTransaction(tx *Transaction) bool {
	if tx == nil {
		return false
	}
	if tx.from.Equals(GenesisCoinbase) {
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
	dynasty, err := genesis.worldState.Dynasty()
	if err != nil {
		return nil, err
	}
	bootstrap := []string{}
	for _, v := range dynasty {
		addr, err := AddressParseFromBytes(v)
		if err != nil {
			return nil, err
		}
		bootstrap = append(bootstrap, addr.String())
	}
	distribution := []*corepb.GenesisTokenDistribution{}
	accounts, err := genesis.worldState.Accounts()
	if err != nil {
		return nil, err
	}
	for _, v := range accounts {
		balance := v.Balance()
		if v.Address().Equals(genesis.Coinbase().Bytes()) {
			continue
		}
		addr, err := AddressParseFromBytes(v.Address())
		if err != nil {
			return nil, err
		}
		distribution = append(distribution, &corepb.GenesisTokenDistribution{
			Address: addr.String(),
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

//CheckGenesisConfByDB check mem and genesis.conf if equal return nil
func CheckGenesisConfByDB(pGenesisDB *corepb.Genesis, pGenesis *corepb.Genesis) error {
	//private function [Empty parameters are checked by the caller]
	if pGenesisDB != nil {
		if pGenesis.Meta.ChainId != pGenesisDB.Meta.ChainId {
			return ErrGenesisNotEqualChainIDInDB
		}

		if len(pGenesis.Consensus.Dpos.Dynasty) != len(pGenesisDB.Consensus.Dpos.Dynasty) {
			return ErrGenesisNotEqualDynastyLenInDB
		}

		if len(pGenesis.TokenDistribution) != len(pGenesisDB.TokenDistribution) {
			return ErrGenesisNotEqualTokenLenInDB
		}

		// check dpos equal
		for _, confDposAddr := range pGenesis.Consensus.Dpos.Dynasty {
			contains := false
			for _, dposAddr := range pGenesisDB.Consensus.Dpos.Dynasty {
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
			for _, distribution := range pGenesisDB.TokenDistribution {
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
