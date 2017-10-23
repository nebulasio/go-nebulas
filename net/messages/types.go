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
	"errors"
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/net"
	netpb "github.com/nebulasio/go-nebulas/net/pb"
)

// BaseMessage base message
type BaseMessage struct {
	t    string
	data interface{}
}

// HelloMessage use to send hello
type HelloMessage struct {
	NodeID        string
	ClientVersion string
}

// NewHelloMessage new hello message
func NewHelloMessage(nodeID string, clientVersion string) *HelloMessage {
	return &HelloMessage{NodeID: nodeID, ClientVersion: clientVersion}
}

// ToProto converts domain HelloMessage to proto HelloMessage
func (h *HelloMessage) ToProto() (proto.Message, error) {
	return &netpb.Hello{
		NodeID:        h.NodeID,
		ClientVersion: h.ClientVersion,
	}, nil
}

// FromProto converts proto HelloMessage to domain HelloMessage
func (h *HelloMessage) FromProto(msg proto.Message) error {
	if msg, ok := msg.(*netpb.Hello); ok {
		h.NodeID = msg.NodeID
		h.ClientVersion = msg.ClientVersion
		return nil
	}
	return errors.New("Pb Message cannot be converted into HelloMessage")
}

// NewBaseMessage new base message
func NewBaseMessage(t string, data interface{}) net.Message {
	return &BaseMessage{t: t, data: data}
}

// MessageType get message type
func (msg *BaseMessage) MessageType() string {
	return msg.t
}

// Data get the message data
func (msg *BaseMessage) Data() interface{} {
	return msg.data
}

// String get the message to string
func (msg *BaseMessage) String() string {
	return fmt.Sprintf("BaseMessage {type:%s; data:%s}",
		msg.t,
		msg.data,
	)
}
