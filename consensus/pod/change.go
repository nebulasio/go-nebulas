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

	"github.com/nebulasio/go-nebulas/util/byteutils"

	"github.com/nebulasio/go-nebulas/core"

	"github.com/nebulasio/go-nebulas/util"

	"github.com/nebulasio/go-nebulas/consensus"
	log "github.com/sirupsen/logrus"
)

// ChangeState presents the prepare stage in pod
type ChangeState struct {
	sm      *consensus.StateMachine
	votes   map[byteutils.HexHash]bool
	context *CreatingContext
}

// NewChangeState create a new prepare state
func NewChangeState(sm *consensus.StateMachine, context *CreatingContext) *ChangeState {
	return &ChangeState{
		sm:      sm,
		context: context,
		votes:   make(map[byteutils.HexHash]bool),
	}
}

func (state *ChangeState) String() string {
	return fmt.Sprintf("ChangeState %p", state)
}

// Event handle event.
func (state *ChangeState) Event(e consensus.Event) (bool, consensus.State) {
	switch e.EventType() {
	case NewChangeVoteEvent:
		voter := e.Data().(byteutils.Hash)
		state.votes[voter.Hex()] = true
		changeVotes := uint32(len(state.votes))
		activeN := uint32(len(state.context.validators)) - state.context.index
		if changeVotes > state.context.maxVotes*2/3 {
			state.context.index++
			return true, NewCreationState(state.sm, state.context)
		}
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
	case NewAbdicateTimeoutEvent:
		return true, NewAbdicateState(state.sm, state.context)
	}
	return false, nil
}

// Enter called when transiting to this state.
func (state *ChangeState) Enter(data interface{}) {
	log.Debug("ChangeState enter.")
	// if the block is on canonical chain, vote & set timeout
	if state.context.onCanonical {
		p := state.sm.Context().(*PoD)
		zero := util.NewUint128()
		nonce := p.chain.TailBlock().GetNonce(p.coinbase.Bytes())
		payload, err := core.NewChangeVotePayload(core.ChangeAction, state.context.parent.Hash(), state.context.index+1).ToBytes()
		if err != nil {
			panic(err)
		}
		changeTx := core.NewTransaction(state.context.parent.ChainID(), p.coinbase, p.coinbase, zero, nonce+1, core.TxPayloadVoteType, payload)
		p.neblet.AccountManager().SignTransaction(p.coinbase, changeTx)
		p.chain.TransactionPool().PushAndBroadcast(changeTx)
		log.WithFields(log.Fields{
			"func":       "PoD.ChangeState",
			"block hash": state.context.parent.Hash(),
			"n":          state.context.index + 1,
		}).Info("Vote Change.")

		time.AfterFunc(120*time.Second, func() {
			state.sm.Event(consensus.NewBaseEvent(NewAbdicateTimeoutEvent, nil))
		})
	}
}

// Leave called when leaving this state.
func (state *ChangeState) Leave(data interface{}) {
	log.Debug("ChangeState leave.")
}
