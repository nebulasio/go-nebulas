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
	libnet "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/nebulasio/go-nebulas/net"
	"github.com/nebulasio/go-nebulas/net/messages"
	byteutils "github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/nebulasio/go-nebulas/util/logging"
	metrics "github.com/rcrowley/go-metrics"
	"github.com/sirupsen/logrus"
)

// connection state
const (
	ProtocolID    = "/neb/1.0.0"
	SNC           = -1
	SHandshaking  = 0
	SOK           = 1
	ClientVersion = "0.2.0"
)

var (
	packetsIn  = metrics.GetOrRegisterMeter("neb.net.packets.in", nil)
	netBytesIn = metrics.GetOrRegisterMeter("neb.net.bytes.in", nil)
)

// MagicNumber the protocol magic number, A constant numerical or text value used to identify protocol.
var MagicNumber = []byte{0x4e, 0x45, 0x42, 0x31}

func (node *Node) messageHandler(s libnet.Stream) {
	var tmpMsg *NebMessage
	var dataLength uint32

	streamBuffer := []byte{}
	sdata := make([]byte, 1024)

	pid := s.Conn().RemotePeer()
	addrs := s.Conn().RemoteMultiaddr()
	key := pid.Pretty()

	for {
		select {
		case <-node.netService.quitCh:
			return
		default:
			n, err := s.Read(sdata)
			if err != nil {
				logging.VLog().WithFields(logrus.Fields{
					"count": n,
					"err":   err,
					"addrs": addrs,
				}).Warn("Read EOF.")
				// node.Bye(pid, []ma.Multiaddr{addrs}, s, key)
				// return
				return
			}
			streamBuffer = append(streamBuffer, sdata[:n]...)

			if tmpMsg == nil {
				// wait to parseHeader
				if len(streamBuffer) < offsetData {
					continue
				}
				tmpMsg, err = node.parseMsgHeader(streamBuffer)
				if err != nil {
					logging.VLog().WithFields(logrus.Fields{
						"addrs": addrs.String(),
						"err":   err,
					}).Error("parse header error")
					node.Bye(pid, []ma.Multiaddr{addrs}, s, key)
					return
				}

				streamBuffer = streamBuffer[offsetData:]
				dataLength = byteutils.Uint32(tmpMsg.dataLength)
			}

			if dataLength > uint32(len(streamBuffer)) {
				// stream data is not enough
				continue
			}

			if err = node.parseMsgData(tmpMsg, streamBuffer); err != nil {
				logging.VLog().WithFields(logrus.Fields{
					"addrs": addrs.String(),
					"err":   err,
				}).Error("parse data error")
				node.Bye(pid, []ma.Multiaddr{addrs}, s, key)
				return
			}
			streamBuffer = streamBuffer[dataLength:]

			msg := tmpMsg
			tmpMsg = nil
			dataLength = 0

			packetsIn.Mark(1)
			netBytesIn.Mark(int64(byteutils.Uint32(msg.dataLength) + uint32(offsetData)))

			switch msg.msgName {
			case HELLO:
				node.handleHelloMsg(msg.data, pid, s, addrs, key)
			case OK:
				node.handleOkMsg(msg.data, pid, s, addrs, key)
			case BYE:

			case SyncRoute:
				node.handleSyncRouteMsg(msg.data, pid, s, addrs, key)
			case SyncRouteReply:
				node.handleSyncRouteReplyMsg(msg.data, pid, s, addrs)
			case NewHashMsg:
				node.handleNewHashMsg(msg.data, pid)
			case NetworkID:
				node.handleNetworkIDMsg(msg.data, pid, s)
			case NetworkIDReply:
				node.handleReNetworkIDMsg(msg.data, pid)
			default:
				var relayness []peer.ID
				logging.VLog().WithFields(logrus.Fields{
					"msgName": msg.msgName,
					"pid":     pid.Pretty(),
				}).Info("receive block & tx message.")

				m, ok := net.PacketsInByTypes.Load(msg.msgName)
				if ok {
					m.(metrics.Meter).Mark(1)
				}

				node.netService.PutMessage(messages.NewBaseMessage(msg.msgName, pid.Pretty(), msg.data))

				peers, exists := node.relayness.Get(byteutils.Uint32(msg.dataChecksum))
				if exists {
					relayness = peers.([]peer.ID)
				}
				node.relayness.Add(byteutils.Uint32(msg.dataChecksum), append(relayness, pid))
			}

		}
	}

}
