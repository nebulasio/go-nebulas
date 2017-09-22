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
	"github.com/nebulasio/go-nebulas/util/byteutils"

	log "github.com/sirupsen/logrus"
	"time"
)

const (
	// Mining mining state key
	Mining = "mining"
)

// MiningState mining state.
type MiningState struct {
	p      *Pow
	quitCh chan bool

	//
	nonce      uint64
	parentHash string
}

// NewMiningState create MiningState instance.
func NewMiningState(p *Pow) *MiningState {
	state := &MiningState{p: p, quitCh: make(chan bool, 1)}
	return state
}

// Event handle event.
func (state *MiningState) Event(e consensus.Event) (bool, consensus.State) {
	if e.EventType() != consensus.NewBlockEvent {
		return false, nil
	}

	// when received new block event, quit searchingNonce(), go to next state.
	block := e.Data().(*core.Block)
	log.WithFields(log.Fields{
		"block": block,
	}).Info("MiningState.Event: receive new block message, transit to MintedState.")
	return true, state.p.states[Minted]
}

// Enter called when transiting to this state.
func (state *MiningState) Enter(data interface{}) {
	log.Debug("MiningState.Enter: enter.")

	// start searching nonce.
	go state.searchingNonce()
}

// Leave called when leaving this state.
func (state *MiningState) Leave(data interface{}) {
	log.Debug("MiningState.Leave: leave.")
	state.quitCh <- true
}

func (state *MiningState) searchingNonce() {
	// transit to MintedState if newBlockReceived is true.
	if state.p.newBlockReceived {
		log.Info("MiningState.Enter: new block received, transit to MintedState.")
		state.p.TransitByKey(Minted, nil)

	} else {
		// calculate hash.
		miningBlock := state.p.miningBlock
		nonce := miningBlock.Nonce()
		parentHash := miningBlock.ParentHash()

	computeHash:
		for {
			select {
			case <-state.quitCh:
				log.Info("MiningState.searchingNonce: quit.")
				return

			default:
				nonce++

				// compute hash..
				miningBlock.SetNonce(nonce)
				miningBlock.SetTimestamp(time.Now())
				resultBytes := HashAndVerify(miningBlock)

				if resultBytes != nil {
					log.WithFields(log.Fields{
						"nonce":      nonce,
						"parentHash": byteutils.Hex(parentHash),
						"hashResult": byteutils.Hex(resultBytes),
					}).Info("MiningState.searchingNonce: found valid nonce, transit to MintedState.")

					// seal block.
					miningBlock.Seal()

					state.p.TransitByKey(Minted, nil)

					// break this for loop.
					break computeHash
				}
			}
		}
	}

	// wait for quit while transiting.
	<-state.quitCh
	log.Info("MiningState.searchingNonce: quit at the end.")
}
