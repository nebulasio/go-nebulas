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
	"fmt"
	"sync"

	"github.com/nebulasio/go-nebulas/util/logging"
	metrics "github.com/rcrowley/go-metrics"
)

// Metrics map for different in/out network msg types
var (
	PacketsInByTypes  = new(sync.Map)
	PacketsOutByTypes = new(sync.Map)
)

// Dispatcher a message dispatcher service.
type Dispatcher struct {
	subscribersMap    *sync.Map
	quitCh            chan bool
	receivedMessageCh chan Message
}

// NewDispatcher create Dispatcher instance.
func NewDispatcher() *Dispatcher {
	dp := &Dispatcher{
		subscribersMap:    new(sync.Map),
		quitCh:            make(chan bool, 10),
		receivedMessageCh: make(chan Message, 1024),
	}

	return dp
}

// Register register subscribers.
func (dp *Dispatcher) Register(subscribers ...*Subscriber) {
	for _, v := range subscribers {
		for _, mt := range v.msgTypes {
			PacketsInByTypes.LoadOrStore(mt, metrics.GetOrRegisterMeter(fmt.Sprintf("neb.net.packets.in.%s", mt), nil))
			PacketsOutByTypes.LoadOrStore(mt, metrics.GetOrRegisterMeter(fmt.Sprintf("neb.net.packets.out.%s", mt), nil))
			m, _ := dp.subscribersMap.LoadOrStore(mt, new(sync.Map))
			m.(*sync.Map).Store(v, true)
		}
	}
}

// Deregister deregister subscribers.
func (dp *Dispatcher) Deregister(subscribers ...*Subscriber) {

	for _, v := range subscribers {
		for _, mt := range v.msgTypes {
			m, _ := dp.subscribersMap.Load(mt)
			if m == nil {
				continue
			}
			m.(*sync.Map).Delete(v)
			dp.subscribersMap.Delete(mt)
		}
	}
}

// Start start message dispatch goroutine.
func (dp *Dispatcher) Start() {
	logging.CLog().Info("Launched Dispatcher.")

	go (func() {
		for {
			// logging.VLog().Info("dispatcher in loop")
			select {
			case <-dp.quitCh:
				logging.CLog().Info("Shutdowned Dispatcher.")
				return

			case msg := <-dp.receivedMessageCh:
				msgType := msg.MessageType()
				v, _ := dp.subscribersMap.Load(msgType)
				m, _ := v.(*sync.Map)
				m.Range(func(key, value interface{}) bool {
					key.(*Subscriber).msgChan <- msg
					return true
				})
			}
		}
	})()
}

// Stop stop goroutine.
func (dp *Dispatcher) Stop() {
	dp.quitCh <- true
}

// PutMessage put new message to chan, then subscribers will be notified to process.
func (dp *Dispatcher) PutMessage(msg Message) {
	dp.receivedMessageCh <- msg
}
