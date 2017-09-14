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
	"encoding/binary"
	"fmt"
	"time"

	"encoding/hex"

	"github.com/nebulasio/go-nebulas/consensus"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/sha3"
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
func (state *MiningState) Event(e consensus.Event) consensus.State {
	log.WithFields(log.Fields{
		"stateType": fmt.Sprintf("%T", state),
		"eventType": fmt.Sprintf("%T", e),
	}).Warn("ignore this event.")
	return state
}

// Enter called when transiting to this state.
func (state *MiningState) Enter(data interface{}) {
	log.Info("MiningState enter.")
	go state.calculateHash()
}

// Leave called when leaving this state.
func (state *MiningState) Leave(data interface{}) {
	log.Info("MiningState leave.")
}

func (state *MiningState) calculateHash() {
	// calculate hash.
	newBlock := state.p.newBlock
	nonce := newBlock.Nonce()
	parentHash := newBlock.ParentHash()

	parentHashBytes := []byte(parentHash)
	nonceBytes := make([]byte, 8)
	resultBytes := make([]byte, 32)

	// expectedHashBytes := make([]byte, 4)

	nonceHash := sha3.New256()

	timeStart := time.Now()
	miningInterval, _ := time.ParseDuration("1s")

	for {
		select {
		case <-state.quitCh:
			log.Info("quit MiningState.")
			return
		default:
			nonce++
			binary.LittleEndian.PutUint64(nonceBytes, nonce)

			// compute hash..
			nonceHash.Reset()
			nonceHash.Write(parentHashBytes)
			nonceHash.Write(nonceBytes)
			nonceHash.Sum(resultBytes[:0])

			// verify.
			// if bytes.Compare(expectedHashBytes[:4], resultBytes[:4]) == 0 {
			if true {
				log.WithFields(log.Fields{
					"parentHash": parentHash,
					"nonce":      nonce,
					"hashResult": hex.EncodeToString(resultBytes),
				}).Info("found, done")

				elapse := time.Since(timeStart)
				if elapse < miningInterval {
					time.Sleep(miningInterval - elapse)
				}

				newBlock.SetNonce(nonce)
				state.p.Transite(state.p.states[Minted], nil)

				return
			}

			log.WithFields(log.Fields{
				"parentHash": parentHash,
				"nonce":      nonce,
				"hashResult": hex.EncodeToString(resultBytes),
			}).Info("not found, continue")

		}
	}

}
