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
	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/common/trie"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/core/state"
)

// Genesis Block Hash
var (
	GenesisHash      = make([]byte, BlockHashLength)
	GenesisTimestamp = int64(0)
	GenesisDynasty   = []string{
		"3e9f008b7feb2cb867004edf1bcc441788516ad40e853d2d",
		"7d843eaf57a0d5842e067a3a75c29c031558db9a1e7301e8",
		"8e07853e0017b49c3ffd24d8388c83e2fa507f359aa57ab0",
		"9b095a889f52fb9411a91c02c79dcc501b63213d28db2c79",
		"13a8135878d2e25e277b9821c17c0bd0e759f66e0cf1266d",
		"36b19a1bd73facc0ffbf39038642da35503b07691f347915",
		"43c04a7992c7b4c59c5d12f2f940e6d1d579d0db07756087",
		"43faedd5182b44ee160cb47e444f9460057abca0daed3a6f",
		"49e3681c8e659c28ba22533484af503bdfdc1a5ea5d492f7",
		"831def042a8060f68c197ea09aab16f8905c096baed8269c",
		"11235560b1af46f22cbc2f70808d21a7284ffe0c24f2196f",
		"a6c3254ad8a449e2c5fa30c64cdd9cd4d8aea77d39139e3d",
		"b54e25518956babbf4e77abff6aca2d3cbadd4a178f6078b",
		"bf493723544625d1eeb1c64c1fbb57866615c3c17b2e877f",
		"c155c7bfda0a714229f92797dfffd3eee213239e7605037c",
		"cb7d1d7c1e46377b1752d1b25dd8574d03e2507c35b71061",
		"d704c735d2896b930f3d46a198d3f976a66ee88985454ba2",
		"dc2e617a9c2724d5afc01133ecddb38dad191f83af7d7270",
		"e43e7290297947732c2aac546750dbc39ab00857f491ca56",
		"efcbaefb78e80c0fb8f1d424636809dfdbd7479501c0f1a9",
		"f732100adc168bd8292463879adb0654908ec5773c4a5de1",
	}
)

// NewGenesisBlock create genesis @Block from file.
func NewGenesisBlock(chainID uint32, chain *BlockChain) *Block {
	accState, _ := state.NewAccountState(nil, chain.storage)
	txsTrie, _ := trie.NewBatchTrie(nil, chain.storage)
	dposContext := NewDposContext(chain.storage)

	genesis := &Block{
		header: &BlockHeader{
			chainID:     chainID,
			hash:        GenesisHash,
			parentHash:  GenesisHash,
			dposContext: &corepb.DposContext{},
			coinbase:    &Address{make([]byte, AddressLength)},
			timestamp:   GenesisTimestamp,
			nonce:       0,
		},
		accState:    accState,
		txsTrie:     txsTrie,
		dposContext: dposContext,
		txPool:      chain.txPool,
		storage:     chain.storage,
		height:      1,
		sealed:      true,
	}

	dynasty, _ := trie.NewBatchTrie(nil, chain.storage)
	delegate, _ := trie.NewBatchTrie(nil, chain.storage)
	for _, v := range GenesisDynasty {
		member, _ := AddressParse(v)
		dynasty.Put(member.Bytes(), member.Bytes())
		vote, _ := proto.Marshal(
			&corepb.Delegate{
				Delegator: member.Bytes(),
				Delegatee: member.Bytes(),
			},
		)
		delegate.Put(member.Bytes(), vote)
	}
	genesis.dposContext.dynastyTrie = dynasty
	genesis.header.dposContext.DynastyRoot = dynasty.RootHash()
	genesis.dposContext.nextDynastyTrie = dynasty
	genesis.header.dposContext.NextDynastyRoot = dynasty.RootHash()
	genesis.dposContext.delegateTrie = delegate
	genesis.header.dposContext.DelegateRoot = delegate.RootHash()

	return genesis
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
