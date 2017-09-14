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

package net

import (
	"github.com/nebulasio/go-nebulas/core"
)

type Manager struct {
	sharedBlockCh chan *core.Block

	handlers   MessageHandlers
	dispatcher *Dispatcher
}

func NewManager(sharedBlockCh chan *core.Block) *Manager {
	nm := &Manager{
		sharedBlockCh: sharedBlockCh,
		handlers:      make(MessageHandlers),
		dispatcher:    NewDispatcher(),
	}
	// nm.receivedBlockCh = make(chan *core.Block)
	return nm
}

func (nm *Manager) Register(handlers ...MessageHandler) {
	for _, v := range handlers {
		nm.handlers[v] = true
		nm.dispatcher.Register(v)
	}
}

func (nm *Manager) Deregister(handlers ...MessageHandler) {
	for _, v := range handlers {
		delete(nm.handlers, v)
		nm.dispatcher.Deregister(v)
	}
}

func (nm *Manager) Start() {
	nm.dispatcher.Start()
}

func (nm *Manager) Stop() {
	nm.dispatcher.Stop()
}

func (nm *Manager) SendNewBlock(block *core.Block) {
	nm.sharedBlockCh <- block
}

func (nm *Manager) ReceiveMessage(msg Message) {
	nm.dispatcher.OnMessageReceived(msg)
}
