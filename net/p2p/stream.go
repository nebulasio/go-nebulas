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
	"time"

	"github.com/nebulasio/go-nebulas/net"
	"github.com/nebulasio/go-nebulas/util/byteutils"

	"github.com/gogo/protobuf/proto"
	libnet "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/nebulasio/go-nebulas/net/messages"
	netpb "github.com/nebulasio/go-nebulas/net/pb"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

const (
	ClientVersion = "0.2.0"
	NebProtocolID = "/neb/1.0.0"
	HELLO         = "hello"
	OK            = "ok"
	BYE           = "bye"
	SYNCROUTE     = "syncroute"
	ROUTETABLE    = "routetable"
	RECVEDMSG     = "recvedmsg"

	// @deprecated.
	SYNCROUTEREPLY = "resyncroute"
)

var (
	ErrShouldCloseConnectionAndExitLoop = errors.New("should close connection and exit loop")
	ErrStreamIsNotConnected             = errors.New("stream is not connected")
)

type Stream struct {
	pid                       peer.ID
	addr                      ma.Multiaddr
	stream                    libnet.Stream
	node                      *Node
	handshakeSucceedCh        chan bool
	messageNotifChan          chan int
	highPriorityMessageChan   chan *NebMessage
	normalPriorityMessageChan chan *NebMessage
	lowPriorityMessageChan    chan *NebMessage
	quitWriteCh               chan bool
	handshakeSucceed          bool
	connectedAt               int64
	latestReadAt              int64
	latestWriteAt             int64
}

// NewStream return a new StreamInfo
func NewStream(stream libnet.Stream, node *Node) *Stream {
	return newStreamInstance(stream.Conn().RemotePeer(), stream.Conn().RemoteMultiaddr(), stream, node)
}

func NewStreamFromPID(pid peer.ID, node *Node) *Stream {
	return newStreamInstance(pid, nil, nil, node)
}

func newStreamInstance(pid peer.ID, addr ma.Multiaddr, stream libnet.Stream, node *Node) *Stream {
	return &Stream{
		pid:                       pid,
		addr:                      addr,
		stream:                    stream,
		node:                      node,
		handshakeSucceedCh:        make(chan bool, 1),
		messageNotifChan:          make(chan int, 4*1024),
		highPriorityMessageChan:   make(chan *NebMessage, 2*1024),
		normalPriorityMessageChan: make(chan *NebMessage, 2*1024),
		lowPriorityMessageChan:    make(chan *NebMessage, 2*1024),
		quitWriteCh:               make(chan bool, 1),
		handshakeSucceed:          false,
		connectedAt:               time.Now().Unix(),
		latestReadAt:              0,
		latestWriteAt:             0,
	}
}

func (s *Stream) Connect() error {
	logging.VLog().WithFields(logrus.Fields{
		"stream": s.String(),
	}).Debug("Connecting to peer.")

	// connect to host.
	stream, err := s.node.host.NewStream(
		s.node.context,
		s.pid,
		NebProtocolID,
	)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"stream": s.String(),
			"err":    err,
		}).Debug("Failed to connect to host.")
		return err
	}
	s.stream = stream
	s.addr = stream.Conn().RemoteMultiaddr()

	return nil
}

func (s *Stream) IsConnected() bool {
	return s.stream != nil
}

func (s *Stream) IsHandshakeSucceed() bool {
	return s.handshakeSucceed
}

func (s *Stream) String() string {
	addrStr := ""
	if s.addr != nil {
		addrStr = s.addr.String()
	}

	return fmt.Sprintf("Peer Stream: %s,%s", s.pid.Pretty(), addrStr)
}

func (s *Stream) SendProtoMessage(messageName string, pb proto.Message, priority int) error {
	data, err := proto.Marshal(pb)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err":         err,
			"messageName": messageName,
			"stream":      s.String(),
		}).Warn("Failed to marshal proto message.")
		return err
	}

	return s.SendMessage(messageName, data, priority)
}

func (s *Stream) SendMessage(messageName string, data []byte, priority int) error {
	message, err := NewNebMessage(s.node.config.ChainID, DefaultReserved, 0, messageName, data)
	if err != nil {
		return err
	}

	// metrics.
	metricsPacketsOutByMessageName(messageName, message.Length())

	// send to pool.
	message.FlagSendMessageAt()

	switch priority {
	case net.MessagePriorityHigh:
		s.highPriorityMessageChan <- message
	case net.MessagePriorityNormal:
		s.normalPriorityMessageChan <- message
	default:
		s.lowPriorityMessageChan <- message
	}
	s.messageNotifChan <- 1

	return nil
}

