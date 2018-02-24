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
	"runtime"

	"github.com/nebulasio/go-nebulas/net"
	"github.com/urfave/cli"
)

var (
	versionCommand = cli.Command{
		Action:    MergeFlags(_version),
		Name:      "version",
		Usage:     "Print version numbers",
		ArgsUsage: " ",
		Category:  "MISC COMMANDS",
	}
	licenseCommand = cli.Command{
		Action:    MergeFlags(_license),
		Name:      "license",
		Usage:     "Display license information",
		ArgsUsage: " ",
		Category:  "MISC COMMANDS",
	}
)

func _version(ctx *cli.Context) error {
	neb, err := makeNeb(ctx)
	if err != nil {
		return err
	}

	fmt.Println("Version:", version)
	if commit != "" {
		fmt.Println("Git Commit:", commit)
	}
	fmt.Println("Protocol Versions:", net.NebProtocolID)
	fmt.Println("Protocol ClientVersion:", net.ClientVersion)
	fmt.Printf("Chain Id: %d\n", neb.Config().Chain.ChainId)
	fmt.Println("Go Version:", runtime.Version())
	fmt.Println("Operating System:", runtime.GOOS)
	fmt.Printf("GOPATH=%s\n", os.Getenv("GOPATH"))
	fmt.Printf("GOROOT=%s\n", runtime.GOROOT())
	return nil
}

func _license(_ *cli.Context) error {
	fmt.Println(`The preferred license for the Nebulas Open Source Project is the GNU Lesser General Public License Version 3.0 (“LGPL v3”), which is commercial friendly, and encourage developers or companies modify and publish their changes.

	However, we also aware that big corporations is favoured by other licenses, for example, Apache Software License 2.0 (“Apache v2.0”), which is more commercial friendly. For the Nebulas Team, we are very glad to see the source code and protocol of Nebulas is widely used both in open source applications and non-open source applications.

	In this way, we are still considering the license choice, which kind of license is the best for nebulas ecosystem. We expect to select one of the LGPL v3, the Apache v2.0 or the MIT license. If the latter is chosen, it will come with an amendment allowing it to be used more widely.`)
	return nil
}
