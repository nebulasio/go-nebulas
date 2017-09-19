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
	"github.com/nebulasio/go-nebulas/core"
	log "github.com/sirupsen/logrus"
)

func (p *Pow) ForkChoice() {
	bc := p.chain
	bkPool := bc.BlockPool()

	log.Info("Pow.ForkChoice: process received block.")
	// TODO: append received blocks to BlockChain.
	blocks := make(map[core.HexHash]*core.Block)

	bkPool.Range(func(key, value interface{}) bool {
		k := key.(core.HexHash)
		v := value.(*core.Block)
		blocks[k] = v
		return true
	})

	bc.PutUnattachedBlockMap(blocks)
	log.Debug("Pow.ForkChoice: all received blocks is ", blocks)

	// link parent.
	log.Debug("Pow.ForkChoice: 1st iteration, link parenet.")
	for k, v := range blocks {
		// get parent block.
		parentBlock := bc.GetBlock(v.ParentHash())
		if parentBlock == nil {
			// waiting for network sync.
			log.Debugf("Pow.ForkChoice: block=%s requires parent from network.", k)
			continue
		}

		v.LinkParentBlock(parentBlock)
		log.Debugf("Pow.ForkChoice: block=%s link to parentBlock=%s.", k, v.ParentHash().Hex())

		// remove block from BlockPool.
		bkPool.Delete(k)
	}

	// find the max depth.
	log.Debug("Pow.ForkChoice: 2nd iteration, calculate depth.")
	highest := uint64(0)
	var block *core.Block
	for _, v := range blocks {
		h := v.Height()
		if h > highest {
			highest = h
			block = v
		}
	}

	if highest > 0 {
		// set longest chain.
		log.WithFields(log.Fields{
			"highest": highest,
			"Block":   block,
		}).Info("Pow.ForkChoice: found the longest chain, set tail block.")
		bc.SetTailBlock(block)
	} else {
		log.Debug("Pow.ForkChoice: can't find the longest chain, return.")
	}

	// dump chain.
	log.Debug("Dump: ", bc.Dump())
}
