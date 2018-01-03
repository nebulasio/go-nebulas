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
	"bytes"
	"errors"
	"hash/crc32"
	"io"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	kbucket "github.com/libp2p/go-libp2p-kbucket"
	libnet "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	peerstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/nebulasio/go-nebulas/net"
	"github.com/nebulasio/go-nebulas/net/messages"
	"github.com/nebulasio/go-nebulas/net/pb"
	byteutils "github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/nebulasio/go-nebulas/util/logging"
	metrics "github.com/rcrowley/go-metrics"
	"github.com/sirupsen/logrus"
)

// connection state
const (
	ProtocolID     = "/neb/1.0.0"
	SNC            = -1
	SHandshaking   = 0
	SOK            = 1
	HELLO          = "hello"
	OK             = "ok"
	BYE            = "bye"
	SyncRoute      = "syncroute"
	SyncRouteReply = "resyncroute"
	NewHashMsg     = "newhashmsg"
	ClientVersion  = "0.2.0"
	NetworkID      = "networkid"
	NetworkIDReply = "renetworkid"
)

// MagicNumber the protocol magic number, A constant numerical or text value used to identify protocol.
var MagicNumber = []byte{0x4e, 0x45, 0x42, 0x31}

var (
	offsetFour        = 4
	offsetEight       = 8
	offsetEleven      = 11
	offsetTwelve      = 12
	offsetTwentyFour  = 24
	offsetTwentyEight = 28
	offsetThirtyTwo   = 32
	offsetThirtySix   = 36
)

var (
	packetsIn   = metrics.GetOrRegisterMeter("neb.net.packets.in", nil)
	packetsOut  = metrics.GetOrRegisterMeter("neb.net.packets.out", nil)
	netBytesIn  = metrics.GetOrRegisterMeter("neb.net.bytes.in", nil)
	netBytesOut = metrics.GetOrRegisterMeter("neb.net.bytes.out", nil)
)

// NetService service for nebulas p2p network
type NetService struct {
	node       *Node
	quitCh     chan bool
	dispatcher *net.Dispatcher
}

/*
Protocol In Nebulas, we define our own wire protocol, as the following:

 0               1               2               3              (bytes)
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                         Magic Number                          |
+---------------------------------------------------------------+
|                         Chain ID                              |
+-----------------------------------------------+---------------+
|                         Reserved              |   Version     |
+-----------------------------------------------+---------------+
|                                                               |
+                                                               +
|                         Message Name                          |
+                                                               +
|                                                               |
+---------------------------------------------------------------+
|                         Data Length                           |
+---------------------------------------------------------------+
|                         Data Checksum                         |
+---------------------------------------------------------------+
|                         Header Checksum                       |
|---------------------------------------------------------------+
|                                                               |
+                         Data                                  +
.                                                               .
|                                                               |
+---------------------------------------------------------------+
*/
type NebMessage struct {
	magicNumber    []byte
	chainID        []byte
	version        byte
	msgName        string
	dataLength     []byte
	dataChecksum   []byte
	headerChecksum []byte
	header         []byte
	data           []byte
	reserved       []byte
}

// NewNetManager create netService
func NewNetManager(n Neblet) (*NetService, error) {
	config := NewP2PConfig(n)
	node, err := NewNode(config)
	if err != nil {
		logging.VLog().Error("NewNetService: node create fail -> ", err)
		return nil, err
	}
	ns := &NetService{node, make(chan bool), net.NewDispatcher()}
	return ns, nil
}

func (ns *NetService) registerNetManager() *NetService {
	// register streamHandler to start loop to handle stream origined from remote node.
	ns.node.host.SetStreamHandler(ProtocolID, ns.streamHandler)
	logging.VLog().Info("RegisterNetService: register netservice success")
	return ns
}

// Addrs return peer address in string
func (ns *NetService) Addrs() ma.Multiaddr {
	len := len(ns.node.host.Addrs())
	if len > 0 {
		return ns.node.host.Addrs()[0]
	}
	return nil

}

// Node return the peer node
func (ns *NetService) Node() *Node {
	return ns.node
}

