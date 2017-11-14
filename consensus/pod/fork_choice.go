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
	log "github.com/sirupsen/logrus"
)

// ForkChoice Rule of PoD Consensus
// 1. Choose the chain with highest score
// 2. if same score, choose the longest chain
// 3. if same length, choose the chain whose tail's hash is smaller
func (p *PoD) ForkChoice() {
	log.WithFields(log.Fields{
		"func":         "PoD.ForkChoice",
		"current tail": p.chain.TailBlock(),
	}).Info("Fork Choice.")
}
