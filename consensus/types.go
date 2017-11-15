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

package consensus

import "github.com/nebulasio/go-nebulas/core"

// EventType list
const (
	NewBlockEvent  = "event.newblock"
	CanMiningEvent = "event.canmining"
)

// Consensus interface of consensus algorithm.
type Consensus interface {
	Start()
	Stop()

	CanMining() bool
	SetCanMining(bool)

	Transit(from, to State, data interface{})

	VerifyBlock(*core.Block) error
}

// MessageType
const (
	MessageTypeNewTx = "newtx"
)
