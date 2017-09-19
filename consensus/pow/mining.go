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
	"time"

	"github.com/nebulasio/go-nebulas/consensus"
	"github.com/nebulasio/go-nebulas/crypto/hash"
	"github.com/nebulasio/go-nebulas/utils/byteutils"

	log "github.com/sirupsen/logrus"
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
	return false, nil
}

// Enter called when transiting to this state.
func (state *MiningState) Enter(data interface{}) {
	log.Debug("MiningState enter.")
	go state.searchingNonce()
}

// Leave called when leaving this state.
func (state *MiningState) Leave(data interface{}) {
	log.Debug("MiningState leave.")
}

func (state *MiningState) searchingNonce() {
	// calculate hash.
	newBlock := state.p.newBlock
	nonce := newBlock.Nonce()

	parentHash := newBlock.ParentHash()

	timeStart := time.Now()
	miningInterval, _ := time.ParseDuration("1s")

	for {
		select {
		case <-state.quitCh:
			log.Info("quit MiningState.")
			return

		default:
			nonce++

			// compute hash..
			resultBytes := hash.Sha256(parentHash, byteutils.FromUint64(nonce))

			// verify.
			if resultBytes[0] == 0 && resultBytes[1] == 0 {
				log.WithFields(log.Fields{
					"nonce":      nonce,
					"parentHash": byteutils.Hex(parentHash),
					"hashResult": byteutils.Hex(resultBytes),
				}).Info("Nonce found, done")

				// FIXME: Debug purpose.
				elapse := time.Since(timeStart)
				if elapse < miningInterval {
					time.Sleep(miningInterval - elapse)
				}

				newBlock.SetNonce(nonce)
				state.p.Transite(state.p.states[Minted], nil)

				return
			}
		}
	}

}
