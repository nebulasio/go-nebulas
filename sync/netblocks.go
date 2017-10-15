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

package sync

import (
	"errors"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/core/pb"
)

// NetBlocks structure
type NetBlocks struct {
	from   string
	blocks []*core.Block
}

// NetBlock structure
type NetBlock struct {
	from  string
	block *core.Block
}

// NewNetBlocks return new Blocks.
func NewNetBlocks(from string, blocks []*core.Block) *NetBlocks {
	bs := &NetBlocks{from: from, blocks: blocks}
	return bs
}

// NewNetBlock return new Blocks.
func NewNetBlock(from string, block *core.Block) *NetBlock {
	b := &NetBlock{from: from, block: block}
	return b
}

// Blocks return blocks.
func (nbs *NetBlocks) Blocks() []*core.Block {
	return nbs.blocks
}

// ToProto converts domain Blocks into proto Blocks
func (nbs *NetBlocks) ToProto() (proto.Message, error) {
	var result []*corepb.Block
	for _, v := range nbs.blocks {
		block, err := v.ToProto()
		if err != nil {
			return nil, err
		}
		if block, ok := block.(*corepb.Block); ok {
			result = append(result, block)
		} else {
			return nil, errors.New("Pb Message cannot be converted into Block")
		}
	}
	return &corepb.NetBlocks{
		From:   nbs.from,
		Blocks: result,
	}, nil
}

// FromProto converts proto Blocks to domain Blocks
func (nbs *NetBlocks) FromProto(msg proto.Message) error {
	if msg, ok := msg.(*corepb.NetBlocks); ok {
		nbs.from = msg.From
		for _, v := range msg.Blocks {
			block := new(core.Block)
			if err := block.FromProto(v); err != nil {
				return err
			}
			nbs.blocks = append(nbs.blocks, block)
		}
		return nil
	}
	return errors.New("Pb Message cannot be converted into NetBlocks")
}

// Block return block.
func (nb *NetBlock) Block() *core.Block {
	return nb.block
}

// ToProto converts domain Block into proto Block
func (nb *NetBlock) ToProto() (proto.Message, error) {
	block, err := nb.block.ToProto()
	if err != nil {
		return nil, err
	}
	return &corepb.NetBlock{
		From:  nb.from,
		Block: block.(*corepb.Block),
	}, nil
}

// FromProto converts proto Blocks to domain Blocks
func (nb *NetBlock) FromProto(msg proto.Message) error {
	if msg, ok := msg.(*corepb.NetBlock); ok {
		nb.from = msg.From
		block := new(core.Block)
		if err := block.FromProto(msg.Block); err != nil {
			return err
		}
		nb.block = block
		return nil
	}
	return errors.New("Pb Message cannot be converted into NetBlock")
}