func (ns *NetService) streamHandler(s libnet.Stream) {
	var tmpMsg *NebMessage
	var dataLength uint32

	streamBuffer := []byte{}
	sdata := make([]byte, 1024)

	node := ns.node
	pid := s.Conn().RemotePeer()
	addrs := s.Conn().RemoteMultiaddr()
	key := pid.Pretty()

	for {
		select {
		case <-ns.quitCh:
			return
		default:
			n, err := s.Read(sdata)
			if err != nil {
				logging.VLog().WithFields(logrus.Fields{
					"err":   err,
					"addrs": addrs,
				}).Error("Connectoin closed.")
				ns.Bye(pid, []ma.Multiaddr{addrs}, s, key)
				return
			}
			streamBuffer = append(streamBuffer, sdata[:n]...)

			if tmpMsg == nil {
				// wait to parseHeader
				if len(streamBuffer) < offsetThirtySix {
					continue
				}
				tmpMsg, err = ns.parseMsgHeader(streamBuffer)
				if err != nil {
					logging.VLog().WithFields(logrus.Fields{
						"addrs": addrs.String(),
						"err":   err,
					}).Error("parse header error")
					ns.Bye(pid, []ma.Multiaddr{addrs}, s, key)
					return
				}

				streamBuffer = streamBuffer[offsetThirtySix:]
				dataLength = byteutils.Uint32(tmpMsg.dataLength)
			}

			if dataLength > uint32(len(streamBuffer)) {
				// stream data is not enough
				continue
			}

			if err = ns.parseMsgData(tmpMsg, streamBuffer); err != nil {
				logging.VLog().WithFields(logrus.Fields{
					"addrs": addrs.String(),
					"err":   err,
				}).Error("parse data error")
				ns.Bye(pid, []ma.Multiaddr{addrs}, s, key)
				return
			}
			streamBuffer = streamBuffer[dataLength:]

			msg := tmpMsg
			tmpMsg = nil
			dataLength = 0

			packetsIn.Mark(1)
			netBytesIn.Mark(int64(byteutils.Uint32(msg.dataLength) + uint32(offsetThirtySix)))

			switch msg.msgName {
			case HELLO:
				ns.handleHelloMsg(msg.data, pid, s, addrs, key)
			case OK:
				ns.handleOkMsg(msg.data, pid, s, addrs, key)
			case BYE:

			case SyncRoute:
				ns.handleSyncRouteMsg(msg.data, pid, s, addrs, key)
			case SyncRouteReply:
				ns.handleSyncRouteReplyMsg(msg.data, pid, s, addrs)
			case NewHashMsg:
				ns.handleNewHashMsg(msg.data, pid)
			case NetworkID:
				ns.handleNetworkIDMsg(msg.data, pid, s)
			case NetworkIDReply:
				ns.handleReNetworkIDMsg(msg.data, pid)
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
					ns.Bye(pid, []ma.Multiaddr{addrs}, s, key)
					return
				}
				if streamStore.(*StreamStore).conn != SOK {
					logging.VLog().Error("peer not shake hand before send message.")
					ns.Bye(pid, []ma.Multiaddr{addrs}, s, key)
					return
				}
				ns.PutMessage(messages.NewBaseMessage(msg.msgName, pid.Pretty(), msg.data))

				peers, exists := node.relayness.Get(byteutils.Uint32(msg.dataChecksum))
				if exists {
					relayness = peers.([]peer.ID)
				}
				node.relayness.Add(byteutils.Uint32(msg.dataChecksum), append(relayness, pid))
			}

		}
	}

}

