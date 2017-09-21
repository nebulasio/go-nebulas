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
	"sort"
	"strconv"
	"syscall"
	"time"

	"github.com/nebulasio/go-nebulas/consensus"
	"github.com/nebulasio/go-nebulas/util/logging"

	"github.com/nebulasio/go-nebulas/components/net"
	"github.com/nebulasio/go-nebulas/components/net/messages"
	"github.com/nebulasio/go-nebulas/components/net/p2p"
	"github.com/nebulasio/go-nebulas/consensus/pow"
	"github.com/nebulasio/go-nebulas/core"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	version   string
	commit    string
	branch    string
	compileAt string
	dummy     bool
	p2pConfig string
)

func run(sharedBlockCh chan interface{}, quitCh chan bool, nmCh chan net.Manager) {
	nm := net.NewDummyManager(sharedBlockCh)
	nmCh <- nm

	bc := core.NewBlockChain(core.TestNetID)
	fmt.Printf("chainID is %d\n", bc.ChainID())
	bc.BlockPool().RegisterInNetwork(nm)

	var cons consensus.Consensus
	cons = pow.NewPow(bc, nm)
	bc.SetConsensusHandler(cons)

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

func runP2p(config *p2p.Config, quitCh chan bool, nmCh chan net.Manager) {
	nm := p2p.NewP2pManager(config)
	nmCh <- nm

	bc := core.NewBlockChain(core.TestNetID)
	fmt.Printf("chainID is %d\n", bc.ChainID())
	bc.BlockPool().RegisterInNetwork(nm)

	var cons consensus.Consensus
	cons = pow.NewPow(bc, nm)
	bc.SetConsensusHandler(cons)

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

func replicateNewBlock(sharedBlockCh chan interface{}, quitCh chan bool, nmCh chan net.Manager) {
	nms := make([]net.Manager, 0, 10)

	count := 0
	for {
		select {
		case v := <-sharedBlockCh:
			count++
			log.Info("replicateNewBlock: repBlockCount = ", count)
			block := v.(*core.Block)
			for _, nm := range nms {
				data, _ := block.Serialize()
				nb := new(core.Block)
				nb.Deserialize(data)

				msg := messages.NewBaseMessage(net.MessageTypeNewBlock, nb)
				nm.(*net.DummyManager).PutMessage(msg)
			}
		case nm := <-nmCh:
			nms = append(nms, nm)
		case <-quitCh:
			return
		}
	}
}

func neb(ctx *cli.Context) error {
	fmt.Printf("dummy: %v,  config: %v\n", dummy, p2pConfig)
	logging.EnableFuncNameLogger()

	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	if dummy {
		godummy()
	} else {
		goP2p()
	}
	for {
		time.Sleep(60 * time.Second) // or runtime.Gosched() or similar per @misterbee
	}
}

func godummy() {
	quitCh := make(chan bool, 10)

	clientCount := 4
	nmCh := make(chan net.Manager, clientCount)

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
}

func goP2p() {
	quitCh := make(chan bool, 1)
	nmCh := make(chan net.Manager, 1)
	config := p2p.DefautConfig()
	go runP2p(config, quitCh, nmCh)

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		quitCh <- true
		os.Exit(1)
	}()
}

func main() {
	app := cli.NewApp()
	app.Action = neb
	app.Name = "neb"
	app.Version = fmt.Sprintf("%s, branch %s, commit %s", version, branch, commit)
	timestamp, _ := strconv.ParseInt(compileAt, 10, 64)
	app.Compiled = time.Unix(timestamp, 0)
	app.Usage = "the go-nebulas command line interface"
	app.Copyright = "Copyright 2017-2018 The go-nebulas Authors"

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:        "dummy",
			Usage:       "use dummy network",
			Destination: &dummy,
		},
		cli.StringFlag{
			Name:        "config, c",
			Usage:       "load configuration from `FILE`",
			Destination: &p2pConfig,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "config",
			Aliases: []string{"c"},
			Usage:   "Show default configuration",
			Action: func(c *cli.Context) error {
				return nil
			},
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	app.Run(os.Args)
}
