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

package rpc

import (
	"github.com/nebulasio/go-nebulas/account"
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/neblet/pb"
	"github.com/nebulasio/go-nebulas/net/p2p"
)

// Neblet interface breaks cycle import dependency and hides unused services.
type Neblet interface {
	Config() nebletpb.Config
	BlockChain() *core.BlockChain
	AccountManager() *account.Manager
	NetService() *p2p.NetService
	EventEmitter() *core.EventEmitter
}

// Server server interface for api & management etc.
type Server interface {
	// Start start server
	Start() error

	// Stop stop server
	Stop()

	// Neblet return neblet
	Neblet() Neblet

	RunGateway() error
}
