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

package messages

import (
	"fmt"

	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/net"
)

const (
	NewBlockMessageType = "NewBlock"
)

type BlockMessage struct {
	t     net.MessageType
	block *core.Block
}

func NewBlockMessage(block *core.Block) *BlockMessage {
	msg := &BlockMessage{t: NewBlockMessageType, block: block}
	return msg
}

func (msg *BlockMessage) MessageType() net.MessageType {
	return msg.t
}

func (msg *BlockMessage) Block() *core.Block {
	return msg.block
}

func (msg *BlockMessage) String() string {
	return fmt.Sprintf("BlockMessage {type:%s; block:%s}",
		msg.t,
		msg.block,
	)
}
