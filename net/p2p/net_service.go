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
	metrics "github.com/rcrowley/go-metrics"
	log "github.com/sirupsen/logrus"
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
	packetInFromNet = metrics.GetOrRegisterMeter("packet_in_from_net", nil)
	packetOut       = metrics.GetOrRegisterMeter("packet_out", nil)
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
type Protocol struct {
	magicNumber    []byte
	chainID        []byte
	version        byte
	msgName        string
	dataLength     []byte
	dataChecksum   []byte
	headerChecksum []byte
	dataHeader     []byte
	data           []byte
}

// NewNetService create netService
func NewNetService(n Neblet) (*NetService, error) {
	config := NewP2PConfig(n)
	node, err := NewNode(config)
	if err != nil {
		log.Error("NewNetService: node create fail -> ", err)
		return nil, err
	}
	ns := &NetService{node, make(chan bool), net.NewDispatcher()}
	return ns, nil
}

func (ns *NetService) registerNetService() *NetService {
	// register streamHandler to start loop to handle stream origined from remote node.
	ns.node.host.SetStreamHandler(ProtocolID, ns.streamHandler)
	log.Infof("RegisterNetService: register netservice success")
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
	for {
		select {
		case <-ns.quitCh:
			return
		default:
			node := ns.node
			pid := s.Conn().RemotePeer()
			addrs := s.Conn().RemoteMultiaddr()
			key := pid.Pretty()
			protocol, err := ns.parse(s)
			if err != nil {
				log.Error("streamHandler: parse network protocol occurs error, ", err)
				ns.Bye(pid, []ma.Multiaddr{addrs}, s, key)
				return
			}

			switch protocol.msgName {
			case HELLO:
				ns.handleHelloMsg(protocol.data, pid, s, addrs, key)
			case OK:
				ns.handleOkMsg(protocol.data, pid, s, addrs, key)
			case BYE:

			case SyncRoute:
				ns.handleSyncRouteMsg(protocol.data, pid, s, addrs, key)
			case SyncRouteReply:
				ns.handleSyncRouteReplyMsg(protocol.data, pid, s, addrs)
			case NewHashMsg:
				ns.handleNewHashMsg(protocol.data, pid)
			default:
				var relayness []peer.ID
				streamStore, ok := node.stream.Load(key)
				if !ok {
					ns.Bye(pid, []ma.Multiaddr{addrs}, s, key)
					return
				}
				if streamStore.(*StreamStore).conn != SOK {
					log.Error("peer not shake hand before send message.")
					ns.Bye(pid, []ma.Multiaddr{addrs}, s, key)
					return
				}
				msg := messages.NewBaseMessage(protocol.msgName, protocol.data)
				ns.PutMessage(msg)
				packetInFromNet.Mark(1)
				peers, exists := node.relayness.Get(byteutils.Uint32(protocol.dataChecksum))
				if exists {
					relayness = peers.([]peer.ID)
				}
				node.relayness.Add(byteutils.Uint32(protocol.dataChecksum), append(relayness, pid))
			}
		}
	}

}

