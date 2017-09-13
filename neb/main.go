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

	"github.com/nebulasio/go-nebulas/blockchain"
	"github.com/nebulasio/go-nebulas/consensus/pow"
)

func main() {
	fmt.Println("starting...")

	bc := blockchain.NewBlockChain(blockchain.TestNetID)
	fmt.Printf("chainID is %d\n", bc.GetChainID())

	p := pow.NewPow(bc)
	p.Start()

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		p.Stop()
		os.Exit(1)
	}()

	fmt.Println("running...")
	fmt.Println("prss Ctrl+C to quit...")

	for {
		time.Sleep(60 * time.Second) // or runtime.Gosched() or similar per @misterbee
	}
}