func (ns *NetService) parseMsgHeader(streamBuffer []byte) (*NebMessage, error) {
	header := streamBuffer

	nebMsg := &NebMessage{}
	nebMsg.magicNumber = header[:offsetFour]
	nebMsg.chainID = header[offsetFour:offsetEight]
	nebMsg.reserved = header[offsetEight:offsetEleven]
	nebMsg.version = header[offsetEleven]
	msgName := header[offsetTwelve:offsetTwentyFour]
	nebMsg.dataLength = header[offsetTwentyFour:offsetTwentyEight]
	nebMsg.dataChecksum = header[offsetTwentyEight:offsetThirtyTwo]
	nebMsg.headerChecksum = header[offsetThirtyTwo:offsetThirtySix]
	nebMsg.header = header[:offsetThirtyTwo]

	index := bytes.IndexByte(msgName, 0)
	if index != -1 {
		msgNameByte := msgName[0:index]
		nebMsg.msgName = string(msgNameByte)
	} else {
		nebMsg.msgName = string(msgName)
	}

	if !ns.verifyHeader(nebMsg) {
		return nil, errors.New("invalid neb message header")
	}

	logging.VLog().WithFields(logrus.Fields{
		"msgName":      nebMsg.msgName,
		"magicNumber":  string(nebMsg.magicNumber),
		"chainID":      byteutils.Uint32(nebMsg.chainID),
		"version":      nebMsg.version,
		"dataChecksum": byteutils.Uint32(nebMsg.dataChecksum),
		"dataLength":   byteutils.Uint32(nebMsg.dataLength),
	}).Info("parse protocol header data.")
	return nebMsg, nil
}

func (ns *NetService) parseMsgData(nebMsg *NebMessage, streamBuffer []byte) error {

	dataLength := byteutils.Uint32(nebMsg.dataLength)
	nebMsg.data = streamBuffer[:dataLength]

	dataChecksumA := crc32.ChecksumIEEE(nebMsg.data)
	if dataChecksumA != byteutils.Uint32(nebMsg.dataChecksum) {
		logging.VLog().WithFields(logrus.Fields{
			"dataChecksumA": dataChecksumA,
			"dataChecksum":  byteutils.Uint32(nebMsg.dataChecksum),
		}).Error("invalid neb message data")
		return errors.New("invalid neb message data")
	}
	return nil
}

func (ns *NetService) handleHelloMsg(data []byte, pid peer.ID, s libnet.Stream, addrs ma.Multiaddr, key string) bool {
	node := ns.node
	result := false
	defer func() {
		if !result {
			ns.Bye(pid, []ma.Multiaddr{addrs}, s, key)
		}
	}()

	hello := new(messages.HelloMessage)
	pb := new(netpb.Hello)
	if err := proto.Unmarshal(data, pb); err != nil {
		logging.VLog().Error("handle hello msg occurs error: ", err)
		return result
	}
	if err := hello.FromProto(pb); err != nil {
		logging.VLog().Error("handle hello msg occurs error: ", err)
		return result
	}

	logging.VLog().WithFields(logrus.Fields{
		"hello.NodeID":  hello.NodeID,
		"pid":           pid,
		"addrs":         addrs.String(),
		"ClientVersion": hello.ClientVersion,
	}).Info("receive hello message.")

	//Todo: clientVersion backwards compatible
	if hello.NodeID == pid.String() && hello.ClientVersion == ClientVersion {
		ok := messages.NewHelloMessage(node.id.String(), ClientVersion)
		pbok, err := ok.ToProto()
		okdata, err := proto.Marshal(pbok)
		if err != nil {
			logging.VLog().Error("handleHelloMsg send ok message occurs error, ", err)
			return result
		}

		node.peerstore.AddAddr(
			pid,
			addrs,
			peerstore.PermanentAddrTTL,
		)

		if err := ns.sendMsg(OK, okdata, s); err != nil {
			logging.VLog().Error("send ok msg occurs error, ", err)
			return result
		}

		networkIDData := byteutils.FromUint32(node.Config().NetworkID)
		if err := ns.sendMsg(NetworkID, networkIDData, s); err != nil {
			logging.VLog().Error("send networkID msg occurs error, ", err)
			return result
		}

		streamStore := NewStreamStore(key, SOK, s)
		node.stream.Store(key, streamStore)
		node.streamCache.Insert(streamStore)
		node.routeTable.Update(pid)
		result = true
		return result
	}
	return result

}