func (ns *NetService) parse(s libnet.Stream) (*Protocol, error) {

	header, err := ReadBytes(s, uint32(offsetThirtySix))
	if err != nil {
		log.Error("parse protocol, read data header occurs error, ", err)
		return nil, err
	}

	protocol := &Protocol{}
	protocol.magicNumber = header[:offsetFour]
	protocol.chainID = header[offsetFour:offsetEight]
	protocol.version = header[offsetEleven]
	msgName := header[offsetTwelve:offsetTwentyFour]
	protocol.dataLength = header[offsetTwentyFour:offsetTwentyEight]
	protocol.dataChecksum = header[offsetTwentyEight:offsetThirtyTwo]
	protocol.dataHeader = header[:offsetThirtyTwo]

	index := bytes.IndexByte(msgName, 0)
	msgNameByte := msgName[0:index]
	protocol.msgName = string(msgNameByte)

	protocol.headerChecksum = header[offsetThirtyTwo:offsetThirtySix]

	if !ns.verifyHeader(protocol) {
		return nil, errors.New("parse protocol, verify header occurs error")
	}

	data, err := ReadBytes(s, byteutils.Uint32(protocol.dataLength))
	if err != nil {
		log.Error("parse protocol, read data occurs error, ", err)
		return nil, err
	}
	protocol.data = data

	dataChecksumA := crc32.ChecksumIEEE(data)
	if dataChecksumA != byteutils.Uint32(protocol.dataChecksum) {
		log.WithFields(log.Fields{
			"dataChecksumA": dataChecksumA,
			"dataChecksum":  byteutils.Uint32(protocol.dataChecksum),
		}).Error("parse protocol, data verification occurs error, dataChecksum is error, the connection will be closed.")
		return nil, errors.New("parse protocol, data verification occurs error, dataChecksum is error")
	}

	log.WithFields(log.Fields{
		"msgName":      protocol.msgName,
		"magicNumber":  string(protocol.magicNumber),
		"chainID":      byteutils.Uint32(protocol.chainID),
		"version":      protocol.version,
		"dataChecksum": byteutils.Uint32(protocol.dataChecksum),
	}).Info("parse protocol header data.")

	return protocol, nil

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
		log.Error("handle hello msg occurs error: ", err)
		return result
	}
	if err := hello.FromProto(pb); err != nil {
		log.Error("handle hello msg occurs error: ", err)
		return result
	}

	log.WithFields(log.Fields{
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
			log.Error("handleHelloMsg send ok message occurs error, ", err)
			return result
		}

		node.peerstore.AddAddr(
			pid,
			addrs,
			peerstore.PermanentAddrTTL,
		)

		totalData := ns.buildData(okdata, OK)

		if err := Write(s, totalData); err != nil {
			log.Error("handleHelloMsg write data occurs error, ", err)
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
	log.Debug("handle ok message")
	result := false
	defer func() {
		if !result {
			ns.Bye(pid, []ma.Multiaddr{addrs}, s, key)
		}
	}()

	ok := new(messages.HelloMessage)
	pb := new(netpb.Hello)
	if err := proto.Unmarshal(data, pb); err != nil {
		log.Error("handle ok msg occurs error: ", err)
		return result
	}
	if err := ok.FromProto(pb); err != nil {
		log.Error("handle ok msg occurs error: ", err)
		return result
	}

	if ok.NodeID == pid.String() && ok.ClientVersion == ClientVersion {
		streamStore := NewStreamStore(key, SOK, s)
		// node.stream[key] = streamStore
		node.stream.Store(key, streamStore)
		node.streamCache.Insert(streamStore)
		// node.conn[key] = SOK
		node.peerstore.AddAddr(
			pid,
			addrs,
			peerstore.PermanentAddrTTL,
		)
		node.routeTable.Update(pid)

		result = true
		return result
	}

	log.Error("handleOkMsg get incorrect response")
	return result

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
			log.WithFields(log.Fields{
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
	log.WithFields(log.Fields{
		"remoteId":    pid.Pretty(),
		"remoteAddrs": addrs,
		"count":       len(peerList),
	}).Debug("reply sync route to remote node")

	peersMessage := messages.NewPeersMessage(peerList)
	pb, err := peersMessage.ToProto()
	data, err = proto.Marshal(pb)
	if err != nil {
		log.Error("handleSyncRouteMsg occurs error, ", err)
		return result
	}

	totalData := ns.buildData(data, SyncRouteReply)

	// if _, ok := node.stream[key]; !ok {
	// 	log.Error("handleSyncRouteMsg occrus error, stream does not exist.")
	// 	return result
	// }
	streamStore, ok := node.stream.Load(key)
	if !ok {
		log.Error("handleSyncRouteMsg occrus error, stream does not exist.")
		return result
	}
	if err := Write(streamStore.(*StreamStore).stream, totalData); err != nil {
		log.Error("handleSyncRouteMsg write data occurs error, ", err)
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
		log.Error("handleSyncRouteReplyMsg occurs error: ", err)
		return false
	}
	if err := peers.FromProto(pb); err != nil {
		log.Error("handleSyncRouteReplyMsg occurs error: ", err)
		return false
	}

	for i := range peers.Peers() {
		id := peers.Peers()[i].ID()
		if node.routeTable.Find(id) != "" || len(peers.Peers()[i].Addrs()) == 0 {
			log.WithFields(log.Fields{
				"id": id.Pretty(),
			}).Warn("node is already exist in route table")
			continue
		}
		var addres []ma.Multiaddr
		for _, v := range peers.Peers()[i].Addrs() {
			addr, _ := ma.NewMultiaddr(v)
			addres = append(addres, addr)
		}

		// address, err := ma.NewMultiaddr(peers.Peers()[i].Addrs())
		// if err != nil {
		// 	log.WithFields(log.Fields{
		// 		"addrs": peers.Peers()[i].Addrs(),
		// 	}).Warn("parse address occurs error")
		// 	continue
		// }
		log.WithFields(log.Fields{
			"id":    id.Pretty(),
			"addrs": addres,
		}).Debug("discover new node")

		node.peerstore.AddAddrs(
			id,
			addres,
			peerstore.ProviderAddrTTL,
		)
		if err := ns.Hello(id); err != nil {
			log.WithFields(log.Fields{
				"id":  id.Pretty(),
				"err": err,
			}).Error("say hello to the peer fail")
			continue
		}
		// node.peerstore.AddAddr(
		// 	id,
		// 	address,
		// 	peerstore.PermanentAddrTTL,
		// )
		// Update the routing table.
		node.routeTable.Update(id)
	}
	return true
}

func (ns *NetService) verifyHeader(protocol *Protocol) bool {

	node := ns.node
	dataHeaderChecksum := crc32.ChecksumIEEE(protocol.dataHeader)

	if !byteutils.Equal(MagicNumber, protocol.magicNumber) {
		log.Error("verifyHeader: data verification occurs error, magic number is error, the connection will be closed.")
		return false
	}

	if node.chainID != byteutils.Uint32(protocol.chainID) {
		log.Error("verifyHeader: data verification occurs error, chainID is error, the connection will be closed.")
		return false
	}

	if node.version != protocol.version {
		log.Error("verifyHeader: data verification occurs error, version is error, the connection will be closed.")
		return false
	}

	if dataHeaderChecksum != byteutils.Uint32(protocol.headerChecksum) {
		log.Error("verifyHeader: data verification occurs error, dataHeaderChecksum is error, the connection will be closed.")
		return false
	}
	return true
}

// Bye say bye to a peer, and close connection.
func (ns *NetService) Bye(pid peer.ID, addrs []ma.Multiaddr, s libnet.Stream, key string) {
	node := ns.node
	ns.clearPeerStore(pid, addrs)
	// delete(node.stream, key)
	node.stream.Delete(key)
	s.Close()
}

func (ns *NetService) clearPeerStore(pid peer.ID, addrs []ma.Multiaddr) {
	node := ns.node
	node.peerstore.SetAddrs(pid, addrs, 0)
	node.routeTable.Remove(pid)
}

// SendMsg send message to a peer
func (ns *NetService) SendMsg(msgName string, msg []byte, key string) {
	node := ns.node
	log.WithFields(log.Fields{
		"key":     key,
		"msgName": msgName,
	}).Info("SendMsg: send message to a peer.")
	data := msg
	totalData := ns.buildData(data, msgName)

	// if _, ok := node.stream[key]; !ok {
	// 	log.Error("SendMsg: send message occrus error, stream does not exist.")
	// 	return
	// }
	// streamStore := node.stream[key]

	streamStore, ok := node.stream.Load(key)
	if !ok {
		log.Error("SendMsg: send message occrus error, stream does not exist.")
		return
	}

	if err := Write(streamStore.(*StreamStore).stream, totalData); err != nil {
		log.Error("SendMsg: write data occurs error, ", err)
		return
	}

	packetOut.Mark(1)

}

// Hello say hello to a peer
func (ns *NetService) Hello(pid peer.ID) error {
	msgName := HELLO
	node := ns.node
	stream, err := node.host.NewStream(
		node.context,
		pid,
		ProtocolID,
	)
	addrs := node.peerstore.PeerInfo(pid).Addrs
	if err != nil {
		log.Error("say hello occurs error, ", err)
		ns.clearPeerStore(pid, addrs)
		return err
	}
	if len(addrs) < 1 {
		log.Error("Hello: wrong pid addrs")
		ns.clearPeerStore(pid, addrs)
		return errors.New("wrong pid addrs")
	}

	hello := messages.NewHelloMessage(node.id.String(), ClientVersion)
	pb, _ := hello.ToProto()
	data, err := proto.Marshal(pb)
	if err != nil {
		return err
	}

	totalData := ns.buildData(data, msgName)
	if err := Write(stream, totalData); err != nil {
		log.Error("Hello: write data occurs error, ", err)
		return errors.New("Hello: write data occurs error")
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
		log.Error("SyncRoutes: wrong pid addrs")
		ns.clearPeerStore(pid, addrs)
		return
	}
	data := []byte{}
	totalData := ns.buildData(data, SyncRoute)
	key := pid.Pretty()

	// if _, ok := node.stream[key]; !ok {
	// 	log.Error("SyncRoutes: send message occrus error, stream does not exist.")
	// 	return
	// }

	// streamStore := node.stream[key]

	streamStore, ok := node.stream.Load(key)
	if !ok {
		log.Error("SyncRoutes: send message occrus error, stream does not exist.")
		return
	}
	if err := Write(streamStore.(*StreamStore).stream, totalData); err != nil {
		log.Error("SyncRoutes: write data occurs error, ", err)
		ns.clearPeerStore(pid, addrs)
		return
	}

}

// buildHeader build header information
func buildHeader(chainID uint32, msgName string, version byte, dataLength uint32, dataChecksum uint32) []byte {
	var metaHeader = make([]byte, offsetThirtyTwo)
	msgNameByte := []byte(msgName)

	copy(metaHeader[:], MagicNumber)
	copy(metaHeader[offsetFour:], byteutils.FromUint32(chainID))
	// 64-88 Reserved field
	copy(metaHeader[offsetEleven:], []byte{version})
	copy(metaHeader[offsetTwelve:], msgNameByte)
	copy(metaHeader[offsetTwentyFour:], byteutils.FromUint32(dataLength))
	copy(metaHeader[offsetTwentyEight:], byteutils.FromUint32(dataChecksum))

	return metaHeader
}

func (ns *NetService) buildData(data []byte, msgName string) []byte {
	node := ns.node
	dataChecksum := crc32.ChecksumIEEE(data)
	metaHeader := buildHeader(node.chainID, msgName, node.version, uint32(len(data)), dataChecksum)
	headerChecksum := crc32.ChecksumIEEE(metaHeader)
	metaHeader = append(metaHeader[:], byteutils.FromUint32(headerChecksum)...)
	totalData := append(metaHeader[:], data...)
	return totalData
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
	log.WithFields(log.Fields{
		"id":    node.ID(),
		"addrs": node.host.Addrs(),
	}).Info("node start")
	if node.running {
		return errors.New("net.start: node already running")
	}
	node.running = true

	ns.registerNetService()

	// TODO: All fail handle
	var success bool
	var wg sync.WaitGroup
	for _, bootNode := range node.config.BootNodes {
		wg.Add(1)
		go func(bootNode ma.Multiaddr) {
			defer wg.Done()
			err := ns.SayHello(bootNode)
			if err != nil {
				log.Error("net.start: can not say hello to trusted node.", bootNode, err)
			} else {
				success = true
			}

		}(bootNode)
	}
	wg.Wait()

	if success || len(node.Config().BootNodes) == 0 {
		go ns.discovery(node.context)
		go ns.manageStreamStore()
		log.Infof("net.start: node start and join to p2p network success and listening for connections on port %d... ", node.config.Port)
	} else {
		log.Error("net.start: node start occurs error, say hello to bootNode fail")
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
			node.peerstore.ClearAddrs(v)
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
		log.WithFields(log.Fields{
			"bootNode": bootNode,
			"error":    err,
		}).Error("parse Address from trustedNode failed")
		return err
	}
	node.peerstore.AddAddr(
		bootID,
		bootAddr,
		peerstore.TempAddrTTL,
	)
	if node.host.Addrs()[0].String() != bootAddr.String() {
		var success = false
		for i := 0; i < 3; i++ {
			err := ns.Hello(bootID)
			if err != nil {
				time.Sleep(time.Second)
				continue
			}
			success = true
			break
		}
		if !success {
			log.WithFields(log.Fields{
				"bootNode": bootNode,
				"error":    err,
			}).Error("say hello to bootNode failed")
			return errors.New("say hello to bootNode failed")
		}
		log.WithFields(log.Fields{
			"bootNode": bootNode,
		}).Debug("say hello to a node success")
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

// GenerateKey generate a key
// func GenerateKey(addrs ma.Multiaddr, pid peer.ID) (string, error) {
// 	if len(strings.Split(addrs.String(), "/")) > 2 {
// 		ip, err := stringIPToInt(strings.Split(addrs.String(), "/")[2])
// 		if err != nil {
// 			return "", err
// 		}
// 		key := fmt.Sprintf("%s:%d", pid.Pretty(), ip)
// 		return key, nil
// 	}
// 	log.WithFields(log.Fields{
// 		"addrs": addrs,
// 		"pid":   pid,
// 	}).Error("GenerateKey: the addrs format is incorrect.")
// 	// TODO return nil, error
// 	err := errors.New("GenerateKey: the addrs format is incorrect")
// 	return "", err
// }

// func stringIPToInt(ipstring string) (int, error) {
// 	ipSegs := strings.Split(ipstring, ".")
// 	if len(ipSegs) != 4 {
// 		return 0, errors.New("The IP format is not correct")
// 	}
// 	var ipInt int
// 	var pos uint = 24
// 	for _, ipSeg := range ipSegs {
// 		tempInt, _ := strconv.Atoi(ipSeg)
// 		tempInt = tempInt << pos
// 		ipInt = ipInt | tempInt
// 		pos -= 8
// 	}
// 	return ipInt, nil
// }
