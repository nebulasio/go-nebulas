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

package core

import (
	"sync"

	log "github.com/sirupsen/logrus"
)

const (

	// TopicSendTransaction the topic of send a transaction.
	TopicSendTransaction = "chain.sendTransaction"

	// TopicDeploySmartContract the topic of deploy a smart contract.
	TopicDeploySmartContract = "chain.deploySmartContract"

	// TopicCallSmartContract the topic of call a smart contract.
	TopicCallSmartContract = "chain.callSmartContract"

	// TopicDelegate the topic of delegate.
	TopicDelegate = "chain.delegate"

	// TopicCandidate the topic of candidate.
	TopicCandidate = "chain.candidate"

	// TopicLinkBlock the topic of link a block.
	TopicLinkBlock = "chain.linkBlock"
)

// Event event structure.
type Event struct {
	Topic string
	Data  string
}

// EventEmitter provide event functionality for Nebulas.
type EventEmitter struct {
	eventSubs *sync.Map
	eventCh   chan *Event
	quitCh    chan int
}

// NewEventEmitter return new EventEmitter.
func NewEventEmitter() *EventEmitter {
	return &EventEmitter{
		eventSubs: new(sync.Map),
		eventCh:   make(chan *Event, 1024),
		quitCh:    make(chan int, 1),
	}
}

// Start start emitter.
func (emitter *EventEmitter) Start() {
	go emitter.loop()
}

// Stop stop emitter.
func (emitter *EventEmitter) Stop() {
	emitter.quitCh <- 1
}

// Trigger trigger event.
func (emitter *EventEmitter) Trigger(e *Event) {
	log.WithFields(log.Fields{
		"topic": e.Topic,
		"data":  e.Data,
	}).Debug("trigger new event")
	emitter.eventCh <- e
}

// Register register event chan.
func (emitter *EventEmitter) Register(topic string, ch chan *Event) error {

	v, ok := emitter.eventSubs.Load(topic)
	if !ok {
		v, _ = emitter.eventSubs.LoadOrStore(topic, new(sync.Map))
	}

	m, _ := v.(*sync.Map)
	m.Store(ch, true)

	return nil
}

// Deregister deregister event chan.
func (emitter *EventEmitter) Deregister(topic string, ch chan *Event) error {

	v, ok := emitter.eventSubs.Load(topic)
	if !ok {
		return nil
	}
	m, _ := v.(*sync.Map)
	m.Delete(ch)

	return nil
}

func (emitter *EventEmitter) loop() {
	for {
		select {
		case <-emitter.quitCh:
			log.Info("EventEmitter.loop: quit.")
			return
		case e := <-emitter.eventCh:

			topic := e.Topic
			v, ok := emitter.eventSubs.Load(topic)
			if !ok {
				continue
			}

			m, _ := v.(*sync.Map)
			m.Range(func(key, value interface{}) bool {
				key.(chan *Event) <- e
				return true
			})
		}
	}
}
