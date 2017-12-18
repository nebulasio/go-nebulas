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
	peer "github.com/libp2p/go-libp2p-peer"
	"github.com/nebulasio/go-nebulas/net"
	netpb "github.com/nebulasio/go-nebulas/net/pb"
)

// BaseMessage base message
type BaseMessage struct {
	t    string
	from string
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
		NodeId:        h.NodeID,
		ClientVersion: h.ClientVersion,
	}, nil
}

// FromProto converts proto HelloMessage to domain HelloMessage
func (h *HelloMessage) FromProto(msg proto.Message) error {
	if msg, ok := msg.(*netpb.Hello); ok {
		h.NodeID = msg.NodeId
		h.ClientVersion = msg.ClientVersion
		return nil
	}
	return errors.New("Pb Message cannot be converted into HelloMessage")
}

// Peers struct
type Peers struct {
	peers []*PeerInfo
}

// PeerInfo peerInfo struct
type PeerInfo struct {
	id    peer.ID
	addrs []string
}

// Peers return peers
func (ps *Peers) Peers() []*PeerInfo {
	return ps.peers
}

// ID return peer`s id
func (p *PeerInfo) ID() peer.ID {
	return p.id
}

// Addrs return peer`s addrs
func (p *PeerInfo) Addrs() []string {
	return p.addrs
}

// NewPeerInfoMessage return a peerInfo instance
func NewPeerInfoMessage(id peer.ID, addrs []string) *PeerInfo {
	return &PeerInfo{id, addrs}
}

// ToProto converts domain PeerInfo to proto PeerInfo
func (p *PeerInfo) ToProto() (proto.Message, error) {
	return &netpb.PeerInfo{
		Id:    string(p.id),
		Addrs: p.addrs,
	}, nil
}

// FromProto converts proto PeerInfo to domain PeerInfo
func (p *PeerInfo) FromProto(msg proto.Message) error {
	if msg, ok := msg.(*netpb.PeerInfo); ok {
		p.id = peer.ID(msg.Id)
		p.addrs = msg.Addrs
		return nil
	}
	return errors.New("Pb Message cannot be converted into PeerInfo")
}

// NewPeersMessage return peers instance
func NewPeersMessage(peers []*PeerInfo) *Peers {
	return &Peers{peers}
}

// ToProto converts domain Peers to proto Peers
func (ps *Peers) ToProto() (proto.Message, error) {
	var result []*netpb.PeerInfo
	for _, v := range ps.peers {
		peer, err := v.ToProto()
		if err != nil {
			return nil, err
		}
		if peer, ok := peer.(*netpb.PeerInfo); ok {
			result = append(result, peer)
		} else {
			return nil, errors.New("Pb Message cannot be converted into Peers")
		}
	}
	return &netpb.Peers{
		Peers: result,
	}, nil
}

// FromProto converts proto Peers to domain Peers
func (ps *Peers) FromProto(msg proto.Message) error {
	if msg, ok := msg.(*netpb.Peers); ok {
		for _, v := range msg.Peers {
			peer := new(PeerInfo)
			if err := peer.FromProto(v); err != nil {
				return err
			}
			ps.peers = append(ps.peers, peer)
		}
		return nil
	}
	return errors.New("Pb Message cannot be converted into Peers")
}

// NewBaseMessage new base message
func NewBaseMessage(t string, from string, data interface{}) net.Message {
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
