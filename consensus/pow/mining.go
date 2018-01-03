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
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	metrics "github.com/rcrowley/go-metrics"

	"time"

	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

const (
	// Mining mining state key
	Mining = "mining"
)

var searchingNonceTimer = metrics.GetOrRegisterTimer("pow_search_nonce", nil)
var nonceRetryCountGauge = metrics.GetOrRegisterGauge("pow_nonce_retry_count", nil)

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

func (state *MiningState) String() string {
	return fmt.Sprintf("MiningState %p", state)
}

// Event handle event.
func (state *MiningState) Event(e consensus.Event) (bool, consensus.State) {
	switch e.EventType() {
	case consensus.NewBlockEvent:
		return true, NewMintedState(state.p)
	default:
		return false, nil
	}
}

// Enter called when transiting to this state.
func (state *MiningState) Enter(data interface{}) {
	logging.VLog().Info("MiningState.Enter: enter.")

	// start searching nonce.
	go state.searchingNonce(state.p.miningBlock)
}

// Leave called when leaving this state.
func (state *MiningState) Leave(data interface{}) {
	logging.VLog().Info("MiningState.Leave: leave.")
	state.quitCh <- true
}

func (state *MiningState) searchingNonce(miningBlock *core.Block) {
	// transit to MintedState if newBlockReceived is true.
	if state.p.newBlockReceived {
		logging.VLog().Info("MiningState.Enter: new block received, transit to MintedState.")
		state.p.Transit(state, NewMintedState(state.p), nil)

	} else {
		nonce := miningBlock.Nonce()
		parentHash := miningBlock.ParentHash()

		now := time.Now()
	computeHash:
		for {
			select {
			case <-state.quitCh:
				logging.VLog().Info("MiningState.searchingNonce: quit.")
				return

			default:
				nonce++
				// compute hash..
				miningBlock.SetNonce(nonce)
				miningBlock.SetTimestamp(time.Now().Unix())
				resultBytes := HashAndVerify(miningBlock)

				if resultBytes != nil {
					logging.VLog().WithFields(logrus.Fields{
						"nonce":      nonce,
						"parentHash": byteutils.Hex(parentHash),
						"hashResult": byteutils.Hex(resultBytes),
					}).Info("MiningState.searchingNonce: found valid nonce, transit to MintedState.")

					// seal block.
					miningBlock.Seal()

					searchingNonceTimer.UpdateSince(now)
					nonceRetryCountGauge.Update(int64(nonce))

					state.p.Transit(state, NewMintedState(state.p), nil)

					// break this for loop.
					break computeHash
				}
			}
		}
	}

	// wait for quit while transiting.
	<-state.quitCh
	logging.VLog().Info("MiningState.searchingNonce: quit at the end.")
}
