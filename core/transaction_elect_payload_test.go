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
	"time"

	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/crypto/keystore/secp256k1"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/stretchr/testify/assert"
)

var (
	ks   = keystore.DefaultKS
	zero = util.NewUint128()
)

func GenerateNewAddress() *Address {
	priv, _ := secp256k1.GeneratePrivateKey()
	pubdata, _ := priv.PublicKey().Encoded()
	addr, _ := NewAddressFromPublicKey(pubdata)
	ks.SetKey(addr.ToHex(), priv, []byte("passphrase"))
	ks.Unlock(addr.ToHex(), []byte("passphrase"), time.Second*60*60*24*365)
	return addr
}

func TestElectPayload_BaseElect(t *testing.T) {
	storage, _ := storage.NewMemoryStorage()
	genesis := NewGenesisBlock(0, storage, nil)
	size, _ := countValidators(genesis.dynastyCandidatesTrie, nil)
	validators := []*Address{}
	for i := 0; i < 10; i++ {
		v := GenerateNewAddress()
		validators = append(validators, v)
	}
	coinbase := validators[0]
	block := NewBlock(genesis.header.chainID, coinbase, genesis)
	block.begin()
	loginPayload, _ := NewElectPayload(LoginAction).ToBytes()
	logoutPayload, _ := NewElectPayload(LogoutAction).ToBytes()
	withdrawPayload, _ := NewElectPayload(WithdrawAction).ToBytes()
	for i := 0; i < 10; i++ {
		v := validators[i]
		if i < 6 {
			// give them enough balance to submit deposit
			account := block.accState.GetOrCreateUserAccount(v.Bytes())
			account.AddBalance(StandardDeposit)
		}
		tx := NewTransaction(block.header.chainID, validators[i], validators[i], zero, 1, TxPayloadElectType, loginPayload)
		giveback, err := block.executeTransaction(tx)
		assert.Equal(t, giveback, false)
		if i < 6 {
			assert.Nil(t, err)
		} else {
			assert.NotNil(t, err)
		}
	}
	cnt, err := countValidators(block.dynastyCandidatesTrie, nil)
	assert.Equal(t, cnt, uint32(size+6))
	assert.Nil(t, err)
	for i := 0; i < 10; i++ {
		v := validators[i]
		if i < 3 {
			// logout success
			tx := NewTransaction(block.header.chainID, v, v, zero, 2, TxPayloadElectType, logoutPayload)
			giveback, err := block.executeTransaction(tx)
			assert.Equal(t, giveback, false)
			assert.Nil(t, err)
		}
		if i >= 6 {
			// logout fail, not login yet
			tx := NewTransaction(block.header.chainID, v, v, zero, 2, TxPayloadElectType, logoutPayload)
			giveback, err := block.executeTransaction(tx)
			assert.Equal(t, giveback, false)
			assert.NotNil(t, err)
		}
	}
	for i := 0; i < 10; i++ {
		v := validators[i]
		if i < 3 {
			// withdraw success
			tx := NewTransaction(block.header.chainID, v, v, zero, 3, TxPayloadElectType, withdrawPayload)
			giveback, err := block.executeTransaction(tx)
			assert.Equal(t, giveback, false)
			assert.Nil(t, err)
		} else if i < 6 {
			// withdraw fail, not logout yet
			tx := NewTransaction(block.header.chainID, v, v, zero, 2, TxPayloadElectType, withdrawPayload)
			giveback, err := block.executeTransaction(tx)
			assert.Equal(t, giveback, false)
			assert.NotNil(t, err)
		} else {
			// withdraw fail, not login yet
			tx := NewTransaction(block.header.chainID, v, v, zero, 3, TxPayloadElectType, withdrawPayload)
			giveback, err := block.executeTransaction(tx)
			assert.Equal(t, giveback, false)
			assert.NotNil(t, err)
		}
	}
	block.commit()
}

func TestElectPayload_WithdrawWhileInNextDynasty(t *testing.T) {
	storage, _ := storage.NewMemoryStorage()
	genesis := NewGenesisBlock(0, storage, nil)
	validators, _ := genesis.NextBlockSortedValidators()
	loginPayload, _ := NewElectPayload(LoginAction).ToBytes()
	logoutPayload, _ := NewElectPayload(LogoutAction).ToBytes()
	withdrawPayload, _ := NewElectPayload(WithdrawAction).ToBytes()

	coinbase := &Address{validators[1]}
	block := NewBlock(genesis.header.chainID, coinbase, genesis)
	block.begin()

	logoutAddress := &Address{validators[0]}
	tx := NewTransaction(block.header.chainID, logoutAddress, logoutAddress, zero, 1, TxPayloadElectType, logoutPayload)
	giveback, err := block.executeTransaction(tx)
	assert.Equal(t, giveback, false)
	assert.Nil(t, err)

	v := GenerateNewAddress()
	account := block.accState.GetOrCreateUserAccount(v.Bytes())
	account.AddBalance(StandardDeposit)
	tx = NewTransaction(block.header.chainID, v, v, zero, 1, TxPayloadElectType, loginPayload)
	giveback, err = block.executeTransaction(tx)
	assert.Equal(t, giveback, false)
	assert.Nil(t, err)

	change, err := block.CheckDynastyRule()
	assert.Nil(t, err)
	assert.Equal(t, change, false)
	block.ChangeDynasty()

	tx = NewTransaction(block.header.chainID, v, v, zero, 2, TxPayloadElectType, logoutPayload)
	giveback, err = block.executeTransaction(tx)
	assert.Equal(t, giveback, false)
	assert.Nil(t, err)
	tx = NewTransaction(block.header.chainID, v, v, zero, 3, TxPayloadElectType, withdrawPayload)
	giveback, err = block.executeTransaction(tx)
	assert.Equal(t, giveback, false)
	assert.NotNil(t, err)
}

func TestElectPayload_DynastyRule(t *testing.T) {
	storage, _ := storage.NewMemoryStorage()
	genesis := NewGenesisBlock(0, storage, nil)
	validators, _ := genesis.NextBlockSortedValidators()
	coinbase := &Address{validators[0]}
	block := NewBlock(genesis.header.chainID, coinbase, genesis)
	block.begin()

	curBlock := block
	curBlock.Seal()
	for i := 0; i < EpochSize-2; i++ {
		curBlock = NewBlock(curBlock.header.chainID, curBlock.Coinbase(), curBlock)
		curBlock.Seal()
	}
	dynastyRoot, err := curBlock.NextBlockDynastyRoot()
	assert.Nil(t, err)
	assert.Equal(t, dynastyRoot, curBlock.CurDynastyRoot())

	curBlock = NewBlock(curBlock.header.chainID, curBlock.Coinbase(), curBlock)
	curBlock.Seal()

	change, err := curBlock.CheckDynastyRule()
	assert.Nil(t, err)
	assert.Equal(t, change, true)
	block.ChangeDynasty()

	dynastyRoot, err = curBlock.NextBlockDynastyRoot()
	assert.Nil(t, err)
	assert.Equal(t, dynastyRoot, curBlock.NextDynastyRoot())
}
