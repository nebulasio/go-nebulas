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

	"github.com/nebulasio/go-nebulas/neblet"
	"github.com/nebulasio/go-nebulas/util/logging"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	version   string
	commit    string
	branch    string
	compileAt string
	config    string
)

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
		cli.StringFlag{
			Name:        "config, c",
			Usage:       "load configuration from `FILE`",
			Value:       "seed.pb.txt",
			Destination: &config,
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))

	app.Commands = []cli.Command{
		accountCommand,
		consoleCommand,
		networkCommand,
	}
	sort.Sort(cli.CommandsByName(app.Commands))

	app.Run(os.Args)
}

func neb(ctx *cli.Context) error {
	logging.EnableFuncNameLogger()

	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	n := makeNeb(ctx)

	runNeb(n)

	// TODO: just use the signal to block main.
	for {
		time.Sleep(60 * time.Second) // or runtime.Gosched() or similar per @misterbee
	}
}

func runNeb(n *neblet.Neblet) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	if err := n.Start(); err != nil {
		panic("Start Neblet Failed: " + err.Error())
	}

	go func() {
		<-c
		n.Stop()

		// TODO: remove this once p2pManager handles stop properly.
		os.Exit(1)
	}()
}

func makeNeb(ctx *cli.Context) *neblet.Neblet {
	conf := neblet.LoadConfig(config)
	n := neblet.New(*conf)
	return n
}

// FatalF fatal format err
func FatalF(format string, args ...interface{}) {
	err := fmt.Sprintf(format, args...)
	fmt.Println(err)
	os.Exit(1)
}
