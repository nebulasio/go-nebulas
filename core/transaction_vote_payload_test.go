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

	"github.com/nebulasio/go-nebulas/storage"
	"github.com/stretchr/testify/assert"
)

func NewBlockWithValidDynasty(t *testing.T, size int) *Block {
	storage, _ := storage.NewMemoryStorage()
	genesis := NewGenesisBlock(0, storage, nil)
	genesis.begin()
	loginPayload, _ := NewElectPayload(LoginAction).ToBytes()
	for i := 0; i < size; i++ {
		v := GenerateNewAddress()
		account := genesis.accState.GetOrCreateUserAccount(v.Bytes())
		account.AddBalance(StandardDeposit)
		tx := NewTransaction(genesis.header.chainID, v, v, zero, 1, TxPayloadElectType, loginPayload)
		giveback, err := genesis.executeTransaction(tx)
		assert.Equal(t, giveback, false)
		assert.Nil(t, err)
	}
	genesis.commit()

	coinbase := GenerateNewAddress()
	block1 := NewBlock(genesis.header.chainID, coinbase, genesis)
	cnt, _ := countValidators(block1.curDynastyTrie, nil)
	assert.Equal(t, cnt, 0)
	cnt, _ = countValidators(block1.nextDynastyTrie, nil)
	assert.Equal(t, cnt, size)
	cnt, _ = countValidators(block1.dynastyCandidatesTrie, nil)
	assert.Equal(t, cnt, size)
	block2 := NewBlock(block1.header.chainID, coinbase, block1)
	cnt, _ = countValidators(block2.curDynastyTrie, nil)
	assert.Equal(t, cnt, size)
	cnt, _ = countValidators(block2.nextDynastyTrie, nil)
	assert.Equal(t, cnt, size)
	cnt, _ = countValidators(block2.dynastyCandidatesTrie, nil)
	assert.Equal(t, cnt, size)
	return block1
}

func TestVotePayload_Execute(t *testing.T) {
	block := NewBlockWithValidDynasty(t, 3)
	block.begin()

}
