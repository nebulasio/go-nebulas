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
// TODO(leon): this interface should be in net package.
type Manager interface {
	Start() error
	Stop()

	Node() *Node

	SendSyncReply(string, net.Serializable)

	Register(...*net.Subscriber)
	Deregister(...*net.Subscriber)

	Broadcast(string, net.Serializable)
	Relay(string, net.Serializable)
	SendMsg(string, []byte, string, int) error

	BroadcastNetworkID([]byte)

	BuildRawMessageData([]byte, string) []byte
}
