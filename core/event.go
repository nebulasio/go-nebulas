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
	"errors"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	// ChainEventCategory events coming from chain, for example, Smart Contract Event, Tx Event, etc.
	ChainEventCategory = iota

	// NodeEventCategory events coming from node, for example, New Transaction Submitted, New Block Received, After Fork Choice, egc.
	NodeEventCategory
)

var (
	// ErrUnsupportedEventCategory unsupported event category.
	ErrUnsupportedEventCategory = errors.New("unsupported event category")
)

// Event event structure.
type Event struct {
	Category int
	Topic    string
	Data     string
	created  time.Time
}

// EventEmitter provide event functionality for Nebulas.
type EventEmitter struct {
	chainEventSubs *sync.Map
	nodeEventSubs  *sync.Map
	eventCh        chan *Event
	quitCh         chan int
}

// NewEventEmitter return new EventEmitter.
func NewEventEmitter() *EventEmitter {
	return &EventEmitter{
		chainEventSubs: new(sync.Map),
		nodeEventSubs:  new(sync.Map),
		eventCh:        make(chan *Event, 1024),
		quitCh:         make(chan int, 1),
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
		"category": e.Category,
		"topic":    e.Topic,
		"data":     e.Data,
	}).Debug("trigger new event")
	emitter.eventCh <- e
}

// Register register event chan.
func (emitter *EventEmitter) Register(category int, topic string, ch chan *Event) error {
	subs, err := emitter.getEventSubscriptors(category)
	if err != nil {
		return err
	}

	// get subs of chan by topic.
	v, ok := subs.Load(topic)
	if !ok {
		v, _ = subs.LoadOrStore(topic, new(sync.Map))
	}

	m, _ := v.(*sync.Map)
	m.Store(ch, true)

	return nil
}

// Deregister deregister event chan.
func (emitter *EventEmitter) Deregister(category int, topic string, ch chan *Event) error {
	subs, err := emitter.getEventSubscriptors(category)
	if err != nil {
		return err
	}

	// get subs of chan by topic.
	v, ok := subs.Load(topic)
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
			category := e.Category
			topic := e.Topic

			subs, err := emitter.getEventSubscriptors(category)
			if err != nil {
				log.WithFields(log.Fields{
					"err":      err,
					"category": category,
					"topic":    topic,
				}).Warnf("the category is unsupported.")
				continue
			}

			v, ok := subs.Load(topic)
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

func (emitter *EventEmitter) getEventSubscriptors(category int) (*sync.Map, error) {
	var subs *sync.Map
	switch category {
	case ChainEventCategory:
		subs = emitter.chainEventSubs
	case NodeEventCategory:
		subs = emitter.nodeEventSubs
	default:
		return nil, ErrUnsupportedEventCategory
	}

	return subs, nil
}
