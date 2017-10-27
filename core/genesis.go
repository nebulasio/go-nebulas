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

	"github.com/nebulasio/go-nebulas/common/batch_trie"
	"github.com/nebulasio/go-nebulas/storage"
)

// NewGenesisBlock create genesis @Block from file.
func NewGenesisBlock(chainID uint32, storage storage.Storage) *Block {
	stateTrie, _ := batchtrie.NewBatchTrie(nil, storage)
	txsTrie, _ := batchtrie.NewBatchTrie(nil, storage)
	// TODO: load genesis block data from file.
	b := &Block{
		header: &BlockHeader{
			chainID:    chainID,
			hash:       make([]byte, BlockHashLength),
			parentHash: make([]byte, BlockHashLength),
			coinbase:   &Address{make([]byte, AddressLength)},
			timestamp:  time.Now().Unix(),
		},
		stateTrie: stateTrie,
		txsTrie:   txsTrie,
		height:    1,
		sealed:    true,
	}

	return b
}
