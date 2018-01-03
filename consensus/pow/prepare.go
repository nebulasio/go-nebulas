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

package pow

import (
	"fmt"

	"github.com/nebulasio/go-nebulas/consensus"
	"github.com/nebulasio/go-nebulas/util/logging"
)

const (
	// Prepare prepare state key.
	Prepare = "prepare"
)

// PrepareState the initial state of @Pow state machine.
// TODO(@roy): can be interrupted
type PrepareState struct {
	p *Pow
}

// NewPrepareState create PrepareState instance.
func NewPrepareState(p *Pow) *PrepareState {
	state := &PrepareState{p: p}
	return state
}

func (state *PrepareState) String() string {
	return fmt.Sprintf("PrepareState %p", state)
}

// Event handle event.
func (state *PrepareState) Event(e consensus.Event) (bool, consensus.State) {
	return false, nil
}

// Enter called when transiting to this state.
func (state *PrepareState) Enter(data interface{}) {
	logging.VLog().Info("PrepareState enter.")

	p := state.p

	if p.miningBlock == nil {
		// start mining from chain tail.
		p.miningBlock, _ = state.p.chain.NewBlock(p.coinbase)
		p.miningBlock.CollectTransactions(2)
	} else if p.miningBlock.Sealed() {
		// start mining from local minted block.
		parentBlock := p.miningBlock
		p.miningBlock, _ = state.p.chain.NewBlockFromParent(p.coinbase, parentBlock)
		p.miningBlock.CollectTransactions(2)
	}

	// move to mining state.
	state.p.Transit(state, NewMiningState(state.p), nil)
}

// Leave called when leaving this state.
func (state *PrepareState) Leave(data interface{}) {
	logging.VLog().Info("PrepareState leave.")
}
