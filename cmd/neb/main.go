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
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nebulasio/go-nebulas/consensus"

	"github.com/nebulasio/go-nebulas/components/net"
	"github.com/nebulasio/go-nebulas/components/net/messages"
	"github.com/nebulasio/go-nebulas/consensus/pow"
	"github.com/nebulasio/go-nebulas/core"

	log "github.com/sirupsen/logrus"
)

func run(sharedBlockCh chan interface{}, quitCh chan bool, nmCh chan *net.Manager) {
	nm := net.NewManager(sharedBlockCh)
	nmCh <- nm

	bc := core.NewBlockChain(core.TestNetID)
	fmt.Printf("chainID is %d\n", bc.ChainID())
	bc.BlockPool().RegisterInNetwork(nm)

	var cons consensus.Consensus
	cons = pow.NewPow(bc, nm)

	// start.
	cons.Start()
	bc.BlockPool().Start()
	nm.Start()

	<-quitCh

	// stop
	nm.Stop()
	bc.BlockPool().Stop()
	cons.Stop()
}

func replicateNewBlock(sharedBlockCh chan interface{}, quitCh chan bool, nmCh chan *net.Manager) {
	nms := make([]*net.Manager, 0, 10)

	count := 0
	for {
		select {
		case block := <-sharedBlockCh:
			count++
			log.Info("replicateNewBlock: repBlockCount = ", count)
			msg := messages.NewBaseMessage(net.MessageTypeNewBlock, block)
			for _, nm := range nms {
				nm.PutMessage(msg)
			}
		case nm := <-nmCh:
			nms = append(nms, nm)
		case <-quitCh:
			return
		}
	}
}

func main() {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	quitCh := make(chan bool, 10)

	clientCount := 2
	nmCh := make(chan *net.Manager, clientCount)

	sharedBlockCh := make(chan interface{}, 50)
	go replicateNewBlock(sharedBlockCh, quitCh, nmCh)

	for i := 0; i < clientCount; i++ {
		go run(sharedBlockCh, quitCh, nmCh)
	}

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		for i := 0; i < clientCount; i++ {
			quitCh <- true
		}
		os.Exit(1)
	}()

	for {
		time.Sleep(60 * time.Second) // or runtime.Gosched() or similar per @misterbee
	}
}
