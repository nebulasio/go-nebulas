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

	changed     uint32
	dynastyRoot byteutils.Hash
	validators  []byteutils.Hash
	maxVotes    uint32

	chosen bool
}

// NewCreatingContext create a new creating context
func NewCreatingContext(coinbase byteutils.Hash, parent *core.Block, chosen bool) (*CreatingContext, error) {
	var err error

	context := &CreatingContext{}
	context.parent = parent
	context.changed = 0

	// check current dynasty
	change, err := parent.CheckDynastyRule()
	if err != nil {
		return nil, err
	}
	if change {
		parent.ChangeDynasty()
	}
	context.dynastyRoot = parent.CurDynastyRoot()
	context.maxVotes, err = parent.DynastySize(context.dynastyRoot)
	if err != nil {
		return nil, err
	}
	context.validators, err = parent.SortedActiveValidators(context.dynastyRoot)
	if err != nil {
		return nil, err
	}

	// check if current coinbase is a validator
	context.chosen = false
	for _, v := range context.validators {
		if coinbase.Equals(v) {
			context.chosen = chosen
		}
	}

	return context, nil
}

// CreatedContext carries the context in createdStateMachine
type CreatedContext struct {
	block *core.Block

	dynastyRoot byteutils.Hash
	validators  []byteutils.Hash
	maxVotes    uint32

	chosen bool
}

// NewCreatedContext create a new creating context
func NewCreatedContext(coinbase byteutils.Hash, block *core.Block, chosen bool) (*CreatedContext, error) {
	var err error

	context := &CreatedContext{}
	context.block = block

	// check current dynasty
	context.dynastyRoot = block.CurDynastyRoot()
	context.maxVotes, err = block.DynastySize(context.dynastyRoot)
	if err != nil {
		return nil, err
	}
	context.validators, err = block.SortedActiveValidators(context.dynastyRoot)
	if err != nil {
		return nil, err
	}

	// check if current coinbase is a validator
	context.chosen = false
	for _, v := range context.validators {
		if coinbase.Equals(v) {
			context.chosen = chosen
		}
	}

	return context, nil
}
