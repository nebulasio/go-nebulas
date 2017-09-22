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

package consensus

import "github.com/nebulasio/go-nebulas/core"

// EventType list
const (
	NetMessageEvent = "event.netmessage"
	NewBlockEvent   = "event.newblock"
)

// Consensus interface of consensus algorithm.
type Consensus interface {
	Start()
	Stop()

	TransitByKey(nextStateKey string, data interface{})
	Transit(nextState State, data interface{})

	VerifyBlock(*core.Block) error
}

// EventType of Events in Consensus State-Machine
type EventType string

// Event in Consensus State-Machine
type Event interface {
	EventType() EventType
	Data() interface{}
}

// State in Consensus State-Machine
type State interface {
	Event(e Event) (bool, State)
	Enter(data interface{})
	Leave(data interface{})
}

// States contains all possiable states in Consensus State-Machine
type States map[string]State

// BaseEvent is a kind of event structure
type BaseEvent struct {
	eventType EventType
	data      interface{}
}

// NewBaseEvent creates an event
func NewBaseEvent(t EventType, data interface{}) Event {
	return &BaseEvent{eventType: t, data: data}
}

// EventType of an event instance
func (e *BaseEvent) EventType() EventType {
	return e.eventType
}

// Data of an event instance
func (e *BaseEvent) Data() interface{} {
	return e.data
}
