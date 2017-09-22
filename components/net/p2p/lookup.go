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
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/libp2p/go-libp2p-kbucket"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-peerstore"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	log "github.com/sirupsen/logrus"
)

const lookupProtocolID = "/nebulas/lookup/1.0.0"

// LookupService is used to explore other node
type LookupService struct {
	node *Node
}

// RegisterLookupService register lookup service
func (node *Node) RegisterLookupService() *LookupService {
	ls := &LookupService{node}
	node.host.SetStreamHandler(lookupProtocolID, ls.LookupHandler)
	log.Infof("RegisterLookupService: node register lookup service success...")
	return ls
}

// Lookup from a node
func (node *Node) Lookup(pid peer.ID) ([]peerstore.PeerInfo, error) {

	log.Infof("Lookup: starting lookup the node %s", pid)
	s, err := node.host.NewStream(node.context, pid, lookupProtocolID)
	if err != nil {
		log.Error("Lookup: node start lookup occurs error ", err)
		return nil, err
	}
	defer s.Close()

	timeout := 30 * time.Second

	size, err := ReadWithTimeout(s, 4, timeout)

	data, err := ReadWithTimeout(
		s,
		byteutils.Uint32(size),
		timeout,
	)

	var sample []peerstore.PeerInfo

	err = json.Unmarshal(data, &sample)
	log.Infof("Lookup: lookup the node %s success and get response...%s", pid, sample)
	return sample, nil
}

// ReadWithTimeout Read data from a stream using a timeout.
func ReadWithTimeout(reader io.Reader, n uint32, timeout time.Duration) ([]byte, error) {
	data := make([]byte, n)
	result := make(chan error, 1)
	go func(reader io.Reader) {
		_, err := io.ReadFull(reader, data)
		result <- err
	}(reader)
	select {
	case err := <-result:
		return data, err
	case <-time.After(timeout):
		select {
		case result <- errors.New("Timeout"):
		default:
		}
		err := <-result
		return data, err
	}
}

// LookupHandler handle lookup request
func (p *LookupService) LookupHandler(s net.Stream) {
	defer s.Close()
	pid := s.Conn().RemotePeer()
	log.Info("LookupHandler: Receiving lookup request from pid: ", pid)

	peers := p.node.routeTable.NearestPeers(kbucket.ConvertPeerID(pid), p.node.config.maxSyncNodes)

	var peerList []peerstore.PeerInfo
	for i := range peers {
		peerInfo := p.node.peerstore.PeerInfo(peers[i])
		peerList = append(peerList, peerInfo)
	}

	log.Info("LookupHandler: handle lookup request and return data...", peerList)
	timeout := 30 * time.Second
	data, err := json.Marshal(peerList)

	if err != nil {
		log.Error("LookupHandler: lookup handle occurs error...", err)
	}
	size := byteutils.FromUint32(uint32(len(data)))
	err = WriteWithTimeout(
		s,
		append(size[:], data...),
		timeout,
	)

	p.node.routeTable.Update(pid)
}

// WriteWithTimeout write data using a timeout
func WriteWithTimeout(writer io.Writer, data []byte, timeout time.Duration) error {
	result := make(chan error, 1)
	go func(writer io.Writer, data []byte) {
		_, err := writer.Write(data)
		result <- err
	}(writer, data)
	select {
	case err := <-result:
		return err
	case <-time.After(timeout):
		select {
		case result <- errors.New("Timeout"):
		default:
		}
		err := <-result
		return err
	}
}