func (s *Stream) Write(data []byte) error {
	if s.stream == nil {
		s.Close(ErrStreamIsNotConnected)
		return ErrStreamIsNotConnected
	}

	n, err := s.stream.Write(data)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err":    err,
			"stream": s.String(),
		}).Warn("Failed to send message to peer.")
		s.Close(err)
		return err
	}
	s.latestWriteAt = time.Now().Unix()

	// metrics.
	metricsPacketsOut.Mark(1)
	metricsBytesOut.Mark(int64(n))

	return nil
}

func (s *Stream) WriteNebMessage(message *NebMessage) error {
	// metrics.
	metricsPacketsOutByMessageName(message.MessageName(), message.Length())

	message.FlagWriteMessageAt()

	err := s.Write(message.Content())

	// debug logs.
	logging.VLog().WithFields(logrus.Fields{
		"stream":      s.String(),
		"err":         err,
		"checksum":    message.DataCheckSum(),
		"messageName": message.MessageName(),
		"latency(ms)": message.LatencyFromSendToWrite(),
	}).Debugf("Written %s message to peer.", message.MessageName())

	return err
}

func (s *Stream) WriteProtoMessage(messageName string, pb proto.Message) error {
	data, err := proto.Marshal(pb)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err":         err,
			"messageName": messageName,
			"stream":      s.String(),
		}).Warn("Failed to marshal proto message.")
		return err
	}

	return s.WriteMessage(messageName, data)
}

func (s *Stream) WriteMessage(messageName string, data []byte) error {
	message, err := NewNebMessage(s.node.config.ChainID, DefaultReserved, 0, messageName, data)
	if err != nil {
		return err
	}

	return s.WriteNebMessage(message)
}

// StartLoop start stream handling loop.
func (s *Stream) StartLoop() {
	go s.writeLoop()
	go s.readLoop()
}

func (s *Stream) readLoop() {
	buf := make([]byte, 1024*4)
	messageBuffer := []byte{}

	var message *NebMessage

	// send Hello to host if stream is not connected.
	if !s.IsConnected() {
		if err := s.Connect(); err != nil {
			s.Close(err)
			return
		}
		if err := s.Hello(); err != nil {
			s.Close(err)
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
					"err":    err,
					"stream": s.String(),
				}).Error("Error occurred when reading data from network connection.")
				s.Close(err)
				return
			}
			messageBuffer = append(messageBuffer, buf[:n]...)
			s.latestReadAt = time.Now().Unix()

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
						"stream":          s.String(),
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
			metricsPacketsIn.Mark(1)
			metricsBytesIn.Mark(int64(message.Length()))
			metricsPacketsInByMessageName(message.MessageName(), message.Length())

			// handle message.
			if err := s.handleMessage(message); err == ErrShouldCloseConnectionAndExitLoop {
				s.Bye()
				return
			}

			// reset message.
			message = nil
		}
	}
}

func (s *Stream) writeLoop() {
	// waiting for handshake succeed.
	handshakeTimeoutTicker := time.NewTicker(30 * time.Second)
	select {
	case <-s.handshakeSucceedCh:
		// handshake succeed.
	case <-s.quitWriteCh:
		logging.VLog().WithFields(logrus.Fields{
			"stream": s.String(),
		}).Debug("Quiting Stream Write Loop.")
		return
	case <-handshakeTimeoutTicker.C:
		logging.VLog().WithFields(logrus.Fields{
			"stream": s.String(),
		}).Debug("Handshaking Stream timeout, quiting.")
		s.Close(errors.New("Handshake timeout"))
		return
	}

	for {
		select {
		case <-s.quitWriteCh:
			logging.VLog().WithFields(logrus.Fields{
				"stream": s.String(),
			}).Debug("Quiting Stream Write Loop.")
			return
		case <-s.messageNotifChan:
			select {
			case message := <-s.highPriorityMessageChan:
				s.WriteNebMessage(message)
				continue
			default:
			}

			select {
			case message := <-s.normalPriorityMessageChan:
				s.WriteNebMessage(message)
				continue
			default:
			}

			select {
			case message := <-s.lowPriorityMessageChan:
				s.WriteNebMessage(message)
				continue
			default:
			}
		}
	}
}

func (s *Stream) handleMessage(message *NebMessage) error {
	messageName := message.MessageName()

	switch messageName {
	case HELLO:
		return s.onHello(message)
	case OK:
		return s.onOk(message)
	case BYE:
		return s.onBye(message)
	}

	// check handshake status.
	if s.handshakeSucceed == false {
		return ErrShouldCloseConnectionAndExitLoop
	}

	switch messageName {
	case SYNCROUTE:
		return s.onSyncRoute(message)
	case SYNCROUTEREPLY:
		return s.onRouteTable(message)
	case ROUTETABLE:
		return s.onRouteTable(message)
	case RECVEDMSG:
		return s.OnRecvedMsg(message)
	default:
		logging.VLog().WithFields(logrus.Fields{
			"messageName": messageName,
			"checksum":    message.DataCheckSum(),
			"stream":      s.String(),
		}).Debugf("Received %s message from peer.", messageName)

		s.node.netService.PutMessage(messages.NewBaseMessage(message.MessageName(), s.pid.Pretty(), message.Data()))

		// record recv message.
		RecordRecvMessage(s, message.DataCheckSum())
	}

	return nil
}

