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
	log "github.com/sirupsen/logrus"
)

type messageSubscriberMap map[MessageType]map[*Subscriber]bool

// Dispatcher a message dispatcher service.
type Dispatcher struct {
	subscribersMap    messageSubscriberMap
	quitCh            chan bool
	receivedMessageCh chan Message
}

// NewDispatcher create Dispatcher instance.
func NewDispatcher() *Dispatcher {
	dp := &Dispatcher{
		subscribersMap:    make(messageSubscriberMap),
		quitCh:            make(chan bool, 10),
		receivedMessageCh: make(chan Message, 1024),
	}

	return dp
}

// Register register subscribers.
func (dp *Dispatcher) Register(subscribers ...*Subscriber) {
	for _, v := range subscribers {
		for _, mt := range v.msgTypes {
			m := dp.subscribersMap[mt]
			if m == nil {
				m = make(map[*Subscriber]bool)
				dp.subscribersMap[mt] = m
			}
			m[v] = true
		}
	}
}

// Deregister deregister subscribers.
func (dp *Dispatcher) Deregister(subscribers ...*Subscriber) {
	for _, v := range subscribers {
		for _, mt := range v.msgTypes {
			m := dp.subscribersMap[mt]
			if m == nil {
				continue
			}
			delete(m, v)
			if len(m) == 0 {
				delete(dp.subscribersMap, mt)
			}
		}
	}
}

// Start start message dispatch goroutine.
func (dp *Dispatcher) Start() {
	go (func() {
		count := 0
		for {
			// log.Info("dispatcher in loop")
			select {
			case <-dp.quitCh:
				log.Info("dispatcher.loop: dispatcher is stopped.")
				return

			case msg := <-dp.receivedMessageCh:
				count++
				log.Debugf("dispatcher.loop: recvMsgCount=%d", count)
				msgType := msg.MessageType()
				msgListener := dp.subscribersMap[msgType]

				if msgListener == nil || len(msgListener) == 0 {
					log.WithFields(log.Fields{
						"MessageType": msgType,
					}).Info("dispatcher.loop: receive message without handler.")
					continue
				}

				// send message to each subscriber.
				for listener := range msgListener {
					listener.msgChan <- msg
				}
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
