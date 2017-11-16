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

	"github.com/nebulasio/go-nebulas/common/trie"
	"github.com/nebulasio/go-nebulas/core/state"
	"github.com/nebulasio/go-nebulas/storage"
)

// Genesis Block Hash
var (
	GenesisHash = make([]byte, BlockHashLength)
)

// NewGenesisBlock create genesis @Block from file.
func NewGenesisBlock(chainID uint32, storage storage.Storage) *Block {
	accState, _ := state.NewAccountState(nil, storage)
	txsTrie, _ := trie.NewBatchTrie(nil, storage)

	dynastyTrie, _ := trie.NewBatchTrie(nil, storage)
	nextDynastyTrie, _ := trie.NewBatchTrie(nil, storage)
	dynastyCandidatesTrie, _ := trie.NewBatchTrie(nil, storage)
	depositTrie, _ := trie.NewBatchTrie(nil, storage)
	prepareVotesTrie, _ := trie.NewBatchTrie(nil, storage)
	commitVotesTrie, _ := trie.NewBatchTrie(nil, storage)
	changeVotesTrie, _ := trie.NewBatchTrie(nil, storage)
	abdicateVotesTrie, _ := trie.NewBatchTrie(nil, storage)
	blocksHeightTrie, _ := trie.NewBatchTrie(nil, storage)

	b := &Block{
		header: &BlockHeader{
			chainID:           chainID,
			hash:              GenesisHash,
			parentHash:        GenesisHash,
			dynastyParentHash: GenesisHash,
			coinbase:          &Address{make([]byte, AddressLength)},
			timestamp:         time.Now().Unix(),
		},
		accState: accState,
		txsTrie:  txsTrie,

		dynastyTrie:           dynastyTrie,
		nextDynastyTrie:       nextDynastyTrie,
		dynastyCandidatesTrie: dynastyCandidatesTrie,
		depositTrie:           depositTrie,
		prepareVotesTrie:      prepareVotesTrie,
		commitVotesTrie:       commitVotesTrie,
		changeVotesTrie:       changeVotesTrie,
		abdicateVotesTrie:     abdicateVotesTrie,
		blocksHeightTrie:      blocksHeightTrie,

		storage: storage,
		height:  1,
		sealed:  true,
	}

	return b
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
