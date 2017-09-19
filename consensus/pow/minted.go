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
	"github.com/nebulasio/go-nebulas/utils/byteutils"
	log "github.com/sirupsen/logrus"
)

const (
	// Minted minted state key
	Minted = "minted"
)

// MintedState minted state, transite from @MiningState
type MintedState struct {
	p *Pow
}

// NewMintedState create MintedState instance.
func NewMintedState(p *Pow) *MintedState {
	state := &MintedState{p: p}
	return state
}

// Event handle event.
func (state *MintedState) Event(e consensus.Event) (bool, consensus.State) {
	return false, nil
}

// Enter called when transiting to this state.
func (state *MintedState) Enter(data interface{}) {
	log.Debug("MintedState.Enter: enter.")

	p := state.p
	bc := p.chain
	bkPool := bc.BlockPool()
	headBlock := bc.TailBlock()

	// process minted block.
	if state.p.miningBlock.Nonce() > 0 {

		log.Info("MintedState.Enter: process minted block.")

		state.p.miningBlock.Sign()

		// send new block to network.
		state.p.nm.BroadcastBlock(state.p.miningBlock)

		bkPool.AddLocalBlock(state.p.miningBlock)
		if state.p.receivedBlock == nil {
			state.p.receivedBlock = state.p.miningBlock
		}
	}

	// process the received block.
	if state.p.receivedBlock != nil {
		state.p.receivedBlock = nil

		log.Info("MintedState.Enter: process received block.")
		// TODO: append received blocks to BlockChain.
		blocks := make(map[string]*core.Block)

		bkPool.Range(func(key, value interface{}) bool {
			k := key.(string)
			v := value.(*core.Block)
			blocks[k] = v
			return true
		})
		log.Debug("MintedState.Enter: all received blocks is ", blocks)

		// link parent.
		log.Debug("MintedState.Enter: 1st iteration, link parenet.")
		for k, v := range blocks {
			if v.LinkParentBlock(headBlock) == true {
				continue
			}

			parentBlockHashKey := byteutils.Hex(v.ParentHash())
			parentBlock := blocks[parentBlockHashKey]

			if parentBlock == nil {
				// waiting for network sync.
				log.Debugf("MintedState.Enter: block=%s requires parent from network.", k)
				continue
			}

			v.LinkParentBlock(parentBlock)
			log.Debugf("MintedState.Enter: block=%s link to parentBlock=%s.", k, parentBlockHashKey)
		}

		// find the max depth.
		log.Debug("MintedState.Enter: 2nd iteration, calculate depth.")
		highest := uint64(0)
		var block *core.Block
		for _, v := range blocks {
			h := v.Height()
			if h > highest {
				highest = h
				block = v
			}
		}

		// set longest chain.
		log.Debug("MintedState.Enter: final set longest chain.")
		log.WithFields(log.Fields{
			"highest": highest,
			"Block":   block,
		}).Info("MintedState.Enter: found the longest chain.")
		bc.SetTailBlock(block)

		// dump chain.
		log.Info("Dump: ", bc.Dump())
	}

	// move to prepare state.
	state.p.TransiteByKey(Prepare, nil)
}

// Leave called when leaving this state.
func (state *MintedState) Leave(data interface{}) {
	log.Debug("MintedState.Leave: leave.")
}
