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
	log "github.com/sirupsen/logrus"
)

const (
	// Start tart state key.
	Start = "Start"
)

// StartState the initial state of @Pow state machine.
// TODO(@roy): can be interrupted
type StartState struct {
	p *Pow
}

// NewStartState create StartState instance.
func NewStartState(p *Pow) *StartState {
	state := &StartState{p: p}
	return state
}

// Event handle event.
func (state *StartState) Event(e consensus.Event) (bool, consensus.State) {
	switch e.EventType() {
	case consensus.CanMiningEvent:
		return true, state.p.states[Prepare]
	case consensus.NewBlockEvent:
		return true, state.p.states[Minted]
	default:
		return false, nil
	}
}

// Enter called when transiting to this state.
func (state *StartState) Enter(data interface{}) {
	log.Debug("StartState enter.")
	if state.p.CanMining() {
		state.p.TransitByKey(Prepare, nil)
	}
}

// Leave called when leaving this state.
func (state *StartState) Leave(data interface{}) {
	log.Debug("StartState leave.")
}
