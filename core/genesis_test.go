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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util/byteutils"
)

var (
	MockDynasty = []string{
		"1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c",
		"2fe3f9f51f9a05dd5f7c5329127f7c917917149b4e16b0b8",
		"333cb3ed8c417971845382ede3cf67a0a96270c05fe2f700",
		"48f981ed38910f1232c1bab124f650c482a57271632db9e3",
		"59fc526072b09af8a8ca9732dae17132c4e9127e43cf2232",
		"75e4e5a71d647298b88928d8cb5da43d90ab1a6c52d0905f",
		"7da9dabedb4c6e121146fb4250a9883d6180570e63d6b080",
		"98a3eed687640b75ec55bf5c9e284371bdcaeab943524d51",
		"a8f1f53952c535c6600c77cf92b65e0c9b64496a8a328569",
		"b040353ec0f2c113d5639444f7253681aecda1f8b91f179f",
		"b414432e15f21237013017fa6ee90fc99433dec82c1c8370",
		"b49f30d0e5c9c88cade54cd1adecf6bc2c7e0e5af646d903",
		"b7d83b44a3719720ec54cdb9f54c0202de68f1ebcb927b4f",
		"ba56cc452e450551b7b9cffe25084a069e8c1e94412aad22",
		"c5bcfcb3fa8250be4f2bf2b1e70e1da500c668377ba8cd4a",
		"c79d9667c71bb09d6ca7c3ed12bfe5e7be24e2ffe13a833d",
		"d1abde197e97398864ba74511f02832726edad596775420a",
		"d86f99d97a394fa7a623fdf84fdc7446b99c3cb335fca4bf",
		"e0f78b011e639ce6d8b76f97712118f3fe4a12dd954eba49",
		"f38db3b6c801dddd624d6ddc2088aa64b5a24936619e4848",
		"fc751b484bd5296f8d267a8537d33f25a848f7f7af8cfcf6",
	}
)

// MockGenesisConf return mock genesis conf
func MockGenesisConf() *corepb.Genesis {
	return &corepb.Genesis{
		Meta: &corepb.GenesisMeta{ChainId: 100},
		Consensus: &corepb.GenesisConsensus{
			Dpos: &corepb.GenesisConsensusDpos{
				Dynasty: MockDynasty,
			},
		},
		TokenDistribution: []*corepb.GenesisTokenDistribution{
			&corepb.GenesisTokenDistribution{
				Address: "1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c",
				Value:   "10000000000000000000000",
			},
			&corepb.GenesisTokenDistribution{
				Address: "2fe3f9f51f9a05dd5f7c5329127f7c917917149b4e16b0b8",
				Value:   "10000000000000000000000",
			},
		},
	}
}

func TestInvalidAddressInTokenDistribution(t *testing.T) {
	mockConf := MockGenesisConf()
	mockConf.TokenDistribution[0].Address = "1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2"
	storage, err := storage.NewMemoryStorage()
	assert.Nil(t, err)
	_, err = NewGenesisBlock(mockConf, &BlockChain{storage: storage})
	assert.Equal(t, err, ErrInvalidAddress)
}

func TestNewGenesisBlock(t *testing.T) {
	conf := MockGenesisConf()
	storage, err := storage.NewMemoryStorage()
	assert.Nil(t, err)
	chain := &BlockChain{storage: storage}
	genesis, err := NewGenesisBlock(conf, chain)
	assert.Nil(t, chain.storeBlockToStorage(genesis))
	assert.Nil(t, err)

	iter, err := genesis.dposContext.dynastyTrie.Iterator(nil)
	assert.Nil(t, err)
	exist, err := iter.Next()
	i := 0
	for exist {
		var addr byteutils.Hash
		addr, _ = byteutils.FromHex(MockDynasty[i])
		assert.Equal(t, addr, byteutils.Hash(iter.Value()))
		i++
		exist, err = iter.Next()
	}

	iter, err = genesis.dposContext.voteTrie.Iterator(nil)
	assert.Nil(t, err)
	exist, err = iter.Next()
	i = 0
	for exist {
		var addr byteutils.Hash
		addr, _ = byteutils.FromHex(MockDynasty[i])
		assert.Equal(t, addr, byteutils.Hash(iter.Value()))
		i++
		exist, err = iter.Next()
	}

	iter, err = genesis.dposContext.delegateTrie.Iterator(nil)
	assert.Nil(t, err)
	exist, err = iter.Next()
	i = 0
	for exist {
		var addr byteutils.Hash
		addr, _ = byteutils.FromHex(MockDynasty[i])
		assert.Equal(t, addr, byteutils.Hash(iter.Value()))
		i++
		exist, err = iter.Next()
	}

	for _, v := range conf.TokenDistribution {
		addr, _ := byteutils.FromHex(v.Address)
		acc, err := genesis.accState.GetOrCreateUserAccount(addr)
		assert.Nil(t, err)
		assert.Equal(t, acc.Balance().String(), v.Value)
	}

	dumpConf, err := DumpGenesis(storage)
	assert.Nil(t, err)
	assert.Equal(t, dumpConf.Meta.ChainId, conf.Meta.ChainId)
	assert.Equal(t, dumpConf.Consensus.Dpos.Dynasty, conf.Consensus.Dpos.Dynasty)
	assert.Equal(t, dumpConf.TokenDistribution, conf.TokenDistribution)
}
