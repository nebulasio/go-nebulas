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

	"github.com/stretchr/testify/assert"
)

func register(emitter *EventEmitter, topic string) *EventSubscriber {
	eventSub := NewEventSubscriber(128, []string{topic})
	emitter.Register(eventSub)
	return eventSub
}

func TestEventEmitter(t *testing.T) {
	// create emitter.
	emitter := NewEventEmitter(1024)
	emitter.Start()

	// topic & categories
	topics := []string{"chain.topic.01", "chain.topic.02", "chain.topic.03", "node.topic.11", "node.topic.12"}

	// prepare chan.
	t1ch := register(emitter, topics[0])
	t2ch := register(emitter, topics[1])
	t3ch := register(emitter, topics[2])

	wg := new(sync.WaitGroup)
	wg.Add(2)

	totalEventCount := 1000
	eventCountDist := make(map[string]int)
	go func() {
		// send message.
		defer wg.Done()
		rand.Seed(time.Now().UnixNano())

		for i := 0; i < totalEventCount; i++ {

			topic := topics[rand.Intn(len(topics))]

			eventCountDist[topic] = eventCountDist[topic] + 1

			e := &Event{
				Topic: topic,
				Data:  fmt.Sprintf("%d", i),
			}
			emitter.Trigger(e)
		}
	}()

	t1c, t2c, t3c := 0, 0, 0
	for len(t1ch.eventCh) > 0 {
		e := <-t1ch.eventCh
		assert.Equal(t, topics[0], e.Topic)
		t1c++
	}
	for len(t2ch.eventCh) > 0 {
		e := <-t2ch.eventCh
		assert.Equal(t, topics[1], e.Topic)
		t2c++
	}
	for len(t3ch.eventCh) > 0 {
		e := <-t3ch.eventCh
		assert.Equal(t, topics[2], e.Topic)
		t3c++
	}

	at1c, at2c, at3c := eventCountDist[topics[0]], eventCountDist[topics[1]], eventCountDist[topics[2]]
	assert.Equal(t, at1c, t1c)
	assert.Equal(t, at2c, t2c)
	assert.Equal(t, at3c, t3c)

	emitter.Stop()
	time.Sleep(time.Millisecond * 100)
}

func TestEventEmitterWithRunningRegDereg(t *testing.T) {
	// create emitter.
	emitter := NewEventEmitter(1024)
	emitter.Start()

	// topic
	topics := []string{"chain.topic.01", "chain.topic.02", "node.topic.11"}

	wg := new(sync.WaitGroup)
	wg.Add(2)

	totalEventCount := 1000
	eventCountDist := make(map[string]int)

	// prepare chan.
	t1ch := register(emitter, topics[0])
	t2ch := register(emitter, topics[1])
	t3ch := register(emitter, topics[2])

	go func() {
		// send message.
		defer wg.Done()
		rand.Seed(time.Now().UnixNano())

		for i := 0; i < totalEventCount; i++ {
			if i%100 == 99 {
				time.Sleep(time.Millisecond * 500)
			}

			topic := topics[rand.Intn(len(topics))]

			eventCountDist[topic] = eventCountDist[topic] + 1

			e := &Event{
				Topic: topic,
				Data:  fmt.Sprintf("%d", i),
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
			case e := <-t1ch.eventCh:
				assert.Equal(t, topics[0], e.Topic)
				t1c++

				if t1c%13 == 2 {
					emitter.Deregister(t2ch)
				} else if t1c%13 == 9 {
					emitter.Register(t2ch)
				}

			case e := <-t2ch.eventCh:
				assert.Equal(t, topics[1], e.Topic)
				t2c++

				if t2c%13 == 4 {
					emitter.Deregister(t3ch)
				} else if t2c%13 == 12 {
					emitter.Register(t3ch)
				}

			case e := <-t3ch.eventCh:
				assert.Equal(t, topics[2], e.Topic)
				t3c++
			}
		}
	}()

	wg.Wait()

	// TODO(Leon): check result

	emitter.Stop()
	time.Sleep(time.Millisecond * 100)
}
