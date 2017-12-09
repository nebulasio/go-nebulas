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
	categories := []int{ChainEventCategory, NodeEventCategory}
	chainTopics := []string{"chain topic 01", "chain topic 02"}
	nodeTopics := []string{"node topic 11"}

	// prepare chan.
	t1ch := register(emitter, ChainEventCategory, chainTopics[0])
	t2ch := register(emitter, ChainEventCategory, chainTopics[1])
	t3ch := register(emitter, NodeEventCategory, nodeTopics[0])

	wg := new(sync.WaitGroup)
	wg.Add(2)

	eventCount := 500
	expectedEvents := make([]string, 0)

	go func() {
		// send message.
		defer wg.Done()

		for i := 0; i < eventCount; i++ {
			category := categories[rand.Intn(len(categories))]
			var topic string
			if category == ChainEventCategory {
				topic = chainTopics[rand.Intn(len(chainTopics))]
			} else {
				topic = nodeTopics[rand.Intn(len(nodeTopics))]
			}

			e := &Event{
				Topic:    topic,
				Category: category,
				Data:     fmt.Sprintf("%d", i),
			}
			emitter.Trigger(e)
		}
	}()

	go func() {
		defer wg.Done()

		for {
			select {
			case <-time.After(time.Second * 1):
				return
			case e := <-t1ch:
				assert.Equal(t, ChainEventCategory, e.Category)
				assert.Equal(t, chainTopics[0], e.Topic)
				expectedEvents = append(expectedEvents, e.Data)
			case e := <-t2ch:
				assert.Equal(t, ChainEventCategory, e.Category)
				assert.Equal(t, chainTopics[1], e.Topic)
				expectedEvents = append(expectedEvents, e.Data)
			case e := <-t3ch:
				assert.Equal(t, NodeEventCategory, e.Category)
				assert.Equal(t, nodeTopics[0], e.Topic)
				expectedEvents = append(expectedEvents, e.Data)
			}
		}
	}()

	wg.Wait()
	assert.Equal(t, eventCount, len(expectedEvents))
}
