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

import (
	log "github.com/sirupsen/logrus"
)

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

// StateMachine structure
type StateMachine struct {
	quitCh            chan bool
	currentState      State
	stateTransitionCh chan *StateTransitionArgs
	context           interface{}
}

// NewStateMachine create a new state machine
func NewStateMachine(context interface{}) *StateMachine {
	return &StateMachine{
		quitCh:            make(chan bool),
		stateTransitionCh: make(chan *StateTransitionArgs, 10),
		context:           context,
	}
}

// SetInitialState set the first state
func (sm *StateMachine) SetInitialState(state State) {
	sm.currentState = state
}

// Start pow service.
func (sm *StateMachine) Start() {
	go sm.stateLoop()
}

// Stop pow service.
func (sm *StateMachine) Stop() {
	// cleanup.
	sm.quitCh <- true
}

// Context return state machine context
func (sm *StateMachine) Context() interface{} {
	return sm.context
}

// StateTransitionArgs contains transition data
type StateTransitionArgs struct {
	from State
	to   State
	data interface{}
}

/*
Event handle events from Network or State.
The whole event process should be as the following:
1. dispatch to currentState to process.
2. if currentState does not captured it, consensus process it by default.
*/
func (sm *StateMachine) Event(e Event) {
	captured, nextState := sm.currentState.Event(e)
	if captured {
		sm.Transit(sm.currentState, nextState, nil)
		return
	}

	// default procedure.
	log.WithFields(log.Fields{
		"eventType":    e.EventType(),
		"currentState": sm.currentState,
	}).Info("StateMachine.Event: not captured event.")
}

// Transit transit state.
func (sm *StateMachine) Transit(from, to State, data interface{}) {
	sm.stateTransitionCh <- &StateTransitionArgs{from: from, to: to, data: data}
}

func (sm *StateMachine) checkValidTransit(from, to State) bool {
	valid := from != nil && to != nil && from != to && sm.currentState == from
	log.WithFields(log.Fields{
		"func":    "StateMachine.CheckTransit",
		"success": valid,
		"current": sm.currentState,
		"from":    from,
		"to":      to,
	}).Debug("State Transition.")
	return valid
}

func (sm *StateMachine) stateLoop() {
	sm.currentState.Enter(nil)

	for {
		select {
		case args := <-sm.stateTransitionCh:
			to := args.to
			data := args.data
			from := args.from

			if !sm.checkValidTransit(from, to) {
				continue
			}

			sm.currentState.Leave(data)
			sm.currentState = to
			sm.currentState.Enter(data)

		case <-sm.quitCh:
			log.Info("quit Pow.loop.")
			return
		}
	}
}
