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
)

// EventType list
const (
	TimeoutEvent         = "event.timeout"
	NewPrepareVoteEvent  = "event.vote.prepare"
	NewCommitVoteEvent   = "event.vote.commit"
	NewChangeVoteEvent   = "event.vote.change"
	NewAbdicateVoteEvent = "event.vote.abdicate"
	CanMiningEvent       = "event.canmining"
)

// StateContext carries the context in state transitions
type StateContext struct {
	parentBlock *core.Block
	miningBlock *core.Block

	prepareVotes  map[string]bool
	commitVotes   map[string]bool
	changeVotes   map[string]bool
	abdicateVotes map[string]bool
}

// NewStateContext create a new state context
func NewStateContext(parentBlock *core.Block) *StateContext {
	return &StateContext{
		parentBlock:   parentBlock,
		prepareVotes:  make(map[string]bool),
		commitVotes:   make(map[string]bool),
		changeVotes:   make(map[string]bool),
		abdicateVotes: make(map[string]bool),
	}
}

// NextProposer change current proposer to next
func (sc *StateContext) NextProposer() {

}
