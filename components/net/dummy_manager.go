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
	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/crypto/keystore"
)

// DummyManager support dummy mode
type DummyManager struct {
	sharedBlockCh chan interface{}
	dispatcher    *Dispatcher
	keystore      *keystore.Keystore
}

// NewDummyManager create DummyManager instance.
func NewDummyManager(sharedBlockCh chan interface{}) *DummyManager {
	nm := &DummyManager{
		sharedBlockCh: sharedBlockCh,
		dispatcher:    NewDispatcher(),
		keystore:      keystore.NewKeystore(),
	}
	return nm
}

// Register register the subscribers.
func (nm *DummyManager) Register(subscribers ...*Subscriber) {
	nm.dispatcher.Register(subscribers...)
}

// Deregister Deregister the subscribers.
func (nm *DummyManager) Deregister(subscribers ...*Subscriber) {
	nm.dispatcher.Deregister(subscribers...)
}

// Start start net services.
func (nm *DummyManager) Start() {
	nm.dispatcher.Start()
}

// Stop stop net services.
func (nm *DummyManager) Stop() {
	nm.dispatcher.Stop()
}

// PutMessage put message to dispatcher.
func (nm *DummyManager) PutMessage(msg Message) {
	nm.dispatcher.PutMessage(msg)
}

// Broadcast network.
func (nm *DummyManager) Broadcast(name string, msg proto.Message) {
	// TODO(@leon): dispather
	nm.sharedBlockCh <- msg
}
