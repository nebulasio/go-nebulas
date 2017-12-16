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
	"github.com/nebulasio/go-nebulas/cmd/console"
	"github.com/urfave/cli"
)

var (
	consoleCommand = cli.Command{
		Action:   MergeFlags(consoleStart),
		Name:     "console",
		Usage:    "Start an interactive JavaScript console",
		Category: "CONSOLE COMMANDS",
		Description: `
The Neb console is an interactive shell for the JavaScript runtime environment.`,
	}
)

func consoleStart(ctx *cli.Context) error {
	neb, err := makeNeb(ctx)
	if err != nil {
		return err
	}

	console := console.New(neb)
	console.Setup()
	console.Interactive()
	defer console.Stop()
	return nil
}
