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
	"github.com/nebulasio/go-nebulas/consensus"
	"github.com/nebulasio/go-nebulas/core"
	log "github.com/sirupsen/logrus"
)

const (
	// Prepare prepare state key.
	Prepare = "prepare"
)

// PrepareState the initial state of @Pow state machine.
type PrepareState struct {
	p *Pow
}

// NewPrepareState create PrepareState instance.
func NewPrepareState(p *Pow) *PrepareState {
	state := &PrepareState{p: p}
	return state
}

// Event handle event.
func (state *PrepareState) Event(e consensus.Event) (bool, consensus.State) {
	return false, nil
}

// Enter called when transiting to this state.
func (state *PrepareState) Enter(data interface{}) {
	log.Debug("PrepareState enter.")

	// get the pending block.
	state.p.newBlock = state.p.chain.NewBlock(core.NewAddress("1234567890"))

	// move to mining state.
	state.p.TransiteByKey(Mining, nil)
}

// Leave called when leaving this state.
func (state *PrepareState) Leave(data interface{}) {
	log.Debug("PrepareState leave.")
}
