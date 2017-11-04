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
	log "github.com/sirupsen/logrus"
)

// ForkChoice Rule of PoW Consensus
// Choose the longest chain
func (p *Pow) ForkChoice() {
	bc := p.chain
	tailBlock := bc.TailBlock()
	detachedTailBlocks := bc.DetachedTailBlocks()

	// find the max depth.
	log.WithFields(log.Fields{
		"func": "Pow.ForkChoice",
	}).Debug("find the highest tail.")

	newTailBlock := tailBlock
	maxHeight := tailBlock.Height()

	for _, v := range detachedTailBlocks {
		h := v.Height()
		if h > maxHeight {
			maxHeight = h
			newTailBlock = v
		}
		// TODO(@roy): remove unused tail from detachedTails.
	}

	if newTailBlock == bc.TailBlock() {
		log.WithFields(log.Fields{
			"func": "Pow.ForkChoice",
		}).Info("current tail is the highest, no change.")
	} else {
		log.WithFields(log.Fields{
			"func":      "Pow.ForkChoice",
			"maxHeight": maxHeight,
			"tailBlock": newTailBlock,
		}).Info("change to new tail.")
		bc.SetTailBlock(newTailBlock)
	}

	// dump chain.
	// log.Debug("Dump: ", bc.Dump())
}
