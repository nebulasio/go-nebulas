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
	"io/ioutil"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"syscall"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/crypto/keystore/secp256k1"
	"github.com/nebulasio/go-nebulas/neblet"
	"github.com/nebulasio/go-nebulas/neblet/pb"
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

	// TODO(larry.wang):add test address to keystore,later remove
	addTestAddress()

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
			Value:       "config.pb.txt",
			Destination: &config,
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	app.Run(os.Args)
}

func neb(ctx *cli.Context) error {
	logging.EnableFuncNameLogger()

	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	conf := loadConfig(config)

	runNeb(conf)

	// TODO: just use the signal to block main.
	for {
		time.Sleep(60 * time.Second) // or runtime.Gosched() or similar per @misterbee
	}
}

func runNeb(config *nebletpb.Config) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	n := neblet.New(*config)
	n.Start()

	go func() {
		<-c
		n.Stop()

		// TODO: remove this once p2pManager handles stop properly.
		os.Exit(1)
	}()
}

func loadConfig(filename string) *nebletpb.Config {
	log.Info("Loading Neb config from file ", filename)
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	str := string(b)
	log.Info("Parsing Neb config text ", str)

	pb := new(nebletpb.Config)
	if err := proto.UnmarshalText(str, pb); err != nil {
		log.Fatal(err)
	}
	log.Info("Loaded Neb config proto ", pb)
	return pb
}

// add test address to keystore
func addTestAddress() {
	ks := keystore.DefaultKS
	for i := 0; i < 3; i++ {
		priv, _ := secp256k1.GeneratePrivateKey()
		pubdata, _ := priv.PublicKey().Encoded()
		addr, _ := core.NewAddressFromPublicKey(pubdata)
		ks.SetKey(addr.ToHex(), priv, []byte("passphrase"))
		ks.Unlock(addr.ToHex(), []byte("passphrase"), time.Second*60*60*24*365)
	}
}
