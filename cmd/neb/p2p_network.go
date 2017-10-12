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

package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/multiformats/go-multiaddr"
	nnet "github.com/nebulasio/go-nebulas/components/net"
	"github.com/nebulasio/go-nebulas/components/net/p2p"
	"github.com/nebulasio/go-nebulas/consensus"
	"github.com/nebulasio/go-nebulas/consensus/pow"
	"github.com/nebulasio/go-nebulas/core"
	log "github.com/sirupsen/logrus"
)

// GoP2p start p2p network
func GoP2p(seed string, port uint) {
	quitCh := make(chan bool, 1)
	nmCh := make(chan nnet.Manager, 1)
	config := p2p.DefautConfig()
	config.IP = localHost()
	if len(seed) > 0 {
		seed, err := multiaddr.NewMultiaddr(seed)
		if err != nil {
			log.Error("param seed error, creating seed node fail", err)
			return
		}
		config.BootNodes = []multiaddr.Multiaddr{seed}
	}
	if port > 0 {
		config.Port = port
	}

	// P2P network randseed, in this release we use port as randseed
	// config.Randseed = time.Now().Unix()
	config.Randseed = int64(config.Port)

	go runP2p(config, quitCh, nmCh)

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		quitCh <- true
		os.Exit(1)
	}()
}

func runP2p(config *p2p.Config, quitCh chan bool, nmCh chan nnet.Manager) {
	nm := p2p.NewManager(config)
	nmCh <- nm

	bc := core.NewBlockChain(core.TestNetID)
	fmt.Printf("chainID is %d\n", bc.ChainID())
	bc.BlockPool().RegisterInNetwork(nm)

	var cons consensus.Consensus
	cons = pow.NewPow(bc, nm)
	bc.SetConsensusHandler(cons)

	// start.
	nm.Start()
	bc.BlockPool().Start()
	cons.Start()

	<-quitCh

	// stop
	cons.Stop()
	bc.BlockPool().Stop()
	nm.Stop()
}

func localHost() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	return ""
}
