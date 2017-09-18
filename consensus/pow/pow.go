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
	"bytes"
	"fmt"
	"time"

	"github.com/nebulasio/go-nebulas/components/net"
	"github.com/nebulasio/go-nebulas/components/net/messages"
	"github.com/nebulasio/go-nebulas/consensus"
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/utils/byteutils"

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
	messageReceivedCh chan net.Message

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
		messageReceivedCh: make(chan net.Message, 128),
	}

	p.states = consensus.States{
		Mining:  NewMiningState(p),
		Minted:  NewMintedState(p),
		Prepare: NewPrepareState(p),
		Stopped: NewStoppedState(p),
	}
	p.currentState = p.states[Prepare]

	nm.Register(net.NewSubscriber(p, p.messageReceivedCh, messages.NewBlockMessageType))

	return p
}

// Start start pow service.
func (p *Pow) Start() {
	// start state machine.
	go p.stateLoop()

	// start goroutine to process received message.
	go p.messageLoop()
}

// Stop stop pow service.
func (p *Pow) Stop() {
	// cleanup.
	p.quitCh <- true
	p.quitCh <- true
}

/*
Event handle events from Network or State.
The whole event process should be as the following:
1. dispatch to currentState to process.
2. if currentState does not captured it, consensus process it by default.
*/
func (p *Pow) Event(e consensus.Event) {
	captured, nextState := p.currentState.Event(e)
	if captured {
		if nextState != nil && p.currentState != nextState {
			p.Transite(nextState, nil)
		}
		return
	}

	// default procedure.
	switch t := e.(type) {
	case *NetMessageEvent:
		switch msg := t.message.(type) {
		case *messages.BlockMessage:
			log.WithFields(log.Fields{
				"block": msg.Block(),
			}).Info("Pow handle BlockMessage.")
		default:
			log.WithFields(log.Fields{
				"messageType": t.message.MessageType(),
				"message":     t.message,
			}).Info("Pow handle NetMessageEvent.")
		}
	default:
		log.WithFields(log.Fields{
			"eventType": fmt.Sprintf("%T", e),
		}).Info("Pow handle this event.")
	}
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
		"bc.latestBlock.header.hash": byteutils.Hex(tailBlockHash),
		"block.header.parentHash":    byteutils.Hex(blockParentHash),
		"block.header.hash":          byteutils.Hex(blockHash),
	}

	if bytes.Compare(tailBlockHash, blockParentHash) == 0 {
		log.WithFields(logFields).Info("New block")
		bc.SetTailBlock(block)

	} else {
		log.WithFields(logFields).Info("New forked block")
		//TODO: implement fork choice algorithm.

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

// TODO: Timeout Event seems useless, consider to remove them.
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

func (p *Pow) messageLoop() {
	for {
		select {
		case msg := <-p.messageReceivedCh:
			p.Event(NewNetMessageEvent(msg))
		case <-p.quitCh:
			// TODO: should provide base goroutine start/stop func to graceful stop them.
			/*
				for example,

				type Stopper struct {
					quitCh chan int // maybe int is better than bool, less confuss.
					count int q		// should use thread-safe int, eg. AtomicInt.
				}
				func NewStopper() *Stopper {
					s := &Stopper{quitCh: make(chan int，16), count : 0}
					return s
				}
				func (s *Stopper) CountMe() {
					s.count++
				}
				func (s *Stopper) QuitMe() {
					for i :=0 ; i<s.count; i++ {
						s.quitCh <- 0
					}
				}
			*/
			log.Info("quit Pow.messageLoop.")
			return
		}
	}
}
