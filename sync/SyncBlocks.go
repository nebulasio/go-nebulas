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

// SyncBlocks structure
type SyncBlocks struct {
	addrs  string
	blocks []*core.Block
}

// NewSyncBlocks return new Blocks.
func NewSyncBlocks(addrs string, blocks []*core.Block) *SyncBlocks {
	result := &SyncBlocks{addrs: addrs, blocks: blocks}
	return result
}

// Blocks return blocks.
func (sb *SyncBlocks) Blocks() []*core.Block {
	return sb.blocks
}

// ToProto converts domain Blocks into proto Blocks
func (sb *SyncBlocks) ToProto() (proto.Message, error) {
	var result []*corepb.Block
	for _, v := range sb.blocks {
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
	return &corepb.SyncBlocks{
		Addrs:  sb.addrs,
		Blocks: result,
	}, nil
}

// FromProto converts proto Blocks to domain Blocks
func (sb *SyncBlocks) FromProto(msg proto.Message) error {
	if msg, ok := msg.(*corepb.SyncBlocks); ok {
		sb.addrs = msg.Addrs
		for _, v := range msg.Blocks {
			block := new(core.Block)
			if err := block.FromProto(v); err != nil {
				return err
			}
			sb.blocks = append(sb.blocks, block)
		}
		return nil
	}
	return errors.New("Pb Message cannot be converted into SyncBlocks")
}
