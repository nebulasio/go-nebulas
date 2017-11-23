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
	"time"

	"github.com/nebulasio/go-nebulas/util/byteutils"

	"github.com/nebulasio/go-nebulas/common/trie"
	"github.com/nebulasio/go-nebulas/core/state"
	"github.com/nebulasio/go-nebulas/storage"
)

// Genesis Block Hash
var (
	GenesisHash = make([]byte, BlockHashLength)
)

// NewGenesisBlock create genesis @Block from file.
func NewGenesisBlock(chainID uint32, storage storage.Storage, txPool *TransactionPool) *Block {
	accState, _ := state.NewAccountState(nil, storage)
	txsTrie, _ := trie.NewBatchTrie(nil, storage)

	curDynastyTrie, _ := trie.NewBatchTrie(nil, storage)
	nextDynastyTrie, _ := trie.NewBatchTrie(nil, storage)
	dynastyCandidatesTrie, _ := trie.NewBatchTrie(nil, storage)
	validatorsTrie, _ := trie.NewBatchTrie(nil, storage)
	depositTrie, _ := trie.NewBatchTrie(nil, storage)
	prepareVotesTrie, _ := trie.NewBatchTrie(nil, storage)
	heightPrepareVotesTrie, _ := trie.NewBatchTrie(nil, storage)
	commitVotesTrie, _ := trie.NewBatchTrie(nil, storage)
	changeVotesTrie, _ := trie.NewBatchTrie(nil, storage)
	abdicateVotesTrie, _ := trie.NewBatchTrie(nil, storage)

	block := &Block{
		header: &BlockHeader{
			chainID:    chainID,
			parentHash: GenesisHash,
			coinbase:   &Address{make([]byte, AddressLength)},
			nonce:      0,
			timestamp:  time.Now().Unix(),
		},
		transactions: make(Transactions, 0),
		parentBlock:  nil,
		accState:     accState,
		txsTrie:      txsTrie,

		curDynastyTrie:         curDynastyTrie,
		nextDynastyTrie:        nextDynastyTrie,
		dynastyCandidatesTrie:  dynastyCandidatesTrie,
		validatorsTrie:         validatorsTrie,
		depositTrie:            depositTrie,
		prepareVotesTrie:       prepareVotesTrie,
		heightPrepareVotesTrie: heightPrepareVotesTrie,
		commitVotesTrie:        commitVotesTrie,
		changeVotesTrie:        changeVotesTrie,
		abdicateVotesTrie:      abdicateVotesTrie,

		storage: storage,
		txPool:  txPool,
		height:  1,
		sealed:  false,
	}

	validators := []string{
		"5cdadc1cfe3da0a3d067e9f1b195b90c5aebfb5afc8d43b4",
		"8a209cec02cbeab7e2f74ad969d2dfe8dd24416aa65589bf",
		"22ac3a9a2b1c31b7a9084e46eae16e761f83f02324092b09",
	}
	for _, v := range validators {
		addr, err := byteutils.FromHex(v)
		if err != nil {
			panic(err)
		}
		block.dynastyCandidatesTrie.Put(addr, addr)
		// default login
		block.addDeposit(addr, StandardDeposit)
	}
	block.changeDynasty()
	block.Seal()

	return block
}

// CheckGenesisBlock if a block is a genesis block
func CheckGenesisBlock(block *Block) bool {
	if block == nil {
		return false
	}
	if block.ParentHash().Equals(GenesisHash) {
		return true
	}
	return false
}