func (ns *NetService) handleOkMsg(data []byte, pid peer.ID, s libnet.Stream, addrs ma.Multiaddr, key string) bool {
	node := ns.node
	result := false
	defer func() {
		if !result {
			ns.Bye(pid, []ma.Multiaddr{addrs}, s, key)
		}
	}()

	ok := new(messages.HelloMessage)
	pb := new(netpb.Hello)
	if err := proto.Unmarshal(data, pb); err != nil {
		logging.VLog().Error("handle ok msg occurs error: ", err)
		return result
	}
	if err := ok.FromProto(pb); err != nil {
		logging.VLog().Error("handle ok msg occurs error: ", err)
		return result
	}

	if ok.NodeID == pid.String() && ok.ClientVersion == ClientVersion {
		streamStore := NewStreamStore(key, SOK, s)
		node.stream.Store(key, streamStore)
		node.streamCache.Insert(streamStore)
		node.peerstore.AddAddr(
			pid,
			addrs,
			peerstore.PermanentAddrTTL,
		)
		node.routeTable.Update(pid)

		result = true
		return result
	}

	logging.VLog().Error("handleOkMsg get incorrect response")
	return result

}

func (ns *NetService) handleNetworkIDMsg(data []byte, pid peer.ID, s libnet.Stream) {
	node := ns.node
	networkID := byteutils.Uint32(data)
	node.networkIDCache.Add(pid.Pretty(), networkID)

	networkIDData := byteutils.FromUint32(node.Config().NetworkID)
	if err := ns.sendMsg(NetworkIDReply, networkIDData, s); err != nil {
		logging.VLog().Error("send networkID msg occurs error, ", err)
	}

}

func (ns *NetService) handleReNetworkIDMsg(data []byte, pid peer.ID) {
	node := ns.node
	networkID := byteutils.Uint32(data)
	node.networkIDCache.Add(pid.Pretty(), networkID)
}

func (ns *NetService) handleNewHashMsg(data []byte, pid peer.ID) {
	var relayness []peer.ID
	node := ns.node
	peers, exists := node.relayness.Get(byteutils.Uint32(data))
	if exists {
		relayness = peers.([]peer.ID)
	}
	node.relayness.Add(byteutils.Uint32(data), append(relayness, pid))
}

func (ns *NetService) handleSyncRouteMsg(data []byte, pid peer.ID, s libnet.Stream, addrs ma.Multiaddr, key string) bool {
	node := ns.node
	result := false
	defer func() {
		if !result {
			ns.Bye(pid, []ma.Multiaddr{addrs}, s, key)
		}
	}()
	peers := node.routeTable.NearestPeers(kbucket.ConvertPeerID(pid), node.config.MaxSyncNodes)
	var peerList []*messages.PeerInfo
	for i := range peers {
		peerInfo := node.peerstore.PeerInfo(peers[i])
		if len(peerInfo.Addrs) == 0 {
			logging.VLog().WithFields(logrus.Fields{
				"nodeId": peerInfo.ID.Pretty(),
			}).Warn("node addrs is nil")
			continue
		}
		var addres []string
		for _, v := range peerInfo.Addrs {
			addres = append(addres, v.String())
		}
		peer := messages.NewPeerInfoMessage(peerInfo.ID, addres)
		peerList = append(peerList, peer)
	}
	logging.VLog().WithFields(logrus.Fields{
		"remoteId":    pid.Pretty(),
		"remoteAddrs": addrs,
		"count":       len(peerList),
	}).Info("reply sync route to remote node")

	peersMessage := messages.NewPeersMessage(peerList)

	pb, err := peersMessage.ToProto()
	data, err = proto.Marshal(pb)
	if err != nil {
		logging.VLog().Error("handleSyncRouteMsg occurs error, ", err)
		return result
	}

	if err := ns.SendMsg(SyncRouteReply, data, key); err != nil {
		return result
	}

	node.routeTable.Update(pid)
	result = true
	return result
}

func (ns *NetService) handleSyncRouteReplyMsg(data []byte, pid peer.ID, s libnet.Stream, addrs ma.Multiaddr) bool {
	node := ns.node
	peers := new(messages.Peers)
	pb := new(netpb.Peers)

	if err := proto.Unmarshal(data, pb); err != nil {
		logging.VLog().Error("handleSyncRouteReplyMsg occurs error: ", err)
		return false
	}
	if err := peers.FromProto(pb); err != nil {
		logging.VLog().Error("handleSyncRouteReplyMsg occurs error: ", err)
		return false
	}

	for i := range peers.Peers() {
		id := peers.Peers()[i].ID()
		if node.routeTable.Find(id) != "" || len(peers.Peers()[i].Addrs()) == 0 {
			logging.VLog().WithFields(logrus.Fields{
				"id": id.Pretty(),
			}).Warn("node is already exist in route table")
			continue
		}
		var addres []ma.Multiaddr
		for _, v := range peers.Peers()[i].Addrs() {
			addr, _ := ma.NewMultiaddr(v)
			addres = append(addres, addr)
		}

		logging.VLog().WithFields(logrus.Fields{
			"id":    id.Pretty(),
			"addrs": addres,
		}).Info("discover new node")

		node.peerstore.AddAddrs(
			id,
			addres,
			peerstore.ProviderAddrTTL,
		)
		if err := ns.Hello(id); err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"id":  id.Pretty(),
				"err": err,
			}).Error("say hello to the peer fail")
			continue
		}

		node.routeTable.Update(id)
	}
	return true
}

