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
	log "github.com/sirupsen/logrus"
)

const (
	Minted = "minted"
)

type MintedState struct {
	p *Pow
}

func NewMintedState(p *Pow) *MintedState {
	state := &MintedState{p: p}
	return state
}

func (state *MintedState) Event(e consensus.Event) consensus.State {
	log.WithFields(log.Fields{"stateType": fmt.Sprintf("%T", state), "eventType": fmt.Sprintf("%T", e)}).Warn("ignore this event.")
	return state
}

func (state *MintedState) Enter(data interface{}) {
	log.Info("MintedState enter.")

	// TODO: submit the block.

	// append to local blockchain.
	state.p.newBlock.Sign()
	state.p.chain.Append(state.p.newBlock)

	// send to net manager.
	state.p.nm.SendNewBlock(state.p.newBlock)

	// move to prepare state.
	state.p.TransiteByKey(Prepare, nil)
}

func (state *MintedState) Leave(data interface{}) {
	log.Info("MintedState leave.")
}
