// Copyright (C) 2018 go-nebulas authors
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
	"fmt"
	"net"
	"time"

	"github.com/gogo/protobuf/proto"
	libnet "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/nebulasio/go-nebulas/net/messages"
	netpb "github.com/nebulasio/go-nebulas/net/pb"
	byteutils "github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/nebulasio/go-nebulas/util/logging"
	metrics "github.com/rcrowley/go-metrics"
	"github.com/sirupsen/logrus"
)

const (
	NebProtocolID = "/neb/1.0.0"

	HELLO          = "hello"
	OK             = "ok"
	BYE            = "bye"
	SyncRoute      = "syncroute"
	SyncRouteReply = "resyncroute"

	ClientVersion = "0.2.0"
)

var (
	ErrShouldCloseConnectionAndExitLoop = errors.New("should close connection and exit loop")
)

type Stream struct {
	pid         peer.ID
	addr        ma.Multiaddr
	stream      libnet.Stream
	node        *Node
	connectedAt int64
}

// NewStream return a new StreamInfo
func NewStream(pid peer.ID, stream libnet.Stream, node *Node) *Stream {
	return &Stream{
		pid:         pid,
		addr:        stream.Conn().RemoteMultiaddr(),
		stream:      stream,
		node:        node,
		connectedAt: time.Now().Unix(),
	}
}

func (s *Stream) Connect() error {
	// connect to host.
	stream, err := s.node.host.NewStream(
		s.node.context,
		s.pid,
		NebProtocolID,
	)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"stream": s.String(),
		}).Debug("Failed to connect to host.")
		return err
	}
	s.stream = stream

	return nil
}

func (s *Stream) IsConnected() bool {
	return s.stream != nil
}

func (s *Stream) String() string {
	return fmt.Sprintf("Peer Stream: %s , %s", s.pid.Pretty(), s.addr.String())
}

func (s *Stream) SendProtoMessage(messageName string, pb proto.Message) error {
	data, err := proto.Marshal(pb)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err":         err,
			"messageName": messageName,
			"stream":      s.String(),
		}).Warn("Failed to marshal proto message.")
		return err
	}

	return s.SendMessage(messageName, data)
}

func (s *Stream) SendMessage(messageName string, data []byte) error {
	message, err := NewNebMessage(s.node.config.ChainID, DefaultReserved, 0, messageName, data)
	if err != nil {
		return err
	}

	// metrics.
	metricsPacketsOutByMessageName(messageName, data)

	return s.Send(message.Content())
}

func (s *Stream) Send(data []byte) error {
	n, err := s.stream.Write(data)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err":    err,
			"stream": s.String(),
		}).Warn("Failed to send message to peer.")
		s.Close()
		return err
	}

	metricsPacketsOut.Mark(1)
	metricsBytesOut.Mark(int64(n))
}

// StartLoop start stream handling loop.
func (s *Stream) StartLoop() {
	go func() {
		buf := make([]byte, 1024*4)
		messageBuffer := []byte{}

		pid := s.pid
		addr := s.addr

		var message *NebMessage

		// send Hello to host if stream is not connected.
		if !s.IsConnected() {
			if err := s.Connect(); err != nil {
				return
			}
			if err := s.Hello(); err != nil {
				return
			}
		}

		// loop.
		for {
			select {
			default:
				n, err := s.stream.Read(buf)
				if err != nil {
					logging.VLog().WithFields(logrus.Fields{
						"err":     err,
						"pid":     pid.Pretty(),
						"address": addr,
					}).Error("Error occurred when reading data from network connection.")
					s.Close()
					return
				}

				messageBuffer = append(messageBuffer, buf[:n]...)

				if message == nil {
					// waiting for header data.
					if len(messageBuffer) < NebMessageHeaderLength {
						// continue reading.
						continue
					}

					message, err = ParseNebMessage(messageBuffer)
					if err != nil {
						s.Bye()
						return
					}

					// check ChainID.
					if s.node.config.ChainID != message.ChainID() {
						logging.VLog().WithFields(logrus.Fields{
							"err":             err,
							"pid":             pid.Pretty(),
							"address":         addr,
							"conf.chainID":    s.node.config.ChainID,
							"message.chainID": message.ChainID(),
						}).Error("Invalid chainID, disconnect the connection.")
						s.Bye()
						return
					}

					// remove header from buffer.
					messageBuffer = messageBuffer[NebMessageHeaderLength:]
				}

				// waiting for data.
				if uint32(len(messageBuffer)) < message.DataLength() {
					// continue reading.
					continue
				}

				if err = message.ParseMessageData(messageBuffer); err != nil {
					s.Bye()
					return
				}

				// remove data from buffer.
				messageBuffer = messageBuffer[message.DataLength():]

				// metrics.
				packetsIn.Mark(1)
				netBytesIn.Mark(message.Length())
				metricsPacketsInByMessageName(message.MessageName(), message.Length())

				// handle message.
				if err := s.handleMessage(message); err == ErrShouldCloseConnectionAndExitLoop {
					s.Bye()
					return
				}
			}
		}

	}()
}

