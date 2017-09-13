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

	"github.com/nebulasio/go-nebulas/blockchain"
	"github.com/nebulasio/go-nebulas/consensus"
	log "github.com/sirupsen/logrus"
)

const (
	Prepare = "prepare"
)

type PrepareState struct {
	p *Pow
}

func NewPrepareState(p *Pow) *PrepareState {
	state := &PrepareState{p: p}
	return state
}

func (state *PrepareState) Event(e consensus.Event) consensus.State {
	log.WithFields(log.Fields{"stateType": fmt.Sprintf("%T", state), "eventType": fmt.Sprintf("%T", e)}).Warn("ignore this event.")
	return state
}

func (state *PrepareState) Enter(data interface{}) {
	log.Info("PrepareState enter.")

	// get the pending block.
	state.p.newBlock = state.p.chain.NewBlock(blockchain.NewAddress("1234567890"))

	// move to mining state.
	state.p.TransiteByKey(Mining, nil)
}

func (state *PrepareState) Leave(data interface{}) {
	log.Info("PrepareState leave.")
}
