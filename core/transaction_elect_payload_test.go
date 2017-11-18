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
	assert.Equal(t, cnt, 6)
	assert.Nil(t, err)
	loginValidators, err := traverseValidators(block.dynastyCandidatesTrie, nil)
	assert.Nil(t, err)
	for _, v := range loginValidators {
		exist := false
		for i := 0; i < 6; i++ {
			if v.Equals(validators[i].Bytes()) {
				exist = true
				continue
			}
		}
		assert.Equal(t, exist, true)
	}
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
	cnt, err = countValidators(block.depositTrie, nil)
	assert.Nil(t, err)
	assert.Equal(t, cnt, 3)
	deposit, err := StandardDeposit.ToFixedSizeByteSlice()
	assert.Nil(t, err)
	for i := 3; i < 6; i++ {
		d, err := block.depositTrie.Get(validators[i].Bytes())
		assert.Nil(t, err)
		assert.Equal(t, deposit, d)
	}
	block.commit()
}

func TestElectPayload_WithdrawWhileInNextDynasty(t *testing.T) {
	storage, _ := storage.NewMemoryStorage()
	genesis := NewGenesisBlock(0, storage, nil)
	coinbase := GenerateNewAddress()
	block := NewBlock(genesis.header.chainID, coinbase, genesis)
	block.begin()
	loginPayload, _ := NewElectPayload(LoginAction).ToBytes()
	logoutPayload, _ := NewElectPayload(LogoutAction).ToBytes()
	withdrawPayload, _ := NewElectPayload(WithdrawAction).ToBytes()
	validators := []*Address{}
	for i := 0; i < 5; i++ {
		v := GenerateNewAddress()
		validators = append(validators, v)
		account := block.accState.GetOrCreateUserAccount(v.Bytes())
		account.AddBalance(StandardDeposit)
		tx := NewTransaction(block.header.chainID, validators[i], validators[i], zero, 1, TxPayloadElectType, loginPayload)
		giveback, err := block.executeTransaction(tx)
		assert.Equal(t, giveback, false)
		assert.Nil(t, err)
	}
	cnt, err := countValidators(block.dynastyCandidatesTrie, nil)
	assert.Nil(t, err)
	assert.Equal(t, cnt, 5)
	cnt, err = countValidators(block.curDynastyTrie, nil)
	assert.Nil(t, err)
	assert.Equal(t, cnt, 0)
	cnt, err = countValidators(block.nextDynastyTrie, nil)
	assert.Nil(t, err)
	assert.Equal(t, cnt, 0)

	change, err := block.checkDynastyRule()
	assert.Nil(t, err)
	assert.Equal(t, change, true)
	block.changeDynasty()

	cnt, err = countValidators(block.curDynastyTrie, nil)
	assert.Nil(t, err)
	assert.Equal(t, cnt, 0)
	cnt, err = countValidators(block.nextDynastyTrie, nil)
	assert.Nil(t, err)
	assert.Equal(t, cnt, 5)

	tx := NewTransaction(block.header.chainID, validators[0], validators[0], zero, 2, TxPayloadElectType, logoutPayload)
	giveback, err := block.executeTransaction(tx)
	assert.Equal(t, giveback, false)
	assert.Nil(t, err)
	tx = NewTransaction(block.header.chainID, validators[0], validators[0], zero, 3, TxPayloadElectType, withdrawPayload)
	giveback, err = block.executeTransaction(tx)
	assert.Equal(t, giveback, false)
	assert.NotNil(t, err)
}

