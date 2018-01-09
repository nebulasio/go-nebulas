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

import (
	"errors"

	libnet "github.com/libp2p/go-libp2p-net"
	"github.com/nebulasio/go-nebulas/net"
	"github.com/nebulasio/go-nebulas/util/logging"
	metrics "github.com/rcrowley/go-metrics"
	"github.com/sirupsen/logrus"
)

// Error types
var (
	ErrNotInSameNetWork = errors.New("target address is not in the same network")
	ErrStreamNotExist   = errors.New("stream is not exist")
)

var (
	packetsOut  = metrics.GetOrRegisterMeter("neb.net.packets.out", nil)
	netBytesOut = metrics.GetOrRegisterMeter("neb.net.bytes.out", nil)
)

// SendMsg send message to a peer
func (node *Node) sendMsgWithStream(msgName string, msg []byte, stream libnet.Stream) error {

	totalData := node.buildData(msg, msgName)

	if err := Write(stream, totalData); err != nil {
		return err
	}
	packetsOut.Mark(1)
	m, ok := net.PacketsOutByTypes.Load(msgName)
	if ok {
		m.(metrics.Meter).Mark(1)
	}
	netBytesOut.Mark(int64(len(totalData)))

	logging.VLog().WithFields(logrus.Fields{
		"msgName": msgName,
	}).Info("send message to a peer")
	return nil
}

// SendMsg send message to a peer
func (node *Node) sendMsg(msgName string, msg []byte, target string) error {
	if msgName != NetworkID && !node.checkNetworkID(target) {
		return ErrNotInSameNetWork
	}
	streamStore, ok := node.stream.Load(target)
	if !ok {
		return ErrStreamNotExist
	}
	return node.sendMsgWithStream(msgName, msg, streamStore.(*StreamStore).stream)
}