func (ns *NetService) verifyHeader(nebMsg *NebMessage) bool {

	node := ns.node
	headerChecksum := crc32.ChecksumIEEE(nebMsg.header)

	if !byteutils.Equal(MagicNumber, nebMsg.magicNumber) {
		logging.VLog().Debug("verifyHeader: data verification occurs error, magic number is error, the connection will be closed.")
		return false
	}

	if node.Config().ChainID != byteutils.Uint32(nebMsg.chainID) {
		logging.VLog().Debug("verifyHeader: data verification occurs error, chainID is error, the connection will be closed.")
		return false
	}

	if node.version != nebMsg.version {
		logging.VLog().Debug("verifyHeader: data verification occurs error, version is error, the connection will be closed.")
		return false
	}

	if headerChecksum != byteutils.Uint32(nebMsg.headerChecksum) {
		logging.VLog().Debug("verifyHeader: data verification occurs error, dataHeaderChecksum is error, the connection will be closed.")
		return false
	}
	return true
}

// Bye say bye to a peer, and close connection.
func (ns *NetService) Bye(pid peer.ID, addrs []ma.Multiaddr, s libnet.Stream, key string) {
	node := ns.node
	ns.clearPeerStore(pid, addrs)
	node.stream.Delete(key)
	s.Close()
}

func (ns *NetService) clearPeerStore(pid peer.ID, addrs []ma.Multiaddr) {
	node := ns.node
	node.peerstore.SetAddrs(pid, addrs, 0)
	if !InArray(pid.Pretty(), node.bootIds) {
		node.routeTable.Remove(pid)
	}
}

// SendMsg send message to a peer
func (ns *NetService) sendMsg(msgName string, msg []byte, stream libnet.Stream) error {

	logging.VLog().WithFields(logrus.Fields{
		"msgName": msgName,
	}).Info("SendMsg: send message to a peer.")
	totalData := ns.buildData(msg, msgName)

	if err := Write(stream, totalData); err != nil {
		logging.VLog().Error("SendMsg: write data occurs error, ", err)
		return err
	}
	packetsOut.Mark(1)
	m, ok := net.PacketsOutByTypes.Load(msgName)
	if ok {
		m.(metrics.Meter).Mark(1)
	}
	netBytesOut.Mark(int64(len(msg)))
	return nil
}

// SendMsg send message to a peer
func (ns *NetService) SendMsg(msgName string, msg []byte, target string) error {

	node := ns.node
	if msgName != NetworkID && !ns.checkNetworkID(target) {
		logging.VLog().Warn("can not send message, target node is not in the same network ", target)
		return errors.New("can not send message, target node is not in the same network")
	}
	streamStore, ok := node.stream.Load(target)
	if !ok {
		return errors.New("handleSyncRouteMsg occrus error, stream does not exist")
	}
	return ns.sendMsg(msgName, msg, streamStore.(*StreamStore).stream)
}

func (ns *NetService) checkNetworkID(target string) bool {
	node := ns.node
	targetNetworkID, ok := node.networkIDCache.Get(target)
	if ok {
		logging.VLog().WithFields(logrus.Fields{
			"targetNetworkID": targetNetworkID,
			"result":          node.config.NetworkID & targetNetworkID.(uint32),
		}).Info("checkNetworkID.")
		if node.config.NetworkID&targetNetworkID.(uint32) > 0 {
			return true
		}
	}
	return false
}

