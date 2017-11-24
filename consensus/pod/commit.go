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

	"github.com/nebulasio/go-nebulas/util"

	"github.com/nebulasio/go-nebulas/util/byteutils"

	"github.com/nebulasio/go-nebulas/consensus"
	"github.com/nebulasio/go-nebulas/core"
	log "github.com/sirupsen/logrus"
)

// CommitState presents the prepare stage in pod
type CommitState struct {
	sm      *consensus.StateMachine
	votes   map[byteutils.HexHash]bool
	context *CreatedContext
}

// NewCommitState create a new prepare state
func NewCommitState(sm *consensus.StateMachine, context *CreatedContext) *CommitState {
	return &CommitState{
		sm:      sm,
		votes:   make(map[byteutils.HexHash]bool),
		context: context,
	}
}

func (state *CommitState) String() string {
	return fmt.Sprintf("CommitState %p", state)
}

// Event handle event.
func (state *CommitState) Event(e consensus.Event) (bool, consensus.State) {
	switch e.EventType() {
	case NewCommitVoteEvent:
		voter := byteutils.Hash(e.Data().([]byte))
		state.votes[voter.Hex()] = true
		commitVotes := uint32(len(state.votes))
		limit := state.context.maxVotes * 2 / 3
		if commitVotes-1 <= limit && commitVotes > limit {
			log.WithFields(log.Fields{
				"func":       "PoD.CommitState",
				"block hash": state.context.block.Hash(),
				"height":     state.context.block.Height(),
			}).Info("Finality.")
		}
		return false, nil
	}
	return false, nil
}

// Enter called when transiting to this state.
func (state *CommitState) Enter(data interface{}) {
	log.Debug("CommitState enter.")
	// if the block is on canonical chain, vote
	log.Infof("Chosen. %v", state.context.chosen)
	if state.context.chosen {
		p := state.sm.Context().(*PoD)
		zero := util.NewUint128()
		payload, err := core.NewCommitVotePayload(core.CommitAction, state.context.block.Hash()).ToBytes()
		if err != nil {
			panic(err)
		}
		commitTx := core.NewTransaction(state.context.block.ChainID(), p.coinbase, p.coinbase, zero, p.nonce+1, core.TxPayloadVoteType, payload)
		p.nonce++
		p.neblet.AccountManager().SignTransaction(p.coinbase, commitTx)
		p.chain.TransactionPool().PushAndBroadcast(commitTx)
		log.WithFields(log.Fields{
			"func":       "PoD.CommitState",
			"block hash": state.context.block.Hash(),
			"height":     state.context.block.Height(),
		}).Info("Vote Commit.")
	}
}

// Leave called when leaving this state.
func (state *CommitState) Leave(data interface{}) {
	log.Debug("CommitState leave.")
}
