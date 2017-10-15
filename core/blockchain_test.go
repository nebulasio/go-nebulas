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

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

func BlockFromNetwork(block *Block) *Block {
	pb, _ := block.ToProto()
	ir, _ := proto.Marshal(pb)
	proto.Unmarshal(ir, pb)
	b := new(Block)
	b.FromProto(pb)
	return b
}

func TestBlockChain_FindCommonAncestorWithTail(t *testing.T) {
	bc := NewBlockChain(0)
	var c MockConsensus
	bc.SetConsensusHandler(c)
	coinbase11 := &Address{[]byte("coinbase11")}
	coinbase12 := &Address{[]byte("coinbase12")}
	coinbase111 := &Address{[]byte("coinbase111")}
	coinbase221 := &Address{[]byte("coinbase221")}
	coinbase222 := &Address{[]byte("coinbase222")}
	coinbase1111 := &Address{[]byte("coinbase1111")}
	/*
		genesisi -- 11 -- 111 -- 1111
				 \_ 12 -- 221
				       \_ 222 tail
	*/
	block11 := bc.NewBlock(coinbase11)
	block12 := bc.NewBlock(coinbase12)
	block11.header.hash = HashBlock(block11)
	block12.header.hash = HashBlock(block12)
	bc.BlockPool().Push(block11)
	bc.BlockPool().Push(block12)
	bc.SetTailBlock(block11)
	block111 := bc.NewBlock(coinbase111)
	block111.header.hash = HashBlock(block111)
	bc.BlockPool().Push(block111)
	bc.SetTailBlock(block12)
	block221 := bc.NewBlock(coinbase221)
	block222 := bc.NewBlock(coinbase222)
	block221.header.hash = HashBlock(block221)
	block222.header.hash = HashBlock(block222)
	bc.BlockPool().Push(block221)
	bc.BlockPool().Push(block222)
	bc.SetTailBlock(block111)
	block1111 := bc.NewBlock(coinbase1111)
	block1111.header.hash = HashBlock(block1111)
	bc.BlockPool().Push(block1111)
	bc.SetTailBlock(block222)
	test := &Block{
		header: &BlockHeader{
			coinbase: &Address{},
		},
	}
	_, err := bc.FindCommonAncestorWithTail(BlockFromNetwork(test))
	assert.NotNil(t, err)
	common1, _ := bc.FindCommonAncestorWithTail(BlockFromNetwork(block1111))
	assert.Equal(t, BlockFromNetwork(common1), BlockFromNetwork(bc.genesisBlock))
	common2, _ := bc.FindCommonAncestorWithTail(BlockFromNetwork(block221))
	assert.Equal(t, BlockFromNetwork(common2), BlockFromNetwork(block12))
	common3, _ := bc.FindCommonAncestorWithTail(BlockFromNetwork(block222))
	assert.Equal(t, BlockFromNetwork(common3), BlockFromNetwork(block222))
	common4, _ := bc.FindCommonAncestorWithTail(BlockFromNetwork(bc.tailBlock))
	assert.Equal(t, BlockFromNetwork(common4), BlockFromNetwork(bc.tailBlock))
	common5, _ := bc.FindCommonAncestorWithTail(BlockFromNetwork(block12))
	assert.Equal(t, BlockFromNetwork(common5), BlockFromNetwork(block12))
}

func TestBlockChain_FetchDescendantInCanonicalChain(t *testing.T) {
	bc := NewBlockChain(0)
	var c MockConsensus
	bc.SetConsensusHandler(c)
	coinbase := &Address{[]byte("coinbase")}
	/*
		genesisi -- 1 - 2 - 3 - 4 - 5 - 6
		         \_ block - block1
	*/
	block := bc.NewBlock(coinbase)
	block.header.timestamp = 0
	block.header.hash = HashBlock(block)
	bc.BlockPool().Push(block)
	block1 := bc.NewBlock(coinbase)
	block1.header.timestamp = 1
	block1.header.hash = HashBlock(block1)
	bc.BlockPool().Push(block1)

	var blocks []*Block
	for i := 0; i < 6; i++ {
		block := bc.NewBlock(coinbase)
		if i > 0 {
			block.header.timestamp = blocks[i-1].header.timestamp + 1
		}
		blocks = append(blocks, block)
		block.header.hash = HashBlock(block)
		bc.BlockPool().Push(block)
		bc.SetTailBlock(block)
	}
	blocks24, _ := bc.FetchDescendantInCanonicalChain(3, blocks[0])
	assert.Equal(t, BlockFromNetwork(blocks24[0]), BlockFromNetwork(blocks[1]))
	assert.Equal(t, BlockFromNetwork(blocks24[1]), BlockFromNetwork(blocks[2]))
	assert.Equal(t, BlockFromNetwork(blocks24[2]), BlockFromNetwork(blocks[3]))
	blocks46, _ := bc.FetchDescendantInCanonicalChain(10, blocks[2])
	assert.Equal(t, len(blocks46), 3)
	assert.Equal(t, BlockFromNetwork(blocks46[0]), BlockFromNetwork(blocks[3]))
	assert.Equal(t, BlockFromNetwork(blocks46[1]), BlockFromNetwork(blocks[4]))
	assert.Equal(t, BlockFromNetwork(blocks46[2]), BlockFromNetwork(blocks[5]))
	blocks13, _ := bc.FetchDescendantInCanonicalChain(3, bc.genesisBlock)
	assert.Equal(t, len(blocks13), 3)
	_, err := bc.FetchDescendantInCanonicalChain(3, block)
	assert.NotNil(t, err)
	blocks0, err0 := bc.FetchDescendantInCanonicalChain(3, blocks[5])
	assert.Equal(t, len(blocks0), 0)
	assert.Nil(t, err0)
}
