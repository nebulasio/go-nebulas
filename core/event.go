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

	"time"

	"github.com/nebulasio/go-nebulas/core/state"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

const (
	// TopicPendingTransaction the topic of pending a transaction in transaction_pool.
	TopicPendingTransaction = "chain.pendingTransaction"

	// TopicLibBlock the topic of latest irreversible block.
	TopicLibBlock = "chain.latestIrreversibleBlock"

	// TopicTransactionExecutionResult the topic of transaction execution result
	TopicTransactionExecutionResult = "chain.transactionResult"

	// TopicNewTailBlock the topic of new tail block set
	TopicNewTailBlock = "chain.newTailBlock"

	// TopicRevertBlock the topic of revert block
	TopicRevertBlock = "chain.revertBlock"

	// TopicDropTransaction drop tx (1): smaller nonce (2) expire txLifeTime
	TopicDropTransaction = "chain.dropTransaction"

	// TopicTransferFromContract transfer from contract
	TopicTransferFromContract = "chain.transferFromContract"

	// TopicInnerTransferContract inner transfer
	TopicInnerContract = "chain.innerContract"

	// TopicPodStateUpdate update pod state
	TopicPodStateUpdate = "chain.podStateUpdate"
)

// EventSubscriber subscriber object
type EventSubscriber struct {
	eventCh chan *state.Event
	topics  []string
}

// NewEventSubscriber returns an EventSubscriber
func NewEventSubscriber(size int, topics []string) *EventSubscriber {
	eventCh := make(chan *state.Event, size)
	subscriber := &EventSubscriber{
		eventCh: eventCh,
		topics:  topics,
	}
	return subscriber
}

// EventChan returns subscriber's eventCh
func (s *EventSubscriber) EventChan() chan *state.Event {
	return s.eventCh
}

// EventEmitter provide event functionality for Nebulas.
type EventEmitter struct {
	eventSubs *sync.Map
	eventCh   chan *state.Event
	quitCh    chan int
	size      int
}

// NewEventEmitter return new EventEmitter.
func NewEventEmitter(size int) *EventEmitter {
	return &EventEmitter{
		eventSubs: new(sync.Map),
		eventCh:   make(chan *state.Event, size),
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
func (emitter *EventEmitter) Trigger(e *state.Event) {
	emitter.eventCh <- e
}

// Register register event chan.
func (emitter *EventEmitter) Register(subscribers ...*EventSubscriber) {

	for _, v := range subscribers {
		for _, topic := range v.topics {
			m, _ := emitter.eventSubs.LoadOrStore(topic, new(sync.Map))
			m.(*sync.Map).Store(v, true)
		}
	}
}

// Deregister deregister event chan.
func (emitter *EventEmitter) Deregister(subscribers ...*EventSubscriber) {
	for _, v := range subscribers {
		for _, topic := range v.topics {
			m, _ := emitter.eventSubs.Load(topic)
			if m == nil {
				continue
			}
			m.(*sync.Map).Delete(v)
		}
	}
}

func (emitter *EventEmitter) loop() {
	logging.CLog().Info("Started EventEmitter.")

	timerChan := time.NewTicker(time.Second).C
	for {
		select {
		case <-timerChan:
			metricsCachedEvent.Update(int64(len(emitter.eventCh)))
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
				select {
				case key.(*EventSubscriber).eventCh <- e:
				default:
					logging.VLog().WithFields(logrus.Fields{
						"topic": topic,
					}).Warn("timeout to dispatch event.")
				}
				return true
			})
		}
	}
}
