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
)

func TestNewGenesisBlock(t *testing.T) {
	neb := testNeb(t)
	chain := neb.chain
	genesis := neb.chain.genesisBlock
	conf := MockGenesisConf()

	for _, v := range conf.TokenDistribution {
		addr, _ := AddressParse(v.Address)
		acc, err := genesis.worldState.GetOrCreateUserAccount(addr.Bytes())
		assert.Nil(t, err)
		assert.Equal(t, acc.Balance().String(), v.Value)
	}

	dumpConf, err := DumpGenesis(chain)
	assert.Nil(t, err)
	assert.Equal(t, dumpConf.Meta.ChainId, conf.Meta.ChainId)
	assert.Equal(t, dumpConf.TokenDistribution, conf.TokenDistribution)
}

func TestInvalidAddressInTokenDistribution(t *testing.T) {
	mockConf := MockGenesisConf()
	mockConf.TokenDistribution[0].Address = "n1UZtMgi94oE913L2Sa2C9XwvAzNTQ82v64121"
	chain := testNeb(t).chain
	_, err := NewGenesisBlock(mockConf, chain)
	assert.Equal(t, err, ErrInvalidAddressFormat)
}
