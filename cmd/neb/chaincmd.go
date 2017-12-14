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
	"strconv"

	"github.com/nebulasio/go-nebulas/core"
	"github.com/urfave/cli"
)

var (
	initCommand = cli.Command{
		Action:    MergeFlags(initGenesis),
		Name:      "init",
		Usage:     "Bootstrap and initialize a new genesis block",
		ArgsUsage: "<genesisPath>",
		Category:  "BLOCKCHAIN COMMANDS",
		Description: `
The init command initializes a new genesis block and definition for the network.`,
	}

	blockDumpCommand = cli.Command{
		Action:    MergeFlags(dumpblock),
		Name:      "dump",
		Usage:     "Dump the number of newest block before tail block from storage",
		ArgsUsage: "<blocknumber>",
		Category:  "BLOCKCHAIN COMMANDS",
		Description: `
Use "./neb dump 10" to dump 10 blocks before tail block.`,
	}
)

func initGenesis(ctx *cli.Context) error {
	filePath := ctx.Args().First()
	genesis, err := core.LoadGenesisConf(filePath)
	if err != nil {
		return err
	}

	neb := makeNeb(ctx)
	neb.SetGenesis(genesis)

	return neb.Setup()
}

func dumpblock(ctx *cli.Context) error {
	neb := makeNeb(ctx)
	if err := neb.Setup(); err != nil {
		return err
	}
	count, err := strconv.Atoi(ctx.Args().First())
	if err != nil {
		return err
	}
	fmt.Printf("blockchain dump: %s\n", neb.BlockChain().Dump(count))
	return nil
}
