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

// CreatingContext carries the context in creatingStateMachine
type CreatingContext struct {
	parent *core.Block

	index          uint32
	dynastyRoot    byteutils.Hash
	dynastyChanged uint32
	validators     []byteutils.Hash
	maxVotes       uint32

	onCanonical bool
}

// NewCreatingContext create a new creating context
func NewCreatingContext(parent *core.Block, tail *core.Block) (*CreatingContext, error) {
	var err error

	context := &CreatingContext{}
	context.parent = parent

	context.index = 0
	context.dynastyChanged = 0
	context.dynastyRoot, err = parent.NextBlockDynastyRoot()
	if err != nil {
		return nil, err
	}
	context.maxVotes, err = parent.DynastySize(context.dynastyRoot)
	if err != nil {
		return nil, err
	}
	context.validators, err = parent.NextBlockSortedValidators()
	if err != nil {
		return nil, err
	}

	context.onCanonical = false
	if parent.Hash().Equals(tail.Hash()) {
		context.onCanonical = true
	}
	return context, nil
}

// CreatedContext carries the context in createdStateMachine
type CreatedContext struct {
	block       *core.Block
	maxVotes    uint32
	onCanonical bool
}

// NewCreatedContext create a new creating context
func NewCreatedContext(block *core.Block, onCanonical bool) (*CreatedContext, error) {
	var err error
	context := &CreatedContext{}
	context.block = block
	context.maxVotes, err = block.DynastySize(block.CurDynastyRoot())
	if err != nil {
		return nil, err
	}
	context.onCanonical = onCanonical
	return context, nil
}
