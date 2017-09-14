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

	"github.com/nebulasio/go-nebulas/blockchain"
	"github.com/nebulasio/go-nebulas/consensus"
	"github.com/nebulasio/go-nebulas/net"
	"github.com/nebulasio/go-nebulas/net/messages"
	log "github.com/sirupsen/logrus"
)

type Pow struct {
	quitCh chan bool

	chain *blockchain.BlockChain
	nm    *net.NetManager

	states       consensus.States
	currentState consensus.State

	newBlock *blockchain.Block
}

func NewPow(bc *blockchain.BlockChain, nm *net.NetManager) *Pow {
	p := &Pow{chain: bc, nm: nm, quitCh: make(chan bool)}
	p.states = consensus.States{
		Mining:  NewMiningState(p),
		Minted:  NewMintedState(p),
		Prepare: NewPrepareState(p),
		Stopped: NewStoppedState(p),
	}

	nm.Register(p)

	return p
}

func (p *Pow) Start() *Pow {
	// start state machine.
	go p.loop()

	// start internal time loop for debug.
	go p.timeLoop()

	return p
}

func (p *Pow) Stop() *Pow {
	// cleanup.
	p.quitCh <- true
	p.quitCh <- true
	p.Event(NewStopEvent())
	return p
}

func (p *Pow) Event(e consensus.Event) *Pow {
	nextState := p.currentState.Event(e)
	p.Transite(nextState, nil)
	return p
}

func (p *Pow) TransiteByKey(stateKey string, data interface{}) {
	p.Transite(p.states[stateKey], data)
}

func (p *Pow) Transite(nextState consensus.State, data interface{}) {
	if p.currentState == nextState {
		return
	}

	p.currentState.Leave(data)
	p.currentState = nextState
	p.currentState.Enter(data)
}

func (p *Pow) SubscribeMessageTypes() []net.MessageType {
	list := make([]net.MessageType, 1)
	list = append(list, messages.NewBlockMessageType)
	return list
}

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