// Hello say hello to a peer
func (ns *NetService) Hello(pid peer.ID) error {
	node := ns.node

	stream, err := node.host.NewStream(
		node.context,
		pid,
		ProtocolID,
	)
	if err != nil {
		return err
	}

	hello := messages.NewHelloMessage(node.id.String(), ClientVersion)
	pb, _ := hello.ToProto()
	data, err := proto.Marshal(pb)
	if err != nil {
		return err
	}
	if err = ns.sendMsg(HELLO, data, stream); err != nil {
		return err
	}
	// call streamHandler explicitly to start loop to handle stream origined from this node.
	go ns.streamHandler(stream)
	return nil
}

// SyncRoutes sync routing table from a peer
func (ns *NetService) SyncRoutes(pid peer.ID) {
	node := ns.node
	addrs := node.peerstore.PeerInfo(pid).Addrs
	if len(addrs) == 0 {
		logging.VLog().Error("SyncRoutes: wrong pid addrs")
		ns.clearPeerStore(pid, addrs)
		return
	}
	data := []byte{}
	if err := ns.SendMsg(SyncRoute, data, pid.Pretty()); err != nil {
		logging.VLog().Error("SyncRoutes: write data occurs error, ", err)
		ns.clearPeerStore(pid, addrs)
		return
	}

}

// buildHeader build header information
func buildHeader(chainID uint32, msgName string, version byte, dataLength uint32, dataChecksum uint32, reserved []byte) []byte {
	var metaHeader = make([]byte, offsetThirtyTwo)
	msgNameByte := []byte(msgName)

	copy(metaHeader[:], MagicNumber)
	copy(metaHeader[offsetFour:], byteutils.FromUint32(chainID))
	// 64-88 Reserved field
	copy(metaHeader[offsetEight:], reserved)
	copy(metaHeader[offsetEleven:], []byte{version})
	copy(metaHeader[offsetTwelve:], msgNameByte)
	copy(metaHeader[offsetTwentyFour:], byteutils.FromUint32(dataLength))
	copy(metaHeader[offsetTwentyEight:], byteutils.FromUint32(dataChecksum))

	return metaHeader
}

func (ns *NetService) buildData(data []byte, msgName string) []byte {
	node := ns.node
	dataChecksum := crc32.ChecksumIEEE(data)
	reserved := []byte{0}
	metaHeader := buildHeader(node.config.ChainID, msgName, node.version, uint32(len(data)), dataChecksum, reserved)
	headerChecksum := crc32.ChecksumIEEE(metaHeader)
	metaHeader = append(metaHeader[:], byteutils.FromUint32(headerChecksum)...)
	totalData := append(metaHeader[:], data...)
	return totalData
}

// BuildData returns net service request data
func (ns *NetService) BuildData(data []byte, msgName string) []byte {
	return ns.buildData(data, msgName)
}

// Start start p2p manager.
func (ns *NetService) Start() error {
	err := ns.start()
	ns.dispatcher.Start()
	return err
}

