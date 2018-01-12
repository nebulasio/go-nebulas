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

package p2p

import "github.com/nebulasio/go-nebulas/net"

// Manager manager interface
// TODO: @robin move Manager to net and rename to NetService.
type Manager interface {
	Start() error
	Stop()

	Node() *Node

	Register(...*net.Subscriber)
	Deregister(...*net.Subscriber)

	Broadcast(string, net.Serializable, int)
	Relay(string, net.Serializable, int)
	SendMsg(string, []byte, string, int) error

	SendMessageToPeers(messageName string, data []byte, priority int, filter net.PeerFilterAlgorithm) int
	SendMessageToPeer(messageName string, data []byte, priority int, peerID string) error

	ClosePeer(peerID string, reason error)

	BroadcastNetworkID([]byte)

	BuildRawMessageData([]byte, string) []byte
}
