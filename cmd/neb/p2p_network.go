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
	"os"
	"os/signal"
	"syscall"

	"github.com/nebulasio/go-nebulas/neblet"
	"github.com/nebulasio/go-nebulas/neblet/pb"
)

// GoP2p start p2p network
func GoP2p(config *nebletpb.Config) {
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
