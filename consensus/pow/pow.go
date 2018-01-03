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
	"errors"

	"github.com/nebulasio/go-nebulas/consensus"
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/neblet/pb"
	"github.com/nebulasio/go-nebulas/net/p2p"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

// Errors in PoW Consensus
var (
	ErrInvalidDataType   = errors.New("invalid data type, should be *core.Block")
	ErrInvalidBlockNonce = errors.New("invalid block nonce")
)

// Neblet interface breaks cycle import dependency and hides unused services.
type Neblet interface {
	Config() nebletpb.Config
	BlockChain() *core.BlockChain
	NetManager() p2p.Manager
}

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

	chain    *core.BlockChain
	nm       p2p.Manager
	coinbase *core.Address

	currentState      consensus.State
	stateTransitionCh chan *stateTransitionArgs

	miningBlock      *core.Block
	newBlockReceived bool

	canMining bool
}

type stateTransitionArgs struct {
	from consensus.State
	to   consensus.State
	data interface{}
}

// NewPow create Pow instance.
func NewPow(neblet Neblet) *Pow {
	p := &Pow{
		chain:             neblet.BlockChain(),
		nm:                neblet.NetManager(),
		quitCh:            make(chan bool, 5),
		stateTransitionCh: make(chan *stateTransitionArgs, 10),
		canMining:         false,
	}

	coinbaseConf := neblet.Config().Chain.Coinbase
	if coinbaseConf != "" {
		coinbase, err := core.AddressParse(coinbaseConf)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err": err,
			}).Info("Pow.NewPow: coinbase parse err.")
			//panic("coinbase should be configed for pow")
		}
		p.coinbase = coinbase
	} else {
		//panic("coinbase should be configed for pow")
	}

	p.currentState = NewStartState(p)

	return p
}

// Start start pow service.
func (p *Pow) Start() {
	// start state machine.
	go p.stateLoop()

	// start goroutine to process received message.
	go p.blockLoop()
}

// Stop stop pow service.
func (p *Pow) Stop() {
	// cleanup.
	p.quitCh <- true
	p.quitCh <- true
}

// CanMining return if consensus can do mining now
func (p *Pow) CanMining() bool {
	return p.canMining
}

// SetCanMining set if consensus can do mining now
func (p *Pow) SetCanMining(canMining bool) {
	logging.VLog().Info("sync over, start mining")
	p.canMining = canMining
	p.Event(consensus.NewBaseEvent(consensus.CanMiningEvent, nil))
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
		p.Transit(p.currentState, nextState, nil)
		return
	}

	// default procedure.
	et := e.EventType()
	switch et {
	case consensus.NewBlockEvent:
		block := e.Data().(*core.Block)
		logging.VLog().WithFields(logrus.Fields{
			"block": block,
		}).Info("Pow.Event: handle BlockMessage.")
	default:
		logging.VLog().WithFields(logrus.Fields{
			"eventType": e,
		}).Info("Pow.Event: handle this event.")
	}
}

// Transit transit state.
func (p *Pow) Transit(from, to consensus.State, data interface{}) {
	p.stateTransitionCh <- &stateTransitionArgs{from: from, to: to, data: data}
}

// VerifyBlock return nil if nonce is right, otherwise return error.
func (p *Pow) VerifyBlock(block *core.Block) error {
	if block == nil {
		logging.VLog().WithFields(logrus.Fields{
			"func": "Pow.VerifyBlock",
			"err":  ErrInvalidDataType,
		}).Error("data is not valid block")
		return ErrInvalidDataType
	}

	ret := HashAndVerify(block)
	if ret == nil {
		logging.VLog().WithFields(logrus.Fields{
			"func":  "Pow.VerifyBlock",
			"err":   ErrInvalidBlockNonce,
			"block": block,
		}).Error("invalid block nonce.")
		return ErrInvalidBlockNonce
	}

	return nil
}

func (p *Pow) checkValidTransit(from, to consensus.State) bool {
	valid := from != nil && to != nil && from != to && p.currentState == from
	logging.VLog().WithFields(logrus.Fields{
		"func":    "Pow.CheckTransit",
		"success": valid,
		"current": p.currentState,
		"from":    from,
		"to":      to,
	}).Info("State Transition.")
	return valid
}

func (p *Pow) stateLoop() {
	p.currentState.Enter(nil)

	for {
		select {
		case args := <-p.stateTransitionCh:
			to := args.to
			data := args.data
			from := args.from

			if !p.checkValidTransit(from, to) {
				continue
			}

			p.currentState.Leave(data)
			p.currentState = to
			p.currentState.Enter(data)

		case <-p.quitCh:
			logging.VLog().Info("quit Pow.loop.")
			return
		}
	}
}

func (p *Pow) blockLoop() {
	for {
		select {
		case block := <-p.chain.BlockPool().ReceivedLinkedBlockCh():
			p.newBlockReceived = true
			p.Event(consensus.NewBaseEvent(consensus.NewBlockEvent, block))
		case <-p.quitCh:
			logging.VLog().Info("Pow.blockLoop: quit.")
			return
		}
	}
}

func (p *Pow) resetMiningStatus() {
	if p.miningBlock != nil && !p.miningBlock.Sealed() {
		p.miningBlock.ReturnTransactions()
		p.miningBlock = nil
	}
	p.newBlockReceived = false
}
