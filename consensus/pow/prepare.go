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
	"github.com/nebulasio/go-nebulas/crypto/keystore"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"time"
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

func randomTx(from *core.Address) *core.Transaction {

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	idx := r.Intn(3)
	for {
		if idx > 0 {
			break
		} else {
			idx = rand.Intn(2)
		}
	}
	alias, _, _ := keystore.DefaultKS.GetKeyByIndex(idx)
	to, _ := core.Parse(alias)

	value := rand.Uint64() % 5

	tx := core.NewTransaction(*from, *to, value, 0, nil)
	tx.Sign()
	return tx
}

// Enter called when transiting to this state.
func (state *PrepareState) Enter(data interface{}) {
	log.Debug("PrepareState enter.")

	p := state.p

	//TODO(larry.wang):later remove test address
	alias, _, _ := keystore.DefaultKS.GetKeyByIndex(0)
	addr, _ := core.Parse(alias)

	if p.miningBlock == nil {
		// start mining from chain tail.
		p.miningBlock = state.p.chain.NewBlock(addr)
		//TODO(larry.wang):test trans
		tx := randomTx(addr)
		p.miningBlock.AddTransactions(tx)
	} else if p.miningBlock.Sealed() {
		// start mining from local minted block.
		parentBlock := p.miningBlock
		p.miningBlock = state.p.chain.NewBlockFromParent(addr, parentBlock)
		//TODO(larry.wang):test trans
		tx := randomTx(addr)
		p.miningBlock.AddTransactions(tx)
	}

	// move to mining state.
	state.p.TransitByKey(Mining, nil)
}

// Leave called when leaving this state.
func (state *PrepareState) Leave(data interface{}) {
	log.Debug("PrepareState leave.")
}
