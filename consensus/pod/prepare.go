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

	"github.com/nebulasio/go-nebulas/util/byteutils"

	"github.com/nebulasio/go-nebulas/consensus"
	log "github.com/sirupsen/logrus"
)

// PrepareState presents the prepare stage in pod
type PrepareState struct {
	sm      *consensus.StateMachine
	votes   map[byteutils.HexHash]bool
	context *CreatedContext
}

// NewPrepareState create a new prepare state
func NewPrepareState(sm *consensus.StateMachine, context *CreatedContext) *PrepareState {
	return &PrepareState{
		sm:      sm,
		context: context,
	}
}

func (state *PrepareState) String() string {
	return fmt.Sprintf("PrepareState %p", state)
}

// Event handle event.
func (state *PrepareState) Event(e consensus.Event) (bool, consensus.State) {
	switch e.EventType() {
	case NewPrepareVoteEvent:
		voter := e.Data().(byteutils.Hash)
		state.votes[voter.Hex()] = true
		prepareVotes := uint32(len(state.votes))
		if prepareVotes > state.context.maxVotes*2/3 {
			return true, NewCommitState(state.sm, state.context)
		}
		return false, nil
	}
	return false, nil
}

// Enter called when transiting to this state.
func (state *PrepareState) Enter(data interface{}) {
	log.Debug("PrepareState enter.")
	// if the block is on canonical chain, vote
	if state.context.onCanonical {
		p := state.sm.Context().(*PoD)
		zero := util.NewUint128()
		nonce := p.chain.TailBlock().GetNonce(p.coinbase.Bytes())
		payload, err := core.NewPrepareVotePayload(core.PrepareAction, state.context.block.Hash(), state.context.block.Height(), 1).ToBytes()
		if err != nil {
			panic(err)
		}
		prepareTx := core.NewTransaction(state.context.block.ChainID(), p.coinbase, p.coinbase, zero, nonce+1, core.TxPayloadVoteType, payload)
		p.neblet.AccountManager().SignTransaction(p.coinbase, prepareTx)
		p.nm.Broadcast(consensus.MessageTypeNewTx, prepareTx)
		log.WithFields(log.Fields{
			"func":       "PoD.PrepareState",
			"block hash": state.context.block.Hash(),
		}).Info("Vote Prepare.")
	}
}

// Leave called when leaving this state.
func (state *PrepareState) Leave(data interface{}) {
	log.Debug("PrepareState leave.")
}
