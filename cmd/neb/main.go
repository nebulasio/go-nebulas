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

	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/consensus/pow"
	"github.com/nebulasio/go-nebulas/net"
	"github.com/nebulasio/go-nebulas/net/messages"

	log "github.com/sirupsen/logrus"
)

func run(sharedBlockCh chan *core.Block, quitCh chan bool, nmCh chan *net.NetManager) {
	nm := net.NewNetManager(sharedBlockCh)
	nmCh <- nm

	bc := core.NewBlockChain(core.TestNetID)
	fmt.Printf("chainID is %d\n", bc.GetChainID())

	// p := pow.NewPow(bc)
	p := pow.NewPow(bc, nm)

	// start Pow.
	p.Start()

	// start net.
	nm.Start()

	<-quitCh
	nm.Stop()
	p.Stop()
}

func replicateNewBlock(sharedBlockCh chan *core.Block, quitCh chan bool, nmCh chan *net.NetManager) {
	nms := make([]*net.NetManager, 0, 10)

	for {
		select {
		case block := <-sharedBlockCh:
			msg := messages.NewBlockMessage(block)
			for _, nm := range nms {
				nm.ReceiveMessage(msg)
			}
		case nm := <-nmCh:
			nms = append(nms, nm)
		case <-quitCh:
			return
		}
	}
}

func main() {
	log.SetFormatter(&log.TextFormatter{ForceColors: true, FullTimestamp: true})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	fmt.Println("starting...")
	quitCh := make(chan bool, 10)

	clientCount := 5
	nmCh := make(chan *net.NetManager, clientCount)

	sharedBlockCh := make(chan *core.Block, 50)
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

	fmt.Println("running...")
	fmt.Println("prss Ctrl+C to quit...")

	for {
		time.Sleep(60 * time.Second) // or runtime.Gosched() or similar per @misterbee
	}
}
