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
	"sort"
	"strconv"
	"time"

	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/crypto/keystore/ecdsa"
	"github.com/nebulasio/go-nebulas/crypto/keystore/key"
	"github.com/nebulasio/go-nebulas/util/logging"
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
	port      uint
	seed      string
)

func main() {

	//TODO(larry.wang):add test addres to keystore,later remove
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
		cli.StringFlag{
			Name:        "seed, s",
			Usage:       "p2p network seed node address",
			Destination: &seed,
		},
		cli.UintFlag{
			Name:        "port, p",
			Usage:       "p2p network port",
			Destination: &port,
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

func neb(ctx *cli.Context) error {
	logging.EnableFuncNameLogger()

	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	if dummy {
		GoDummy()
	} else {
		GoP2p(seed, port)
	}
	for {
		time.Sleep(60 * time.Second) // or runtime.Gosched() or similar per @misterbee
	}
}

// add test address to keystore
func addTestAddress() {
	ks := keystore.DefaultKS
	arr := make([][]byte, 3)
	arr[0] = []byte{59, 144, 87, 239, 199, 27, 51, 230, 209, 177, 177, 166, 161, 23, 23, 195, 197, 245, 56, 156, 171, 40, 209, 7, 25, 1, 32, 0, 75, 69, 145, 30}
	arr[1] = []byte{208, 98, 189, 16, 69, 97, 14, 44, 112, 56, 253, 61, 195, 100, 88, 245, 99, 14, 70, 22, 173, 172, 243, 186, 46, 128, 18, 39, 93, 125, 27, 186}
	arr[2] = []byte{217, 81, 120, 192, 22, 101, 123, 205, 222, 253, 237, 63, 248, 9, 226, 102, 97, 202, 124, 1, 248, 178, 7, 69, 14, 63, 254, 127, 61, 158, 126, 65}
	for _, pdata := range arr {
		priv, _ := ecdsa.ToPrivateKey(pdata)
		adata, _ := ecdsa.ToAddressData(priv)
		addr, _ := core.NewAddress(adata)
		ps := ecdsa.NewPrivateStoreKey(priv)
		ks.SetKeyPassphrase(key.Alias(addr.ToHex()), ps, []byte("passphrase"))
		pass, _ := key.NewPassphrase([]byte("passphrase"))
		ks.Unlock(key.Alias(addr.ToHex()), pass)
	}
}
