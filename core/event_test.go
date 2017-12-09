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
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func register(emitter *EventEmitter, category int, topic string) chan *Event {
	ch := make(chan *Event, 128)
	emitter.Register(category, topic, ch)
	return ch
}

func TestEventEmitter(t *testing.T) {
	// create emitter.
	emitter := NewEventEmitter()
	emitter.Start()

	// topic & categories
	NewEventCatgory := NodeEventCategory + 123
	categories := []int{ChainEventCategory, NodeEventCategory, NewEventCatgory}
	chainTopics := []string{"chain topic 01", "chain topic 02", "chain topic 03"}
	nodeTopics := []string{"node topic 11", "node topic 12"}

	// prepare chan.
	t1ch := register(emitter, ChainEventCategory, chainTopics[0])
	t2ch := register(emitter, ChainEventCategory, chainTopics[1])
	t3ch := register(emitter, NodeEventCategory, nodeTopics[0])

	wg := new(sync.WaitGroup)
	wg.Add(2)

	totalEventCount := 1000
	eventCountDist := make(map[string]int)
	genEventKey := func(category int, topic string) string {
		return fmt.Sprintf("%d-%s", category, topic)
	}

	go func() {
		// send message.
		defer wg.Done()
		rand.Seed(time.Now().UnixNano())

		for i := 0; i < totalEventCount; i++ {
			category := categories[rand.Intn(len(categories))]

			var topic string
			if category == ChainEventCategory {
				topic = chainTopics[rand.Intn(len(chainTopics))]
			} else if category == NodeEventCategory {
				topic = nodeTopics[rand.Intn(len(nodeTopics))]
			} else {
				topic = "New Category Topic"
			}

			key := genEventKey(category, topic)
			eventCountDist[key] = eventCountDist[key] + 1

			e := &Event{
				Topic:    topic,
				Category: category,
				Data:     fmt.Sprintf("%d", i),
			}
			emitter.Trigger(e)
		}
	}()

	t1c, t2c, t3c := 0, 0, 0
	go func() {
		defer wg.Done()

		for {
			select {
			case <-time.After(time.Second * 1):
				return
			case e := <-t1ch:
				assert.Equal(t, ChainEventCategory, e.Category)
				assert.Equal(t, chainTopics[0], e.Topic)
				t1c++
			case e := <-t2ch:
				assert.Equal(t, ChainEventCategory, e.Category)
				assert.Equal(t, chainTopics[1], e.Topic)
				t2c++
			case e := <-t3ch:
				assert.Equal(t, NodeEventCategory, e.Category)
				assert.Equal(t, nodeTopics[0], e.Topic)
				t3c++
			}
		}
	}()

	wg.Wait()

	at1c, at2c, at3c := eventCountDist[genEventKey(ChainEventCategory, chainTopics[0])], eventCountDist[genEventKey(ChainEventCategory, chainTopics[1])], eventCountDist[genEventKey(NodeEventCategory, nodeTopics[0])]
	log.Infof("actual vs. expected: %d vs. %d, %d vs. %d, %d vs. %d", at1c, t1c, at2c, t2c, at3c, t3c)
	assert.Equal(t, at1c, t1c)
	assert.Equal(t, at2c, t2c)
	assert.Equal(t, at3c, t3c)

	emitter.Stop()
	time.Sleep(time.Millisecond * 100)
}

func TestEventEmitterWithRunningRegDereg(t *testing.T) {
	// create emitter.
	emitter := NewEventEmitter()
	emitter.Start()

	// topic & categories
	categories := []int{ChainEventCategory, NodeEventCategory}
	chainTopics := []string{"chain topic 01", "chain topic 02"}
	nodeTopics := []string{"node topic 11"}

	wg := new(sync.WaitGroup)
	wg.Add(2)

	totalEventCount := 1000
	eventCountDist := make(map[string]int)
	genEventKey := func(category int, topic string) string {
		return fmt.Sprintf("%d-%s", category, topic)
	}

	// prepare chan.
	t1ch := register(emitter, ChainEventCategory, chainTopics[0])
	t2ch := register(emitter, ChainEventCategory, chainTopics[1])
	t3ch := register(emitter, NodeEventCategory, nodeTopics[0])

	go func() {
		// send message.
		defer wg.Done()
		rand.Seed(time.Now().UnixNano())

		for i := 0; i < totalEventCount; i++ {
			if i%100 == 99 {
				time.Sleep(time.Millisecond * 500)
			}

			category := categories[rand.Intn(len(categories))]

			var topic string
			if category == ChainEventCategory {
				topic = chainTopics[rand.Intn(len(chainTopics))]
			} else {
				topic = nodeTopics[rand.Intn(len(nodeTopics))]
			}

			key := genEventKey(category, topic)
			eventCountDist[key] = eventCountDist[key] + 1

			e := &Event{
				Topic:    topic,
				Category: category,
				Data:     fmt.Sprintf("%d", i),
			}
			emitter.Trigger(e)
		}
	}()

	t1c, t2c, t3c := 0, 0, 0
	go func() {
		defer wg.Done()

		for {
			select {
			case <-time.After(time.Second * 1):
				log.Infof("timeout")
				return
			case e := <-t1ch:
				assert.Equal(t, ChainEventCategory, e.Category)
				assert.Equal(t, chainTopics[0], e.Topic)
				t1c++

				if t1c%13 == 2 {
					emitter.Deregister(ChainEventCategory, chainTopics[1], t2ch)
				} else if t1c%13 == 9 {
					emitter.Register(ChainEventCategory, chainTopics[1], t2ch)
				}

			case e := <-t2ch:
				assert.Equal(t, ChainEventCategory, e.Category)
				assert.Equal(t, chainTopics[1], e.Topic)
				t2c++

				if t2c%13 == 4 {
					emitter.Deregister(NodeEventCategory, nodeTopics[0], t3ch)
				} else if t2c%13 == 12 {
					emitter.Register(NodeEventCategory, nodeTopics[0], t3ch)
				}

			case e := <-t3ch:
				assert.Equal(t, NodeEventCategory, e.Category)
				assert.Equal(t, nodeTopics[0], e.Topic)
				t3c++
			}
		}
	}()

	wg.Wait()

	at1c, at2c, at3c := eventCountDist[genEventKey(ChainEventCategory, chainTopics[0])], eventCountDist[genEventKey(ChainEventCategory, chainTopics[1])], eventCountDist[genEventKey(NodeEventCategory, nodeTopics[0])]

	log.Infof("actual vs. expected: %d vs. %d, %d vs. %d, %d vs. %d", at1c, t1c, at2c, t2c, at3c, t3c)

	emitter.Stop()
	time.Sleep(time.Millisecond * 100)
}

func TestEventEmitterWithUnsupportedCategory(t *testing.T) {
	// create emitter.
	emitter := NewEventEmitter()

	ch := make(chan *Event, 1)
	assert.Equal(t, ErrUnsupportedEventCategory, emitter.Register(ChainEventCategory+1234, "wow", ch))
	assert.Equal(t, ErrUnsupportedEventCategory, emitter.Deregister(ChainEventCategory+1234, "wow", ch))
}

func TestEventEmitterDeregister(t *testing.T) {
	// create emitter.
	emitter := NewEventEmitter()

	ch := make(chan *Event, 1)
	assert.Nil(t, emitter.Deregister(ChainEventCategory, "wow", ch))
}
