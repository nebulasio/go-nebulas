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
	"hash/crc32"
	"time"

	nnet "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	peerstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/nebulasio/go-nebulas/components/net"
	"github.com/nebulasio/go-nebulas/components/net/messages"
	byteutils "github.com/nebulasio/go-nebulas/util/byteutils"
	log "github.com/sirupsen/logrus"
)

// ProtocolID protocol id
const ProtocolID = "/neb/1.0.0"
const HELLO = "hello"
const OK = "ok"
const BYE = "bye"

// MagicNumber the protocol magic number, A constant numerical or text value used to identify protocol.
var MagicNumber = []byte{0x4e, 0x45, 0x42, 0x31}

// NetService service for nebulas p2p network
type NetService struct {
	node       *Node
	dispatcher *net.Dispatcher
}

// RegisterNetService register to Netservice
func (node *Node) RegisterNetService() *NetService {
	ns := &NetService{node, net.NewDispatcher()}
	node.host.SetStreamHandler(ProtocolID, ns.msgHandler)
	log.Infof("NetService: node register net service success...")
	return ns
}

func (ns *NetService) msgHandler(s nnet.Stream) {
	// TODO handle message coming
	defer s.Close()
	log.Info("msgHandler: handle coming msg ...")
	pid := s.Conn().RemotePeer()
	node := ns.node
	timeout := 30 * time.Second
	magicNumber, _ := ReadWithTimeout(s, 4, timeout)
	chainID, _ := ReadWithTimeout(s, 4, timeout)
	ReadWithTimeout(s, 3, timeout)
	version, _ := ReadWithTimeout(s, 1, timeout)
	msgName, _ := ReadWithTimeout(s, 12, timeout)
	dataLength, _ := ReadWithTimeout(s, 4, timeout)
	dataChecksum, _ := ReadWithTimeout(s, 4, timeout)
	data, _ := ReadWithTimeout(s, byteutils.Uint32(dataLength), timeout)

	addrs := node.peerstore.PeerInfo(pid).Addrs

	if !byteutils.Equal(MagicNumber, magicNumber) {
		log.Error("msgHandler: data verification occurs error, magic number is error, the connection will be closed.")
		node.Bye(pid, addrs)
		return
	}

	if node.chainID != byteutils.Uint32(chainID) {
		log.Error("msgHandler: data verification occurs error, chainID is error, the connection will be closed.")
		node.Bye(pid, addrs)
		return
	}

	if !byteutils.Equal([]byte{node.version}, version) {
		log.Error("msgHandler: data verification occurs error, version is error, the connection will be closed.")
		node.Bye(pid, addrs)
		return
	}

	dataChecksumA := crc32.ChecksumIEEE(data)
	if dataChecksumA != byteutils.Uint32(dataChecksum) {
		log.Error("msgHandler: data verification occurs error, dataChecksum is error, the connection will be closed.")
		node.Bye(pid, addrs)
		return
	}

	switch byteutils.Hex(msgName) {
	case HELLO:
		ok, _ := byteutils.FromHex(OK)
		size := byteutils.FromUint32(uint32(len(ok)))
		WriteWithTimeout(
			s,
			append(size[:], ok...),
			timeout,
		)
		node.conn[addrs[0].String()] = S_OK
	case OK:
	case BYE:
	default:
		// only peers connection state is OK, they can send msg freely
		if node.conn[addrs[0].String()] == S_OK {
			msg := messages.NewBaseMessage(byteutils.Hex(msgName), data)
			ns.PutMessage(msg)
		}

	}

}

// Bye say bye to a peer, and close connection.
func (node *Node) Bye(pid peer.ID, addrs []ma.Multiaddr) {
	node.peerstore.SetAddrs(pid, addrs, 0)
	node.routeTable.Remove(pid)
	node.conn[addrs[0].String()] = S_NC
	// Say Bye bye!
}

func (node *Node) SendMsg(msg []byte, pid peer.ID) {
	// TODO Send message

}

func (node *Node) Hello(pid peer.ID) error {
	// TODO Say Hello
	msgName := HELLO
	stream, err := node.host.NewStream(
		node.context,
		pid,
		ProtocolID,
	)

	addrs := node.peerstore.PeerInfo(pid).Addrs

	if err != nil {
		node.peerstore.SetAddrs(pid, addrs, 0)
		node.routeTable.Remove(pid)
		return err
	}

	defer stream.Close()

	data, err := byteutils.FromHex(msgName)
	dataChecksum := crc32.ChecksumIEEE(data)
	metaHeader := buildHeader(node.chainID, msgName, []uint8{node.version}, uint32(len(data)), dataChecksum)
	headerChecksum := crc32.ChecksumIEEE(metaHeader)
	metaHeader = append(metaHeader[:], byteutils.FromUint32(headerChecksum)...)
	totalData := append(metaHeader[:], data...)

	timeout := 30 * time.Second
	err = WriteWithTimeout(
		stream,
		totalData,
		timeout,
	)

	node.conn[addrs[0].String()] = S_Handshaking

	okSize, _ := ReadWithTimeout(stream, 4, timeout)
	ok, _ := ReadWithTimeout(
		stream,
		uint32(len(okSize)),
		timeout,
	)

	if byteutils.Hex(ok) != OK {
		return errors.New("Hello: say hello get incorrect response")
	}

	node.conn[addrs[0].String()] = S_OK
	return nil
}

func (node *Node) SyncRoutes(pid peer.ID) ([]peerstore.PeerInfo, error) {
	return nil, nil
}

func buildHeader(chainId uint32, msgName string, version []uint8, dataLength uint32, dataChecksum uint32) []byte {
	var metaHeader = make([]byte, 256)
	msgNameHex, _ := byteutils.FromHex(msgName)

	copy(metaHeader[00:], MagicNumber)
	copy(metaHeader[32:], byteutils.FromUint32(chainId))
	// 64-88 Reserved field
	copy(metaHeader[88:], version)
	copy(metaHeader[96:], msgNameHex)
	copy(metaHeader[192:], byteutils.FromUint32(dataLength))
	copy(metaHeader[224:], byteutils.FromUint32(dataChecksum))

	return metaHeader
}

// Register register the subscribers.
func (ns *NetService) Register(subscribers ...*net.Subscriber) {
	ns.dispatcher.Register(subscribers...)
}

// Deregister Deregister the subscribers.
func (ns *NetService) Deregister(subscribers ...*net.Subscriber) {
	ns.dispatcher.Deregister(subscribers...)
}

// PutMessage put message to dispatcher.
func (ns *NetService) PutMessage(msg net.Message) {
	ns.dispatcher.PutMessage(msg)
}

// BroadcastBlock broadcast block message
func (ns *NetService) BroadcastBlock(block interface{}) {
	//TODO: broadcast block via underlying network lib to whole network.
	ns.node.Broadcast(block)
}
