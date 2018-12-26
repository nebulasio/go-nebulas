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
	"testing"
	"time"

	"github.com/nebulasio/go-nebulas/account"
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/crypto/keystore/secp256k1"
	"github.com/nebulasio/go-nebulas/neblet/pb"
	"github.com/nebulasio/go-nebulas/nf/nbre"
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
	nbre   core.Nbre
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

func (n *mockNeb) Nbre() core.Nbre {
	return nil
}

type mockNbre struct {
}

func (m *mockNbre) Start() error {
	return nil
}
func (m *mockNbre) Execute(command string, params []byte) ([]byte, error) {
	if command == nbre.CommandDIPList {
		data := &DIPData{
			StartHeight: 1,
			EndHeight:   100,
			Dips:        make([]*DIPItem, 1),
		}
		item := &DIPItem{
			Address: dipAddr.String(),
			Reward:  "1",
		}
		data.Dips[0] = item
		bytes, _ := data.ToBytes()
		return bytes, nil
	} else {
		return nil, nil
	}
}
func (m *mockNbre) Shutdown() error {
	return nil
}

func testNeb(t *testing.T) *mockNeb {

	account, err := account.NewManager(nil)
	assert.Nil(t, err)
	neb := &mockNeb{
		config: &nebletpb.Config{Chain: &nebletpb.ChainConfig{ChainId: 1},
			Nbre: &nebletpb.NbreConfig{},
		},
		am:   account,
		nbre: &mockNbre{},
	}
	return neb
}

func TestDip_CheckReward(t *testing.T) {
	neb := testNeb(t)
	dip, err := NewDIP(neb)
	assert.Nil(t, err)

	tests := []struct {
		key    string
		height uint64
		txType string
		from   *core.Address
		to     *core.Address
		err    error
	}{
		{
			key:    "binary tx",
			height: 1,
			txType: core.TxPayloadBinaryType,
			from:   mockAddress(),
			to:     nil,
			err:    nil,
		},
		{
			key:    "binary tx with dip from",
			height: 1,
			txType: core.TxPayloadBinaryType,
			from:   dip.RewardAddress(),
			to:     nil,
			err:    ErrUnsupportedTransactionFromDipAddress,
		},
		{
			key:    "deploy tx",
			height: 1,
			txType: core.TxPayloadDeployType,
			from:   mockAddress(),
			to:     nil,
			err:    nil,
		},
		{
			key:    "deploy tx with dip from",
			height: 1,
			txType: core.TxPayloadDeployType,
			from:   dip.RewardAddress(),
			to:     nil,
			err:    ErrUnsupportedTransactionFromDipAddress,
		},
		{
			key:    "call tx",
			height: 1,
			txType: core.TxPayloadCallType,
			from:   mockAddress(),
			to:     nil,
			err:    nil,
		},
		{
			key:    "call tx with dip from",
			height: 1,
			txType: core.TxPayloadCallType,
			from:   dip.RewardAddress(),
			to:     nil,
			err:    ErrUnsupportedTransactionFromDipAddress,
		},
		{
			key:    "protocol tx",
			height: 1,
			txType: core.TxPayloadProtocolType,
			from:   mockAddress(),
			to:     nil,
			err:    nil,
		},
		{
			key:    "protocol tx",
			height: 1,
			txType: core.TxPayloadProtocolType,
			from:   dip.RewardAddress(),
			to:     nil,
			err:    ErrUnsupportedTransactionFromDipAddress,
		},
		{
			key:    "dip tx",
			height: 1,
			txType: core.TxPayloadDipType,
			from:   mockAddress(),
			to:     nil,
			err:    ErrInvalidDipAddress,
		},
		{
			key:    "dip tx",
			height: 1,
			txType: core.TxPayloadDipType,
			from:   dip.RewardAddress(),
			to:     nil,
			err:    ErrDipNotFound,
		},
		{
			key:    "dip tx",
			height: 1,
			txType: core.TxPayloadDipType,
			from:   dip.RewardAddress(),
			to:     dipAddr,
			err:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			to := tt.to
			if to == nil {
				to = mockAddress()
			}
			tx, err := core.NewTransaction(11, tt.from, to, util.NewUint128(), 1, tt.txType, nil, core.TransactionGasPrice, core.TransactionMaxGas)
			assert.Nil(t, err)
			err = dip.CheckReward(tt.height, tx)
			assert.Equal(t, tt.err, err)
		})
	}
}
