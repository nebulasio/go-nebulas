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
	"github.com/nebulasio/go-nebulas/core"
	log "github.com/sirupsen/logrus"
)

// ForkChoice Rule of PoD Consensus
// 1. Choose the chain with highest score
// 2. if same score, choose the longest chain
// 3. if same length, choose the chain whose tail's HashInt64 is bigger
func (p *PoD) ForkChoice(newTail *core.Block) {
	oldTail := p.chain.TailBlock()
	log.WithFields(log.Fields{
		"func":     "PoD.ForkChoice",
		"old tail": oldTail,
	}).Info("Fork Choice.")
	oldTailScores, err := core.CalScoresOnChain(oldTail)
	if err != nil {
		log.WithFields(log.Fields{
			"func": "PoD.ForkChoice",
			"err":  err,
		}).Error("Error. Calculate Old Tail Scores.")
		return
	}
	newTailScores, err := core.CalScoresOnChain(newTail)
	if err != nil {
		log.WithFields(log.Fields{
			"func": "PoD.ForkChoice",
			"err":  err,
		}).Error("Error. Calculate New Tail Scores.")
		return
	}
	// compare scores
	if newTailScores < oldTailScores {
		return
	}
	if newTailScores > oldTailScores {
		log.WithFields(log.Fields{
			"func":          "PoD.ForkChoice",
			"newTailScores": newTailScores,
			"oldTailScores": oldTailScores,
		}).Info("Bigger Scores. Found New Tail.")
		p.chain.SetTailBlock(newTail)
		return
	}
	// compare height
	if newTail.Height() < oldTail.Height() {
		return
	}
	if newTail.Height() > oldTail.Height() {
		log.WithFields(log.Fields{
			"func":          "PoD.ForkChoice",
			"newTailHeight": newTail.Height(),
			"oldTailHeight": oldTail.Height(),
		}).Info("Longer Chain. Found New Tail.")
		p.chain.SetTailBlock(newTail)
		return
	}
	// compare tail hash
	oldHashInt64 := core.HashBytesToInt64(oldTail.Hash(), oldTail.Hash())
	newHashInt64 := core.HashBytesToInt64(newTail.Hash(), newTail.Hash())
	if newHashInt64 > oldHashInt64 {
		log.WithFields(log.Fields{
			"func":        "PoD.ForkChoice",
			"newTailHash": newHashInt64,
			"oldTailHash": oldHashInt64,
		}).Info("Bigger Hash. Found New Tail.")
		p.chain.SetTailBlock(newTail)
	}
}
