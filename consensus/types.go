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

const (
	NetMessageEvent = "event.netmessage"
)

// Consensus interface of consensus algorithm.
type Consensus interface {
	Start()
	Stop()
	TransiteByKey(nextStateKey string, data interface{})
	Transite(nextState State, data interface{})

	// AppendBlock add block to blockchain according to for choice algorithm.
	AppendBlock(block *core.Block) error
}

type EventType string

type Event interface {
	EventType() EventType
	Data() interface{}
}

type State interface {
	Event(e Event) (bool, State)
	Enter(data interface{})
	Leave(data interface{})
}

type States map[string]State

type BaseEvent struct {
	eventType EventType
	data      interface{}
}

func NewBaseEvent(t EventType, data interface{}) Event {
	return &BaseEvent{eventType: t, data: data}
}

func (e *BaseEvent) EventType() EventType {
	return e.eventType
}

func (e *BaseEvent) Data() interface{} {
	return e.data
}