func TestElectPayload_DynastyRule(t *testing.T) {
	storage, _ := storage.NewMemoryStorage()
	genesis := NewGenesisBlock(0, storage, nil)
	coinbase := GenerateNewAddress()
	block := NewBlock(genesis.header.chainID, coinbase, genesis)
	block.begin()
	loginPayload, _ := NewElectPayload(LoginAction).ToBytes()
	logoutPayload, _ := NewElectPayload(LogoutAction).ToBytes()
	validators := []*Address{}
	for i := 0; i < DynastySize*2/3; i++ {
		v := GenerateNewAddress()
		validators = append(validators, v)
		account := block.accState.GetOrCreateUserAccount(v.Bytes())
		account.AddBalance(StandardDeposit)
		tx := NewTransaction(block.header.chainID, validators[i], validators[i], zero, 1, TxPayloadElectType, loginPayload)
		giveback, err := block.executeTransaction(tx)
		assert.Equal(t, giveback, false)
		assert.Nil(t, err)
	}
	cnt, err := countValidators(block.dynastyCandidatesTrie, nil)
	assert.Nil(t, err)
	assert.Equal(t, cnt, DynastySize*2/3)
	cnt, err = countValidators(block.curDynastyTrie, nil)
	assert.Nil(t, err)
	assert.Equal(t, cnt, 0)
	cnt, err = countValidators(block.nextDynastyTrie, nil)
	assert.Nil(t, err)
	assert.Equal(t, cnt, 0)

	// cur: 0, next: 2/3, candidates: 2/3
	change, err := block.checkDynastyRule()
	assert.Nil(t, err)
	assert.Equal(t, change, true)
	block.changeDynasty()

	cnt, err = countValidators(block.curDynastyTrie, nil)
	assert.Nil(t, err)
	assert.Equal(t, cnt, 0)
	cnt, err = countValidators(block.nextDynastyTrie, nil)
	assert.Nil(t, err)
	assert.Equal(t, cnt, DynastySize*2/3)

	tx := NewTransaction(block.header.chainID, validators[0], validators[0], zero, 2, TxPayloadElectType, logoutPayload)
	giveback, err := block.executeTransaction(tx)
	assert.Equal(t, giveback, false)
	assert.Nil(t, err)

	// cur: 2/3, next: 2/3-1, canidates: 2/3-1
	change, err = block.checkDynastyRule()
	assert.Nil(t, err)
	assert.Equal(t, change, true)
	block.changeDynasty()

	cnt, err = countValidators(block.curDynastyTrie, nil)
	assert.Nil(t, err)
	assert.Equal(t, cnt, DynastySize*2/3)
	cnt, err = countValidators(block.nextDynastyTrie, nil)
	assert.Nil(t, err)
	assert.Equal(t, cnt, DynastySize*2/3-1)
	cnt, err = countValidators(block.validatorsTrie, block.curDynastyTrie.RootHash())
	assert.Nil(t, err)
	assert.Equal(t, cnt, DynastySize*2/3)
	cnt, err = countValidators(block.validatorsTrie, block.nextDynastyTrie.RootHash())
	assert.Nil(t, err)
	assert.Equal(t, cnt, DynastySize*2/3-1)

	// cannot change dynasty
	change, err = block.checkDynastyRule()
	assert.Nil(t, err)
	assert.Equal(t, change, false)

	curBlock := block
	curBlock.Seal()
	dynastyRoot, err := block.NextBlockDynastyRoot()
	assert.Nil(t, err)
	assert.Equal(t, dynastyRoot, block.CurDynastyRoot())
	for i := 0; i < EpochSize-2; i++ {
		curBlock = NewBlock(curBlock.header.chainID, curBlock.Coinbase(), curBlock)
		curBlock.Seal()
	}
	// create EpochSize - 1 blocks, cannot change
	dynastyRoot, err = curBlock.NextBlockDynastyRoot()
	assert.Nil(t, err)
	assert.Equal(t, dynastyRoot, curBlock.CurDynastyRoot())

	curBlock = NewBlock(curBlock.header.chainID, curBlock.Coinbase(), curBlock)
	curBlock.Seal()
	// create EpochSize blocks, change dynasty
	// cur: 2/3-1, next: 2/3 -1, candidates: 2/3-1
	change, err = curBlock.checkDynastyRule()
	assert.Nil(t, err)
	assert.Equal(t, change, true)
	block.changeDynasty()
	// < 2/3 validators, change dynasty
	change, err = curBlock.checkDynastyRule()
	assert.Nil(t, err)
	assert.Equal(t, change, true)
}