func (s *Stream) handleMessage(message *NebMessage) error {
	switch message.MessageName() {
	case HELLO:
		return s.onHello(message)
	case OK:
		return s.onOk(message)
	case BYE:
		return s.onBye(message)
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

		streamStore, ok := node.stream.Load(key)
		if !ok {
			node.Bye(pid, []ma.Multiaddr{addrs}, s, key)
			return
		}
		if streamStore.(*StreamStore).conn != SOK {
			logging.VLog().Error("peer not shake hand before send message.")
			node.Bye(pid, []ma.Multiaddr{addrs}, s, key)
			return
		}
		node.netService.PutMessage(messages.NewBaseMessage(msg.msgName, pid.Pretty(), msg.data))

		peers, exists := node.relayness.Get(byteutils.Uint32(msg.dataChecksum))
		if exists {
			relayness = peers.([]peer.ID)
		}
		node.relayness.Add(byteutils.Uint32(msg.dataChecksum), append(relayness, pid))
	}
}

func (s *Stream) Close() {
	s.stream.Close()
	s.node.streamManager.Remove(s)
	s.node.routeTable.RemovePeerStream(s)
	s.stream = nil
}

func (s *Stream) Bye() {
	s.SendMessage(BYE, []byte{})
	s.Close()
}

func (s *Stream) Hello() error {
	msg := &netpb.Hello{
		NodeId:        s.node.id.String(),
		ClientVersion: ClientVersion,
	}
	return s.SendProtoMessage(HELLO, msg)
}

func (s *Stream) onBye(message *NebMessage) error {
	logging.VLog().WithFields(logrus.Fields{
		"stream": s.String(),
	}).Debug("Received Bye message, close the connection.")
	return ErrShouldCloseConnectionAndExitLoop
}

func (s *Stream) onHello(message *NebMessage) bool {
	msg, err := netpb.HelloMessageFromProto(message.data)
	if err != nil {
		return ErrShouldCloseConnectionAndExitLoop
	}

	if msg.NodeID != s.pid.String() || !CheckClientVersionCompability(clientVersion, msg.ClientVersion) {
		// invalid client, bye().
		logging.VLog().WithFields(logrus.Fields{
			"pid":               s.pid.Pretty(),
			"address":           s.addr,
			"ok.node_id":        msg.NodeId,
			"ok.client_version": msg.ClientVersion,
		}).Debug("Invalid NodeId or incompatible client version.")
		return ErrShouldCloseConnectionAndExitLoop
	}

	// add to route table.
	s.node.routeTable.AddPeerStream(s)

	// send OK response.
	resp := &netpb.OK{
		NodeId:        s.node.id.String(),
		ClientVersion: ClientVersion,
	}

	return s.SendProtoMessage(OK, resp)
}

func (s *Stream) onOk(message *NebMessage) {
	msg, err := netpb.OKMessageFromProto(message.data)
	if err != nil {
		return ErrShouldCloseConnectionAndExitLoop
	}

	if msg.NodeId != s.pid.String() || !CheckClientVersionCompability(clientVersion, msg.ClientVersion) {
		// invalid client, bye().
		logging.VLog().WithFields(logrus.Fields{
			"pid":               s.pid.Pretty(),
			"address":           s.addr,
			"ok.node_id":        msg.NodeId,
			"ok.client_version": msg.ClientVersion,
		}).Debug("Invalid NodeId or incompatible client version.")
		return ErrShouldCloseConnectionAndExitLoop
	}

	// add to route table.
	s.node.routeTable.AddPeerStream(s)

	return nil
}

func CheckClientVersionCompability(v1, v2 string) bool {
	return v1 == v2
}
