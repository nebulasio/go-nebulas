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

	"github.com/nebulasio/go-nebulas/consensus"
	log "github.com/sirupsen/logrus"
)

// CreationState presents the prepare stage in pod
type CreationState struct {
	sm  *consensus.StateMachine
	pod *PoD
}

// NewCreationState create a new prepare state
func NewCreationState(sm *consensus.StateMachine, pod *PoD) *CreationState {
	return &CreationState{
		sm:  sm,
		pod: pod,
	}
}

func (state *CreationState) String() string {
	return fmt.Sprintf("CreationState %p", state)
}

// Event handle event.
func (state *CreationState) Event(e consensus.Event) (bool, consensus.State) {
	return false, nil
}

// Enter called when transiting to this state.
func (state *CreationState) Enter(data interface{}) {
	log.Debug("CreationState enter.")
}

// Leave called when leaving this state.
func (state *CreationState) Leave(data interface{}) {
	log.Debug("CreationState leave.")
}
