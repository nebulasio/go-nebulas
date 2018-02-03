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

	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

const (

	// TopicPendingTransaction the topic of pending a transaction in transaction_pool.
	TopicPendingTransaction = "chain.pendingTransaction"

	// TopicSendTransaction the topic of send a transaction.
	TopicSendTransaction = "chain.sendTransaction"

	// TopicDeploySmartContract the topic of deploy a smart contract.
	TopicDeploySmartContract = "chain.deployContract"

	// TopicCallSmartContract the topic of call a smart contract.
	TopicCallSmartContract = "chain.callContract"

	// TopicDelegate the topic of delegate.
	TopicDelegate = "chain.delegate"

	// TopicCandidate the topic of candidate.
	TopicCandidate = "chain.candidate"

	// TopicLinkBlock the topic of link a block.
	TopicLinkBlock = "chain.linkBlock"

	// TopicLibBlock the topic of latest irreversible block.
	TopicLibBlock = "chain.latestIrreversibleBlock"

	// TopicExecuteTxFailed the topic of execute a transaction failed.
	TopicExecuteTxFailed = "chain.executeTxFailed"

	// TopicExecuteTxSuccess the topic of execute a transaction success.
	TopicExecuteTxSuccess = "chain.executeTxSuccess"
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
	size      int
}

// NewEventEmitter return new EventEmitter.
func NewEventEmitter(size int) *EventEmitter {
	return &EventEmitter{
		eventSubs: new(sync.Map),
		eventCh:   make(chan *Event, size),
		quitCh:    make(chan int, 1),
		size:      size,
	}
}

// Start start emitter.
func (emitter *EventEmitter) Start() {
	logging.CLog().WithFields(logrus.Fields{
		"size": emitter.size,
	}).Info("Starting EventEmitter...")

	go emitter.loop()
}

// Stop stop emitter.
func (emitter *EventEmitter) Stop() {
	logging.CLog().WithFields(logrus.Fields{
		"size": emitter.size,
	}).Info("Stopping EventEmitter...")

	emitter.quitCh <- 1
}

// Trigger trigger event.
func (emitter *EventEmitter) Trigger(e *Event) {
	/* 	logging.VLog().WithFields(logrus.Fields{
		"topic": e.Topic,
		"data":  e.Data,
	}).Debug("Trigger new event") */
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
	logging.CLog().Info("Started EventEmitter.")

	for {
		select {
		case <-emitter.quitCh:
			logging.CLog().Info("Stopped EventEmitter.")
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
