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
	u "github.com/ipfs/go-ipfs-util"
	gnet "github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	log "github.com/sirupsen/logrus"
	"io"
	"time"
)

const PingSize = 32

const pingProtocolID = "/nebulas/ping/1.0.0"

const pingTimeout = time.Second * 60

type PingService struct {
	node *Node
}

// register ping service
func (node *Node) RegisterPingService() *PingService {
	ps := &PingService{node}
	node.host.SetStreamHandler(pingProtocolID, ps.PingHandler)
	log.Infof("node register ping service success...")
	return ps
}

// handle others ping
func (p *PingService) PingHandler(s gnet.Stream) {
	log.Infof("node handle ping request...")
	buf := make([]byte, PingSize)

	errCh := make(chan error, 1)
	defer close(errCh)
	timer := time.NewTimer(pingTimeout)
	defer timer.Stop()

	go func() {
		select {
		case <-timer.C:
			log.Info("ping timeout")
			s.Reset()
		case err, ok := <-errCh:
			if ok {
				log.Debug(err)
				if err == io.EOF {
					s.Close()
				} else {
					s.Reset()
				}
			} else {
				log.Error("ping loop failed without error")
			}
		}
	}()

	for {
		_, err := io.ReadFull(s, buf)
		if err != nil {
			errCh <- err
			return
		}

		_, err = s.Write(buf)
		if err != nil {
			errCh <- err
			return
		}

		timer.Reset(pingTimeout)
	}
}

//Ping a peer
func (node *Node) Ping(pid peer.ID) error {
	log.Infof("Ping: node start ping peer %s", node.host.Addrs(), pid)
	ctx := node.context
	s, err := node.host.NewStream(ctx, pid, pingProtocolID)
	if err != nil {
		return err
	}

	out := make(chan time.Duration)
	go func() {
		defer close(out)
		defer s.Close()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				t, err := ping(s)
				if err != nil {
					s.Reset()
					log.Errorf("ping error: %s", err)
					return
				}

				node.host.Peerstore().RecordLatency(pid, t)
				select {
				case out <- t:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return nil
}

func ping(s gnet.Stream) (time.Duration, error) {
	buf := make([]byte, PingSize)
	u.NewTimeSeededRand().Read(buf)

	before := time.Now()
	_, err := s.Write(buf)
	if err != nil {
		return 0, err
	}

	rbuf := make([]byte, PingSize)
	_, err = io.ReadFull(s, rbuf)
	if err != nil {
		return 0, err
	}

	if !bytes.Equal(buf, rbuf) {
		return 0, errors.New("ping packet was incorrect")
	}

	return time.Since(before), nil
}
