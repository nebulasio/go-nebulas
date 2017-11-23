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

	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/util"

	"github.com/nebulasio/go-nebulas/consensus"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	log "github.com/sirupsen/logrus"
)

// AbdicateState presents the prepare stage in pod
type AbdicateState struct {
	sm      *consensus.StateMachine
	votes   map[byteutils.HexHash]bool
	context *CreatingContext
}

// NewAbdicateState create a new prepare state
func NewAbdicateState(sm *consensus.StateMachine, context *CreatingContext) *AbdicateState {
	return &AbdicateState{
		sm:      sm,
		context: context,
		votes:   make(map[byteutils.HexHash]bool),
	}
}

func (state *AbdicateState) String() string {
	return fmt.Sprintf("AbdicateState %p", state)
}

// Event handle event.
func (state *AbdicateState) Event(e consensus.Event) (bool, consensus.State) {
	switch e.EventType() {
	case NewAbdicateVoteEvent:
		voter := e.Data().(byteutils.Hash)
		state.votes[voter.Hex()] = true
		abdicateVotes := uint32(len(state.votes))
		activeN := uint32(len(state.context.validators)) - abdicateVotes
		if activeN <= state.context.maxVotes*2/3 {
			state.context.index = 0
			state.context.dynastyChanged++
			// TODO(roy): change to next dynasty, emergency
			state.context.dynastyRoot = state.context.parent.NextDynastyRoot()
			var err error
			state.context.maxVotes, err = state.context.parent.DynastySize(state.context.dynastyRoot)
			if err != nil {
				panic(err)
			}
			state.context.validators, err = state.context.parent.DynastyValidators(state.context.dynastyRoot)
			if err != nil {
				panic(err)
			}
			return true, NewCreationState(state.sm, state.context)
		}
		return false, nil
	}
	return false, nil
}

// Enter called when transiting to this state.
func (state *AbdicateState) Enter(data interface{}) {
	log.Debug("AbdicateState enter.")
	// if the block is on canonical chain, vote
	if state.context.onCanonical {
		p := state.sm.Context().(*PoD)
		zero := util.NewUint128()
		nonce := p.chain.TailBlock().GetNonce(p.coinbase.Bytes())
		payload, err := core.NewAbdicateVotePayload(core.AbdicateAction, state.context.parent.Hash(), state.context.dynastyRoot).ToBytes()
		if err != nil {
			panic(err)
		}
		abdicateTx := core.NewTransaction(state.context.parent.ChainID(), p.coinbase, p.coinbase, zero, nonce+1, core.TxPayloadVoteType, payload)
		p.neblet.AccountManager().SignTransaction(p.coinbase, abdicateTx)
		p.chain.TransactionPool().PushAndBroadcast(abdicateTx)
		log.WithFields(log.Fields{
			"func":         "PoD.Abdicate",
			"block hash":   state.context.parent.Hash(),
			"dynasty root": state.context.dynastyRoot,
		}).Info("Vote Abdicate.")
	}
}

// Leave called when leaving this state.
func (state *AbdicateState) Leave(data interface{}) {
	log.Debug("AbdicateState leave.")
}
