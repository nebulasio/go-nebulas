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

	"github.com/nebulasio/go-nebulas/crypto/hash"
	"github.com/nebulasio/go-nebulas/util/byteutils"

	"github.com/nebulasio/go-nebulas/net"
)

// BaseMessage base message
type BaseMessage struct {
	t    string
	from string
	data []byte
}

// NewBaseMessage new base message
func NewBaseMessage(t string, from string, data []byte) net.Message {
	return &BaseMessage{t: t, from: from, data: data}
}

// MessageType get message type
func (msg *BaseMessage) MessageType() string {
	return msg.t
}

// MessageFrom get message who send
func (msg *BaseMessage) MessageFrom() string {
	return msg.from
}

// Data get the message data
func (msg *BaseMessage) Data() []byte {
	return msg.data
}

// Hash return the message hash
func (msg *BaseMessage) Hash() string {
	return byteutils.Hex(hash.Sha3256(msg.data))
}

// String get the message to string
func (msg *BaseMessage) String() string {
	return fmt.Sprintf("BaseMessage {type:%s; data:%s; from:%s}",
		msg.t,
		msg.data,
		msg.from,
	)
}
