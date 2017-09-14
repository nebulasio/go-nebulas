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

type Dispatcher struct {
	messageHandlerMap MessageHandlerMap

	quitCh            chan bool
	receivedMessageCh chan Message
}

type MessageHandlerMap map[MessageType]MessageHandlers

func NewDispatcher() *Dispatcher {
	dp := &Dispatcher{
		messageHandlerMap: make(MessageHandlerMap),

		quitCh:            make(chan bool, 10),
		receivedMessageCh: make(chan Message, 1024),
	}

	return dp
}

func (dp *Dispatcher) Register(handler MessageHandler) {
	for _, v := range handler.SubscribeMessageTypes() {
		msgHandlers := dp.messageHandlerMap[v]
		if msgHandlers == nil {
			msgHandlers = make(MessageHandlers)
			dp.messageHandlerMap[v] = msgHandlers
		}
		msgHandlers[handler] = true
	}
}

func (dp *Dispatcher) Deregister(handler MessageHandler) {
	for _, v := range handler.SubscribeMessageTypes() {
		msgHandlers := dp.messageHandlerMap[v]
		if msgHandlers == nil {
			continue
		}
		delete(msgHandlers, handler)
		if len(msgHandlers) == 0 {
			delete(dp.messageHandlerMap, v)
		}
	}
}

func (dp *Dispatcher) Start() {
	go (func() {
		for {
			select {
			case <-dp.quitCh:
				log.Info("dispatcher is stopped.")
				return
			case msg := <-dp.receivedMessageCh:
				msgID := msg.MessageType()
				msgHandlers := dp.messageHandlerMap[msgID]
				if msgHandlers == nil || len(msgHandlers) == 0 {
					log.WithFields(log.Fields{
						"MessageType": msgID,
					}).Info("receive message without handler.")
					break
				}

				for handler := range msgHandlers {
					handler.OnMessageReceived(msg)
				}
			}
		}
	})()
}

func (dp *Dispatcher) Stop() {
	dp.quitCh <- true
}

func (dp *Dispatcher) OnMessageReceived(msg Message) {
	dp.receivedMessageCh <- msg
}
