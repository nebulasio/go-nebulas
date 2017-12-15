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

	"github.com/nebulasio/go-nebulas/neblet"
	"github.com/urfave/cli"
)

var (
	configCommand = cli.Command{
		Name:     "config",
		Usage:    "Manage config",
		Category: "CONFIG COMMANDS",
		Description: `
Manage neblas config, generate a default config file.`,

		Subcommands: []cli.Command{
			{
				Name:      "new",
				Usage:     "Generate a default config file",
				Action:    MergeFlags(createDefaultConfig),
				ArgsUsage: "<filename>",
				Description: `
Generate a a default config file.`,
			},
		},
	}
)

// accountCreate creates a new account into the keystore
func createDefaultConfig(ctx *cli.Context) error {
	fileName := ctx.Args().First()
	if len(fileName) == 0 {
		fmt.Println("please give a config file arg!!!")
		return nil
	}
	neblet.CreateDefaultConfigFile(fileName)
	fmt.Printf("create default config %s\n", fileName)
	return nil
}
