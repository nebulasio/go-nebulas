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

package pod

import (
	"errors"

	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/net"

	"github.com/hashicorp/golang-lru"
	"github.com/nebulasio/go-nebulas/consensus"
	"github.com/nebulasio/go-nebulas/core/state"
	"github.com/nebulasio/go-nebulas/neblet/pb"
	"github.com/nebulasio/go-nebulas/net/p2p"
	log "github.com/sirupsen/logrus"
)

// Errors in PoW Consensus
var (
	ErrInvalidDataType   = errors.New("invalid data type, should be *core.Block")
	ErrInvalidBlockNonce = errors.New("invalid block nonce")
	ErrInvalidPoDConfig  = errors.New("invalid pod config")
)

// Neblet interface breaks cycle import dependency and hides unused services.
type Neblet interface {
	Config() nebletpb.Config
	BlockChain() *core.BlockChain
	NetService() *p2p.NetService
}

/*
PoD implementation of Proof-of-Devotion consensus, designed to be a state machine.
The following is the state diagram:

@startuml pod
[*] --> Nomal
[*] --> Dumb
state Nomal {
    [*] --> Creation
    Creation: confirm current validators set
    Creation --> Prepare: receive a valid block
    Creation --> Commit: create a new block[-1x for each validator] & \nvote prepare default [+1x to me]
    Prepare: vote prepare [+1x to me]
    Prepare: collect others' prepare votes
    Prepare --> Prepare: receive prepare vote [+1x to sender]
    Commit: vote commit
    Commit: collect others' commit votes
    Prepare --> Commit: > 2/3 prepare votes
    Commit --> Commit: receive commit vote
    Finality: final this block
    Commit --> Finality: > 2/3 commit votes [+1.5x to all current validators]
    Finality --> [*]
}
state Dumb {
    [*] --> _Creation
    _Creation: confirm current validators set
    _Creation --> Change: timeout(12s)
    Change: vote Change
    Change: collect change votes
    Change --> Change: receive change vote
    Change --> _Creation: > 2/3 change votes \n[remove the dumb proposer &\n change to new proposer]
    Abdication: vote Abdication
    Abdication: collect abdication votes
    Abdication --> Abdication: receive abdication vote
    Change --> Abdication: timeout(120s)
    Abdication --> _Creation: > 1/3 abdication votes \n[remove the validator]
}
Nomal --> [*] : stop
Dumb --> [*]
@enduml
*/
type PoD struct {
	quitCh chan bool

	chain    *core.BlockChain
	nm       net.Manager
	coinbase *core.Address

	// contain many state machines
	// each block has a state machine
	stateMachineContainer *lru.Cache

	currentDynasty    uint64
	currentValidators map[string]state.Account

	canMining bool
}

// NewPoD create PoD instance.
func NewPoD(neblet Neblet) (*PoD, error) {
	stateMachineContainer, err := lru.New(1024)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Info("Pow.NewPow: fail to create lru cache for state machines.")
		return nil, err
	}
	p := &PoD{
		chain:                 neblet.BlockChain(),
		nm:                    neblet.NetService(),
		quitCh:                make(chan bool),
		stateMachineContainer: stateMachineContainer,
		canMining:             false,
	}

	cfg := neblet.Config().Pod
	if cfg != nil {
		coinbase, err := core.AddressParse(cfg.GetCoinbase())
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Info("PoD.NewPoD: coinbase parse err.")
			return nil, err
		}
		p.coinbase = coinbase
	} else {
		log.WithFields(log.Fields{
			"err": err,
		}).Info("PoD.NewPoD: cannot find pod config.")
		return nil, ErrInvalidPoDConfig
	}

	return p, nil
}

// Start start pod service.
func (p *PoD) Start() {
	go p.blockLoop()
}

// Stop stop pow service.
func (p *PoD) Stop() {
	// cleanup.
	p.quitCh <- true
	for _, key := range p.stateMachineContainer.Keys() {
		if stateMachine, ok := p.stateMachineContainer.Get(key); ok {
			stateMachine.(*consensus.StateMachine).Stop()
		}
	}
}

// CanMining return if consensus can do mining now
func (p *PoD) CanMining() bool {
	return p.canMining
}

// SetCanMining set if consensus can do mining now
func (p *PoD) SetCanMining(canMining bool) {
	log.Info("sync over, start mining")
	p.canMining = canMining
}

// VerifyBlock return nil if nonce is right, otherwise return error.
func (p *PoD) VerifyBlock(block *core.Block) error {
	if block == nil {
		log.WithFields(log.Fields{
			"func": "PoD.VerifyBlock",
			"err":  ErrInvalidDataType,
		}).Error("data is not valid block")
		return ErrInvalidDataType
	}
	return nil
}

func (p *PoD) blockLoop() {
	for {
		select {
		case block := <-p.chain.BlockPool().ReceivedBlockCh():
			log.Debugf("PoD.blockLoop: new block message received. %v", block)
			sm := consensus.NewStateMachine()
			sm.SetInitialState(NewCreationState(sm, p))
			p.stateMachineContainer.Add(sm)
			sm.Start()
		case tx := <-p.chain.TransactionPool().ReceivedTransactionCh():
			log.Debugf("PoD.blockLoop: new transaction message received. %v", tx)
		case <-p.quitCh:
			log.Info("PoD.blockLoop: quit.")
			return
		}
	}
}
