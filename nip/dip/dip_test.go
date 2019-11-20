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

package dip

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/nebulasio/go-nebulas/storage"

	"github.com/nebulasio/go-nebulas/account"
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/crypto/keystore/secp256k1"
	nebletpb "github.com/nebulasio/go-nebulas/neblet/pb"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/stretchr/testify/assert"
)

var (
	dipAddr, _ = core.AddressParse("n1HWE22w4DaE1gz6E9D9TGokrxfHBegoyB8")
)

func mockAddress() *core.Address {
	ks := keystore.DefaultKS
	priv1 := secp256k1.GeneratePrivateKey()
	pubdata1, _ := priv1.PublicKey().Encoded()
	addr, _ := core.NewAddressFromPublicKey(pubdata1)
	ks.SetKey(addr.String(), priv1, []byte("passphrase"))
	ks.Unlock(addr.String(), []byte("passphrase"), time.Second*60*60*24*365)
	return addr
}

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

func TestDip_CheckReward(t *testing.T) {
	neb := testNeb(t)
	dip, err := NewDIP(neb)
	assert.Nil(t, err)

	tests := []struct {
		key    string
		txType string
		from   *core.Address
		to     *core.Address
		value  string
		err    error
	}{
		{
			key:    "binary tx",
			txType: core.TxPayloadBinaryType,
			from:   mockAddress(),
			to:     nil,
			value:  "0",
			err:    nil,
		},
		{
			key:    "binary tx with dip from",
			txType: core.TxPayloadBinaryType,
			from:   dip.RewardAddress(),
			to:     nil,
			value:  "0",
			err:    ErrUnsupportedTransactionFromDipAddress,
		},
		{
			key:    "deploy tx",
			txType: core.TxPayloadDeployType,
			from:   mockAddress(),
			to:     nil,
			value:  "0",
			err:    nil,
		},
		{
			key:    "deploy tx with dip from",
			txType: core.TxPayloadDeployType,
			from:   dip.RewardAddress(),
			to:     nil,
			value:  "0",
			err:    ErrUnsupportedTransactionFromDipAddress,
		},
		{
			key:    "call tx",
			txType: core.TxPayloadCallType,
			from:   mockAddress(),
			to:     nil,
			value:  "0",
			err:    nil,
		},
		{
			key:    "call tx with dip from",
			txType: core.TxPayloadCallType,
			from:   dip.RewardAddress(),
			to:     nil,
			value:  "0",
			err:    ErrUnsupportedTransactionFromDipAddress,
		},
		{
			key:    "protocol tx",
			txType: core.TxPayloadProtocolType,
			from:   mockAddress(),
			to:     nil,
			value:  "0",
			err:    nil,
		},
		{
			key:    "protocol tx with dip from",
			txType: core.TxPayloadProtocolType,
			from:   dip.RewardAddress(),
			to:     nil,
			value:  "0",
			err:    ErrUnsupportedTransactionFromDipAddress,
		},
		{
			key:    "dip tx not form dip addr",
			txType: core.TxPayloadDipType,
			from:   mockAddress(),
			to:     nil,
			value:  "0",
			err:    ErrInvalidDipAddress,
		},
		{
			key:    "dip tx addr not found in list",
			txType: core.TxPayloadDipType,
			from:   dip.RewardAddress(),
			to:     nil,
			value:  "1",
			err:    ErrDipNotFound,
		},
		{
			key:    "dip tx value not found",
			txType: core.TxPayloadDipType,
			from:   dip.RewardAddress(),
			to:     dipAddr,
			value:  "0",
			err:    ErrDipNotFound,
		},
		{
			key:    "dip tx normal",
			txType: core.TxPayloadDipType,
			from:   dip.RewardAddress(),
			to:     dipAddr,
			value:  "1",
			err:    nil,
		},
	}

	//data, err := dip.GetDipList(1)
	//assert.Nil(t, err)
	//dipData := data.(*DIPData)
	//println("dip addr:", dipData.Dips[0].Address)
	//println("dip value:", dipData.Dips[0].Reward)

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			to := tt.to
			if to == nil {
				to = mockAddress()
			}
			value, err := util.NewUint128FromString(tt.value)
			assert.Nil(t, err)
			var (
				payloadBytes []byte
			)
			if tt.txType == core.TxPayloadDipType {
				payload, err := core.NewDipPayload(1, 1, 1, "")
				assert.Nil(t, err)
				payloadBytes, err = payload.ToBytes()
				assert.Nil(t, err)
			}
			tx, err := core.NewTransaction(11, tt.from, to, value, 1, tt.txType, payloadBytes, core.TransactionGasPrice, core.TransactionMaxGas)
			assert.Nil(t, err)
			err = dip.CheckReward(tx)
			assert.Equal(t, tt.err, err)
		})
	}
}

func readNbreDB(key string) ([]byte, error) {
	rs, err := storage.NewRocksStorage("../../mainnet/nbre/nbre.db")
	if err != nil {
		return nil, err
	}
	data, err := rs.Get([]byte(key))
	if err != nil {
		return nil, err
	}
	return data, nil
}

func TestDip_ReadDIP(t *testing.T) {
	dbdata, err := readNbreDB("dip_rewards")
	assert.Nil(t, err)
	data := DipReward{}
	err = json.Unmarshal(dbdata, &data)
	assert.Nil(t, err)
	items := make([]*DIPData, len(data.DipRewards))
	for i, str := range data.DipRewards {
		item := DIPData{}
		err := json.Unmarshal([]byte(str), &item)
		assert.Nil(t, err)
		items[i] = &item
	}
	assert.Nil(t, err)
	//recordData, err := json.Marshal(items)
	//assert.Nil(t, err)
	//util.FileWrite("./dipdata/dip_data.json", recordData, true)
}

func TestNR_LoadCache(t *testing.T) {
	neb := testNeb(t)
	dip, err := NewDIP(neb)
	assert.Nil(t, err)
	dip.loadCache()
}
