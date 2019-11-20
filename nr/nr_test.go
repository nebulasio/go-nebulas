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

package nr

import (
	"encoding/json"
	"testing"

	"github.com/nebulasio/go-nebulas/account"
	"github.com/nebulasio/go-nebulas/core"
	nebletpb "github.com/nebulasio/go-nebulas/neblet/pb"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/stretchr/testify/assert"
)

type mockNeb struct {
	config *nebletpb.Config
	chain  *core.BlockChain
	am     core.AccountManager
}

func (n *mockNeb) Config() *nebletpb.Config {
	return n.config
}

func (n *mockNeb) AccountManager() core.AccountManager {
	return n.am
}

func (n *mockNeb) BlockChain() *core.BlockChain {
	return n.chain
}

func testNeb(t *testing.T) *mockNeb {

	neb := &mockNeb{
		config: &nebletpb.Config{Chain: &nebletpb.ChainConfig{ChainId: 1},
			Nbre: &nebletpb.NbreConfig{},
		},
	}

	account, _ := account.NewManager(neb)
	//assert.Nil(t, err)
	neb.am = account
	return neb
}

func readNbreDB(key string) ([]byte, error) {
	rs, err := storage.NewRocksStorage("../mainnet/nbre/nbre.db")
	if err != nil {
		return nil, err
	}
	data, err := rs.Get([]byte(key))
	if err != nil {
		return nil, err
	}
	return data, nil
}

func TestNR_ReadSum(t *testing.T) {
	dbdata, err := readNbreDB("nr_sums")
	assert.Nil(t, err)
	data := NRSums{}
	err = json.Unmarshal(dbdata, &data)
	assert.Nil(t, err)
	items := make([]*NRSummary, len(data.NrSums))
	for i, str := range data.NrSums {
		item := NRSummary{}
		err := json.Unmarshal([]byte(str), &item)
		assert.Nil(t, err)
		items[i] = &item
	}
	assert.Nil(t, err)
	//recordData, err := json.Marshal(items)
	//assert.Nil(t, err)
	//util.FileWrite("./nrdata/sum_data.json", recordData, true)
}

func TestNR_ReadNR(t *testing.T) {
	dbdata, err := readNbreDB("nr_results")
	assert.Nil(t, err)
	data := NRResults{}
	err = json.Unmarshal(dbdata, &data)
	assert.Nil(t, err)
	items := make([]*NRData, len(data.NrResults))
	for i, str := range data.NrResults {
		item := NRData{}
		err := json.Unmarshal([]byte(str), &item)
		assert.Nil(t, err)
		items[i] = &item
	}
	assert.Nil(t, err)
	//recordData, err := json.Marshal(items)
	//assert.Nil(t, err)
	//util.FileWrite("./nrdata/nr_data.json", recordData, true)
}

func TestNR_LoadSumCache(t *testing.T) {
	neb := testNeb(t)
	dip, err := NewNR(neb)
	assert.Nil(t, err)
	dip.loadSumCache()
}

func TestNR_LoadNRCache(t *testing.T) {
	neb := testNeb(t)
	dip, err := NewNR(neb)
	assert.Nil(t, err)
	dip.loadNRCache()
}
