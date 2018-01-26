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

	app.Flags = append(app.Flags, ConfigFlag)
	app.Flags = append(app.Flags, NetworkFlags...)
	app.Flags = append(app.Flags, ChainFlags...)
	app.Flags = append(app.Flags, RPCFlags...)
	app.Flags = append(app.Flags, AppFlags...)
	app.Flags = append(app.Flags, StatsFlags...)

	sort.Sort(cli.FlagsByName(app.Flags))

	app.Commands = []cli.Command{
		initCommand,
		genesisCommand,
		accountCommand,
		consoleCommand,
		networkCommand,
		versionCommand,
		licenseCommand,
		configCommand,
		blockDumpCommand,
		serializeCommand,
	}
	sort.Sort(cli.CommandsByName(app.Commands))

	app.Run(os.Args)
}

func neb(ctx *cli.Context) error {
	n, err := makeNeb(ctx)
	if err != nil {
		return err
	}

	logging.Init(n.Config().App.LogFile, n.Config().App.LogLevel, n.Config().App.LogAge)

	// enable crash report if open the switch and configure the url
	if n.Config().App.EnableCrashReport && len(n.Config().App.CrashReportUrl) > 0 {
		InitCrashReporter(n.Config().App)
	}

	select {
	case <-runNeb(ctx, n):
		return nil
	}
}

func runNeb(ctx *cli.Context, n *neblet.Neblet) chan bool {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// start net pprof if config.App.Pprof.HttpListen configured
	err := n.StartPprof(n.Config().App.Pprof.HttpListen)
	if err != nil {
		FatalF("start pprof failed:%s", err)
	}

	n.Setup()
	n.Start()

	quitCh := make(chan bool, 1)

	go func() {
		<-c

		n.Stop()

		quitCh <- true
		return
	}()

	return quitCh
}

func makeNeb(ctx *cli.Context) (*neblet.Neblet, error) {
	conf := neblet.LoadConfig(config)
	conf.App.Version = version

	// load config from cli args
	networkConfig(ctx, conf.Network)
	chainConfig(ctx, conf.Chain)
	rpcConfig(ctx, conf.Rpc)
	appConfig(ctx, conf.App)
	statsConfig(ctx, conf.Stats)

	n, err := neblet.New(conf)
	if err != nil {
		return nil, err
	}
	return n, nil
}

// FatalF fatal format err
func FatalF(format string, args ...interface{}) {
	err := fmt.Sprintf(format, args...)
	fmt.Println(err)
	os.Exit(1)
}