func (s *Stream) Close(reason error) {
	logging.VLog().WithFields(logrus.Fields{
		"stream": s.String(),
		"reason": reason,
	}).Debug("Disconnecting stream.")

	// quit.
	s.quitWriteCh <- true

	// close stream.
	if s.stream != nil {
		s.stream.Close()
	}

	// cleanup.
	s.node.streamManager.RemoveStream(s)
	s.node.routeTable.RemovePeerStream(s)
}

func (s *Stream) Bye() {
	s.WriteMessage(BYE, []byte{})
	s.Close(errors.New("bye: force close"))
}

func (s *Stream) onBye(message *NebMessage) error {
	logging.VLog().WithFields(logrus.Fields{
		"stream": s.String(),
	}).Debug("Received Bye message, close the connection.")
	return ErrShouldCloseConnectionAndExitLoop
}

func (s *Stream) Hello() error {
	msg := &netpb.Hello{
		NodeId:        s.node.id.String(),
		ClientVersion: ClientVersion,
	}
	return s.WriteProtoMessage(HELLO, msg)
}

func (s *Stream) onHello(message *NebMessage) error {
	msg, err := netpb.HelloMessageFromProto(message.Data())
	if err != nil {
		return ErrShouldCloseConnectionAndExitLoop
	}

	if msg.NodeId != s.pid.String() || !CheckClientVersionCompability(ClientVersion, msg.ClientVersion) {
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

	// handshake finished.
	s.finishHandshake()

	return s.Ok()
}

func (s *Stream) Ok() error {
	// send OK.
	resp := &netpb.OK{
		NodeId:        s.node.id.String(),
		ClientVersion: ClientVersion,
	}

	return s.WriteProtoMessage(OK, resp)
}

func (s *Stream) onOk(message *NebMessage) error {
	msg, err := netpb.OKMessageFromProto(message.Data())
	if err != nil {
		return ErrShouldCloseConnectionAndExitLoop
	}

	if msg.NodeId != s.pid.String() || !CheckClientVersionCompability(ClientVersion, msg.ClientVersion) {
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

	// handshake finished.
	s.finishHandshake()

	return nil
}

func (s *Stream) SyncRoute() error {
	return s.SendMessage(SYNCROUTE, []byte{}, net.MessagePriorityHigh)
}

func (s *Stream) onSyncRoute(message *NebMessage) error {
	return s.RouteTable()
}

func (s *Stream) RouteTable() error {
	// get nearest peers from routeTable
	peers := s.node.routeTable.GetNearestPeers(s.pid)

	// prepare the protobuf message.
	msg := &netpb.Peers{
		Peers: make([]*netpb.PeerInfo, len(peers)),
	}

	for i, v := range peers {
		pi := &netpb.PeerInfo{
			Id:    string(v.ID),
			Addrs: make([]string, len(v.Addrs)),
		}
		for j, addr := range v.Addrs {
			pi.Addrs[j] = addr.String()
		}
		msg.Peers[i] = pi
	}

	logging.VLog().WithFields(logrus.Fields{
		"stream":          s.String(),
		"routetableCount": len(peers),
	}).Debug("Replied sync route message.")

	// @deprecated.
	s.SendProtoMessage(SYNCROUTEREPLY, msg, net.MessagePriorityHigh)

	return s.SendProtoMessage(ROUTETABLE, msg, net.MessagePriorityHigh)
}

func (s *Stream) onRouteTable(message *NebMessage) error {
	peers := new(netpb.Peers)
	if err := proto.Unmarshal(message.Data(), peers); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Debug("Invalid Peers proto message.")
		return ErrShouldCloseConnectionAndExitLoop
	}

	for _, v := range peers.Peers {
		s.node.routeTable.AddPeerInfo(v.Id, v.Addrs)
	}

	return nil
}

func (s *Stream) RecvedMsg(hash uint32) error {
	return s.SendMessage(RECVEDMSG, byteutils.FromUint32(hash), net.MessagePriorityHigh)
}

func (s *Stream) OnRecvedMsg(message *NebMessage) error {
	hash := byteutils.Uint32(message.Data())
	RecordRecvMessage(s, hash)

	return nil
}

func (s *Stream) finishHandshake() {
	logging.VLog().WithFields(logrus.Fields{
		"stream": s.String(),
	}).Debug("Finished handshake.")

	s.handshakeSucceed = true
	s.handshakeSucceedCh <- true
}

func CheckClientVersionCompability(v1, v2 string) bool {
	return v1 == v2
}
