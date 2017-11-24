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
	var err error

	switch e.EventType() {
	case NewAbdicateVoteEvent:
		voter := e.Data().(byteutils.Hash)
		state.votes[voter.Hex()] = true
		abdicateVotes := uint32(len(state.votes))
		activeN := uint32(len(state.context.validators)) - abdicateVotes
		if activeN <= state.context.maxVotes*2/3 {
			currentDynasty := state.context.parent.CurDynastyRoot()
			state.context.parent.ChangeDynasty()
			state.context.dynastyRoot = state.context.parent.CurDynastyRoot()
			state.context.maxVotes, err = state.context.parent.DynastySize(state.context.dynastyRoot)
			if err != nil {
				log.Error(err)
				break
			}
			state.context.validators, err = state.context.parent.SortedActiveValidators(state.context.dynastyRoot)
			if err != nil {
				log.Error(err)
				break
			}
			log.WithFields(log.Fields{
				"func":       "PoD.AbdicateState",
				"block hash": state.context.parent.Hash(),
				"height":     state.context.parent.Height(),
				"from":       currentDynasty,
				"to":         state.context.dynastyRoot.Hex(),
			}).Info("Dynasty Changed.")
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
	log.Infof("Chosen. %v", state.context.chosen)
	if state.context.chosen {
		p := state.sm.Context().(*PoD)
		zero := util.NewUint128()
		nonce := p.chain.TailBlock().GetNonce(p.coinbase.Bytes())
		payload, err := core.NewAbdicateVotePayload(core.AbdicateAction, state.context.parent.Hash(), state.context.dynastyRoot).ToBytes()
		if err != nil {
			log.Error(err)
			return
		}
		_, err = p.abdicateDynastyTrie.Put(state.context.dynastyRoot, p.coinbase.Bytes())
		if err != nil {
			log.Error(err)
			return
		}
		abdicateTx := core.NewTransaction(state.context.parent.ChainID(), p.coinbase, p.coinbase, zero, nonce+1, core.TxPayloadVoteType, payload)
		p.neblet.AccountManager().SignTransaction(p.coinbase, abdicateTx)
		p.chain.TransactionPool().PushAndBroadcast(abdicateTx)

		// check abdicate votes

		log.WithFields(log.Fields{
			"func":       "PoD.AbdicateState",
			"block hash": state.context.parent.Hash(),
			"from":       state.context.dynastyRoot,
		}).Info("Vote Abdicate.")
	}
}

// Leave called when leaving this state.
func (state *AbdicateState) Leave(data interface{}) {
	log.Debug("AbdicateState leave.")
}
