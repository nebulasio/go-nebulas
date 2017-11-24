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
	"fmt"
	"time"

	"github.com/nebulasio/go-nebulas/consensus"
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	log "github.com/sirupsen/logrus"
)

// CreationState presents the prepare stage in pod
type CreationState struct {
	sm      *consensus.StateMachine
	got     bool
	context *CreatingContext
}

// NewCreationState create a new prepare state
func NewCreationState(sm *consensus.StateMachine, context *CreatingContext) *CreationState {
	return &CreationState{
		sm:      sm,
		got:     false,
		context: context,
	}
}

func (state *CreationState) String() string {
	return fmt.Sprintf("CreationState %p", state)
}

// Proposer return current proposer who should propose the new block
func (state *CreationState) Proposer() byteutils.Hash {
	return state.context.validators[0]
}

// Event handle event.
func (state *CreationState) Event(e consensus.Event) (bool, consensus.State) {
	switch e.EventType() {
	case NewBlockEvent:
		block := e.Data().(*core.Block)
		p := state.sm.Context().(*PoD)
		createdStateMachine := p.newAndStartCreatedStateMachine(block)
		p.createdStateMachines.Add(block.Hash().Hex(), createdStateMachine)
		// stop the state machine
		state.sm.Stop()
		p.creatingStateMachines.Remove(state.context.parent.Hash().Hex())
		state.got = true
		return false, nil
	case NewChangeTimeoutEvent:
		if !state.got {
			return true, NewChangeState(state.sm, state.context)
		}
	}
	return false, nil
}

// Enter called when transiting to this state.
func (state *CreationState) Enter(data interface{}) {
	log.Debugf("CreationState enter. %p", state)
	// if the block is on canonical chain, create or set timeout
	log.Infof("Chosen. %v", state.context.chosen)
	if state.context.chosen {
		p := state.sm.Context().(*PoD)
		parent := p.chain.GetBlock(state.context.parent.Hash())
		log.Infof("Proposer %s height %s me %s", state.Proposer().Hex(), parent.Height(), p.coinbase.ToHex())
		if state.Proposer().Equals(p.coinbase.Bytes()) {
			block := core.NewBlock(state.context.parent.ChainID(), p.coinbase, parent)
			block.CollectTransactions(200)
			block.Seal()
			log.WithFields(log.Fields{
				"func":   "PoD.CreationState",
				"block":  block,
				"parent": state.context.parent,
			}).Infof("Create New Block.")
			go p.chain.BlockPool().PushAndBroadcast(block)
		} else {
			time.AfterFunc(5*time.Second, func() {
				state.sm.Event(consensus.NewBaseEvent(NewChangeTimeoutEvent, nil))
			})
		}
	}
}

// Leave called when leaving this state.
func (state *CreationState) Leave(data interface{}) {
	log.Debugf("CreationState leave. %p", state)
}
