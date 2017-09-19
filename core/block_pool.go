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

package core

import (
	"sync"

	"github.com/nebulasio/go-nebulas/components/net"
)

// BlockPool a pool of all received blocks from network.
// Blocks will be sent to Consensus when it passes signature verification.
type BlockPool struct {
	inner           *sync.Map
	receivedBlockCh chan net.Message

	quitCh chan int
}

// NewBlockPool return new #BlockPool instance.
func NewBlockPool() *BlockPool {
	bp := &BlockPool{
		inner:           new(sync.Map),
		receivedBlockCh: make(chan net.Message, 128),
	}
	return bp
}

// ReceivedBlockCh return received block chan.
func (pool *BlockPool) ReceivedBlockCh() chan net.Message {
	return pool.receivedBlockCh
}

// RegisterInNetwork register message subscriber in network.
func (pool *BlockPool) RegisterInNetwork(nm *net.Manager) {
	nm.Register(net.NewSubscriber(pool, pool.ReceivedBlockCh(), net.MessageTypeNewBlock))
}
