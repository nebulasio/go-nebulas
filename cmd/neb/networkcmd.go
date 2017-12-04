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
	"encoding/base64"
	"fmt"

	"github.com/libp2p/go-libp2p-crypto"
	"github.com/nebulasio/go-nebulas/account"
	"github.com/nebulasio/go-nebulas/net/p2p"
	"github.com/urfave/cli"
)

var (
	networkCommand = cli.Command{
		Name:     "network",
		Usage:    "Manage network",
		Category: "NETWORK COMMANDS",
		Description: `
Manage neblas network, generate a private key for node.`,

		Subcommands: []cli.Command{
			{
				Name:   "ssh-keygen",
				Usage:  "Generate a private key for network node",
				Action: generatePrivateKey,
				Description: `

Generate a private key for network node.

If the private key of a network node is exist, the nodeID will not change.

Make sure that the seed node should have a private key.`,
			},
		},
	}
)

// accountCreate creates a new account into the keystore
func generatePrivateKey(ctx *cli.Context) error {
	priv, _, err := p2p.GenerateEd25519Key()
	privb, err := crypto.MarshalPrivateKey(priv)
	fmt.Printf("priv: %s\n", base64.StdEncoding.EncodeToString(privb))

	account.WriteFile("id_ed25519", []byte(base64.StdEncoding.EncodeToString(privb)))
	return err
}
