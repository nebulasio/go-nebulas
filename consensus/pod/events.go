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

package pod

import (
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/util/byteutils"
)

// Event Type List
const (
	NewBlockEvent           = "event.newblock"
	NewChangeTimeoutEvent   = "event.timeout.change"
	NewAbdicateTimeoutEvent = "event.timeout.abdicate"
	NewPrepareVoteEvent     = "event.vote.prepare"
	NewCommitVoteEvent      = "event.vote.commit"
	NewChangeVoteEvent      = "event.vote.change"
	NewAbdicateVoteEvent    = "event.vote.abdicate"
	CanMiningEvent          = "event.canmining"
)

// StateContext carries the context in state transitions
type StateContext struct {
	Tails     map[byteutils.HexHash]*core.Block
	Proposers map[byteutils.HexHash]byteutils.Hash
}

// NewStateContext create a new state context
func NewStateContext(tails []*core.Block) *StateContext {
	ctx := &StateContext{
		Tails: make(map[byteutils.HexHash]*core.Block),
	}
	for _, v := range tails {
		ctx.Tails[v.Hash().Hex()] = v
	}
	return ctx
}
