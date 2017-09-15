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

/*
Pow implementation of Proof-of-Work consensus, designed to be a state machine.
The following is the state diagram:

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
	nm    *net.Manager

	states            consensus.States
	currentState      consensus.State
	stateTransitionCh chan *stateTransitionArgs

	newBlock *core.Block
}

type stateTransitionArgs struct {
	nextState consensus.State
	data      interface{}
}

// NewPow create Pow instance.
func NewPow(bc *core.BlockChain, nm *net.Manager) *Pow {
	p := &Pow{
		chain:             bc,
		nm:                nm,
		quitCh:            make(chan bool, 5),
		stateTransitionCh: make(chan *stateTransitionArgs, 10),
	}

	p.states = consensus.States{
		Mining:  NewMiningState(p),
		Minted:  NewMintedState(p),
		Prepare: NewPrepareState(p),
		Stopped: NewStoppedState(p),
	}
	p.currentState = p.states[Prepare]

	nm.Register(p)

	return p
}

// Start start pow service.
func (p *Pow) Start() {
	// start state machine.
	go p.stateLoop()

	// start internal time loop for debug.
	go p.timeLoop()
}

// Stop stop pow service.
func (p *Pow) Stop() {
	// cleanup.
	p.quitCh <- true
	p.quitCh <- true
	p.Event(NewStopEvent())
}

// Event handle event.
func (p *Pow) Event(e consensus.Event) {
	nextState := p.currentState.Event(e)
	p.Transite(nextState, nil)
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

	p.stateTransitionCh <- &stateTransitionArgs{nextState: nextState, data: data}
}

// AppendBlock implement new block at tail algorithm.
func (p *Pow) AppendBlock(block *core.Block) error {
	bc := p.chain

	tailBlockHash := bc.TailBlock().Hash()
	blockParentHash := block.ParentHash()
	blockHash := block.Hash()

	logFields := log.Fields{
		"bc.latestBlock.header.hash": tailBlockHash,
		"block.header.parentHash":    blockParentHash,
		"block.header.hash":          blockHash,
	}

	if tailBlockHash == blockParentHash {
		log.WithFields(logFields).Info("New block")
		bc.SetTailBlock(block)

	} else {
		log.WithFields(logFields).Info("New forked block")

		// // find the root block in detached blocks.
		// rootParentBlock := block
		// for {
		// 	i, _ := bc.detachedBlocks.Get(rootParentBlock.header.parentHash)
		// 	if ib, ok := i.(*Block); ok {
		// 		ib.nextBlock = rootParentBlock
		// 		rootParentBlock.previousBlock = ib
		// 		rootParentBlock = ib
		// 	} else {
		// 		break
		// 	}
		// }

		// // recursively find the common ancestor.
		// ancestor := bc.latestBlock
		// for ; ancestor != nil && ancestor.header.hash != rootParentBlock.header.hash; ancestor = ancestor.previousBlock {
		// 	bc.detachedBlocks.Add(ancestor.header.hash, ancestor)
		// }

		// if ancestor == nil {
		// 	log.WithFields(logFields).Error("No common ancestor")
		// 	return bc, errors.New("No common ancestor")
		// }

		// // alter the chain.
		// ancestor.nextBlock = rootParentBlock
		// rootParentBlock.previousBlock = ancestor
		// bc.latestBlock = block
	}

	return nil
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

func (p *Pow) stateLoop() {
	p.currentState.Enter(nil)

	for {
		select {
		case args := <-p.stateTransitionCh:
			nextState := args.nextState
			data := args.data

			if p.currentState == nextState {
				continue
			}

			p.currentState.Leave(data)
			p.currentState = nextState
			p.currentState.Enter(data)

		case <-p.quitCh:
			log.Info("quit Pow.loop.")
			return
		}
	}
}
