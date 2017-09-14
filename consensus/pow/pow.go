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

package pow

import (
	"time"

	"github.com/nebulasio/go-nebulas/consensus"
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/net"
	"github.com/nebulasio/go-nebulas/net/messages"
	log "github.com/sirupsen/logrus"
)

// Pow implementation of Proof-of-Work consensus.
// Pow is designed to be a state machine.
// The following is the state diagram:
/*
@startuml
[*] --> Prepare
Prepare --> Mining : start mining
Mining --> Prepare : new block received
Mining --> Minted : found the nonce/block
Minted --> Prepare : broadcast the block, and start over
Prepare --> [*] : stop
Mining --> [*] : stop
Minted --> [*] : stop
@enduml
*/
type Pow struct {
	quitCh chan bool

	chain *core.BlockChain
	nm    *net.NetManager

	states       consensus.States
	currentState consensus.State

	newBlock *core.Block
}

// NewPow create Pow instance.
func NewPow(bc *core.BlockChain, nm *net.NetManager) *Pow {
	p := &Pow{chain: bc, nm: nm, quitCh: make(chan bool, 5)}
	p.states = consensus.States{
		Mining:  NewMiningState(p),
		Minted:  NewMintedState(p),
		Prepare: NewPrepareState(p),
		Stopped: NewStoppedState(p),
	}

	nm.Register(p)

	return p
}

// Start start pow service.
func (p *Pow) Start() *Pow {
	// start state machine.
	go p.loop()

	// start internal time loop for debug.
	go p.timeLoop()

	return p
}

// Stop stop pow service.
func (p *Pow) Stop() *Pow {
	// cleanup.
	p.quitCh <- true
	p.quitCh <- true
	p.Event(NewStopEvent())
	return p
}

// Event handle event.
func (p *Pow) Event(e consensus.Event) *Pow {
	nextState := p.currentState.Event(e)
	p.Transite(nextState, nil)
	return p
}

// TransiteByKey transite state by stateKey.
func (p *Pow) TransiteByKey(stateKey string, data interface{}) {
	p.Transite(p.states[stateKey], data)
}

// Transite transite state.
func (p *Pow) Transite(nextState consensus.State, data interface{}) {
	if p.currentState == nextState {
		return
	}

	p.currentState.Leave(data)
	p.currentState = nextState
	p.currentState.Enter(data)
}

// SubscribeMessageTypes return all message types wanting to subscribe in network.
// Subscribe the following message types:
// @NewBlockMessageType
func (p *Pow) SubscribeMessageTypes() []net.MessageType {
	list := make([]net.MessageType, 1)
	list = append(list, messages.NewBlockMessageType)
	return list
}

// OnMessageReceived handle new received network message.
func (p *Pow) OnMessageReceived(msg net.Message) {
	log.WithFields(log.Fields{
		"msg": msg,
	}).Info("Pow received message.")
}

func (p *Pow) timeLoop() {
	ticker := time.NewTicker(time.Second * 5)
	for {
		select {
		case <-ticker.C:
			p.Event(NewTimeoutEvent(time.Now()))
			continue
		case <-p.quitCh:
			log.Debug("quit Pow.timeLoop.")
			return
		}
	}
}

func (p *Pow) loop() {
	p.currentState = p.states[Prepare]
	p.currentState.Enter(nil)

	for {
		select {
		case <-p.quitCh:
			log.Info("quit Pow.loop.")
			return
		}
	}
}
