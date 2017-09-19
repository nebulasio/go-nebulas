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

// Manager the net manager.
type Manager struct {
	sharedBlockCh chan interface{}
	dispatcher    *Dispatcher
}

// NewManager create Manager instance.
// TODO: remove sharedBlockCH, using underlying network lib instead.
func NewManager(sharedBlockCh chan interface{}) *Manager {
	nm := &Manager{
		sharedBlockCh: sharedBlockCh,
		dispatcher:    NewDispatcher(),
	}
	return nm
}

// Register register the subscribers.
func (nm *Manager) Register(subscribers ...*Subscriber) {
	nm.dispatcher.Register(subscribers...)
}

// Deregister Deregister the subscribers.
func (nm *Manager) Deregister(subscribers ...*Subscriber) {
	nm.dispatcher.Deregister(subscribers...)
}

// Start start net services.
func (nm *Manager) Start() {
	nm.dispatcher.Start()
}

// Stop stop net services.
func (nm *Manager) Stop() {
	nm.dispatcher.Stop()
}

// PutMessage put message to dispatcher.
func (nm *Manager) PutMessage(msg Message) {
	nm.dispatcher.PutMessage(msg)
}

// BroadcastBlock broadcast block to network.
func (nm *Manager) BroadcastBlock(block interface{}) {
	//TODO: broadcast block via underlying network lib to whole network.
	nm.sharedBlockCh <- block
}