// Stop stop p2p manager.
func (ns *NetService) Stop() {
	ns.dispatcher.Stop()
	ns.quitCh <- true
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

func (ns *NetService) start() error {

	node := ns.node
	logging.CLog().WithFields(logrus.Fields{
		"id":    node.ID(),
		"addrs": node.host.Addrs(),
	}).Info("node start")
	if node.running {
		return errors.New("net.start: node already running")
	}
	node.running = true

	ns.registerNetManager()

	// TODO: All fail handle
	var success bool
	var wg sync.WaitGroup
	for _, bootNode := range node.config.BootNodes {
		wg.Add(1)
		go func(bootNode ma.Multiaddr) {
			defer wg.Done()
			err := ns.SayHello(bootNode)
			if err != nil {
				logging.VLog().Error("net.start: can not say hello to trusted node.", bootNode, err)
			} else {
				logging.CLog().Info("net.start: say hello to trusted node.", bootNode)
				success = true
			}

		}(bootNode)
	}
	wg.Wait()

	if success || len(node.Config().BootNodes) == 0 {
		go ns.discovery(node.context)
		go ns.manageStreamStore()
		logging.CLog().Infof("net.start: node start and join to p2p network success and listening for connections on %s... ", node.config.Listen)
	} else {
		logging.VLog().Error("net.start: node start occurs error, say hello to bootNode fail")
		return errors.New("net.start: node start occurs error, say hello to bootNode fail")
	}
	return nil
}

func (ns *NetService) manageStreamStore() {
	second := 30 * time.Second
	ticker := time.NewTicker(second)
	for {
		select {
		case <-ticker.C:
			ns.clearStreamStore()
			ns.cleanPeerStore()
		case <-ns.quitCh:
			return
		}
	}
}

func (ns *NetService) cleanPeerStore() {
	node := ns.node
	for _, v := range node.peerstore.Peers() {
		if _, ok := node.stream.Load(v.Pretty()); !ok {
			if !InArray(v.Pretty(), node.bootIds) {
				node.peerstore.ClearAddrs(v)
			}
		}
	}
}

func (ns *NetService) clearStreamStore() {
	node := ns.node
	// do clear streamStore only when the count of stream in cache exceed the cache size.
	if ns.node.streamCache.Len() > ns.node.config.StreamStoreSize {
		overflowSize := ns.node.streamCache.Len() - ns.node.config.StreamStoreSize
		for i := 0; i < overflowSize; i++ {
			streamStore := node.streamCache.PopMin().(*StreamStore)
			key := streamStore.key

			if streamStore, ok := node.stream.Load(key); ok {
				streamStore.(*StreamStore).stream.Close()
				node.stream.Delete(key)
			}
		}
	}
}

// Write write bytes to stream
func Write(writer io.Writer, data []byte) error {
	result := make(chan error, 1)
	go func(writer io.Writer, data []byte) {
		if writer == nil {
			result <- errors.New("write data occurs error, write is nil")
			return
		}
		_, err := writer.Write(data)
		result <- err
	}(writer, data)
	err := <-result
	return err
}

// ReadBytes read bytes from a stream
func ReadBytes(reader io.Reader, n uint32) ([]byte, error) {
	data := make([]byte, n)
	result := make(chan error, 1)
	go func(reader io.Reader) {
		_, err := io.ReadFull(reader, data)
		result <- err
	}(reader)
	err := <-result
	return data, err
}

// SayHello Say hello to trustedNode
func (ns *NetService) SayHello(bootNode ma.Multiaddr) error {
	node := ns.node
	bootAddr, bootID, err := parseAddressFromMultiaddr(bootNode)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"bootNode": bootNode,
			"error":    err,
		}).Error("parse Address from trustedNode failed")
		return err
	}
	node.bootIds = append(node.bootIds, bootID.Pretty())
	node.peerstore.AddAddr(
		bootID,
		bootAddr,
		peerstore.ProviderAddrTTL,
	)
	if node.host.Addrs()[0].String() != bootAddr.String() {
		if err := ns.Hello(bootID); err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"bootNode": bootNode,
				"error":    err,
			}).Error("say hello to bootNode failed")
			return errors.New("say hello to bootNode failed")
		}
		logging.CLog().WithFields(logrus.Fields{
			"bootNode": bootNode,
		}).Info("say hello to a node success")
		node.peerstore.AddAddr(
			bootID,
			bootAddr,
			peerstore.PermanentAddrTTL)
		// Update the routing table.
		node.routeTable.Update(bootID)
	}
	return nil
}

func parseAddressFromMultiaddr(address ma.Multiaddr) (ma.Multiaddr, peer.ID, error) {

	addr, err := ma.NewMultiaddr(
		strings.Split(address.String(), "/ipfs/")[0],
	)
	if err != nil {
		return nil, "", err
	}

	b58, err := address.ValueForProtocol(ma.P_IPFS)
	if err != nil {
		return nil, "", err
	}

	id, err := peer.IDB58Decode(b58)
	if err != nil {
		return nil, "", err
	}

	return addr, id, nil

}

// InArray returns whether an object exists in an array.
func InArray(obj interface{}, array interface{}) bool {
	arrayValue := reflect.ValueOf(array)
	if reflect.TypeOf(array).Kind() == reflect.Array || reflect.TypeOf(array).Kind() == reflect.Slice {
		for i := 0; i < arrayValue.Len(); i++ {
			if arrayValue.Index(i).Interface() == obj {
				return true
			}
		}
	}
	return false
}
