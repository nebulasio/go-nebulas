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

	"github.com/nebulasio/go-nebulas/account"

	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/net"

	"github.com/hashicorp/golang-lru"
	"github.com/nebulasio/go-nebulas/consensus"
	"github.com/nebulasio/go-nebulas/neblet/pb"
	"github.com/nebulasio/go-nebulas/net/p2p"
	log "github.com/sirupsen/logrus"
)

// Errors in PoW Consensus
var (
	ErrInvalidDataType   = errors.New("invalid data type, should be *core.Block")
	ErrInvalidBlockNonce = errors.New("invalid block nonce")
	ErrDuplicateBlock    = errors.New("dup block from block pool")
	ErrInvalidPoDConfig  = errors.New("invalid pod config")
)

// Neblet interface breaks cycle import dependency and hides unused services.
type Neblet interface {
	Config() nebletpb.Config
	BlockChain() *core.BlockChain
	NetService() *p2p.NetService
	AccountManager() *account.Manager
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
	neblet   Neblet

	createdStateMachines  *lru.Cache
	creatingStateMachines *lru.Cache

	nonce              uint64
	votesTransactionCh chan *core.Transaction

	canMining bool
}

// NewPoD create PoD instance.
func NewPoD(neblet Neblet) *PoD {
	cfg := neblet.Config().Pod
	if cfg == nil {
		panic("PoD.NewPoD: cannot find pod config.")
	}
	coinbase, err := core.AddressParse(cfg.GetCoinbase())
	if err != nil {
		panic("PoD.NewPoD: coinbase parse err.")
	}
	createdStateMachines, err1 := lru.New(1024)
	creatingStateMachines, err2 := lru.New(1024)
	if err1 != nil || err2 != nil {
		panic("Pow.NewPow: fail to create lru cache for state machines.")
	}

	p := &PoD{
		quitCh: make(chan bool),

		chain:    neblet.BlockChain(),
		nm:       neblet.NetService(),
		neblet:   neblet,
		coinbase: coinbase,

		createdStateMachines:  createdStateMachines,
		creatingStateMachines: creatingStateMachines,
		votesTransactionCh:    make(chan *core.Transaction, 1024),
		nonce:                 0,

		canMining: false,
	}

	tail := p.chain.TailBlock()
	creatingStateMachine := p.newCreatingStateMachine(tail)
	p.creatingStateMachines.Add(tail.Hash().Hex(), creatingStateMachine)
	return p
}

func (p *PoD) newCreatingStateMachine(parent *core.Block) *consensus.StateMachine {
	creatingStateMachine := consensus.NewStateMachine(p)
	context, err := NewCreatingContext(parent, p.chain.TailBlock())
	if err != nil {
		log.WithFields(log.Fields{
			"func":    "PoD.BlockLoop",
			"channel": "New Block",
			"parent":  parent,
			"err":     err,
		}).Error("cannot create new state machine")
		panic("cannot create new state machine")
	}
	creatingStateMachine.SetInitialState(NewCreationState(creatingStateMachine, context))
	log.WithFields(log.Fields{
		"func":   "PoD.newCreatingStateMachine",
		"parent": parent,
	}).Info("create new creating state machine")
	return creatingStateMachine
}

func (p *PoD) newCreatedStateMachine(block *core.Block) *consensus.StateMachine {
	onCanonical := block.Hash().Equals(p.chain.TailBlock().Hash())
	createdStateMachine := consensus.NewStateMachine(p)
	context, err := NewCreatedContext(block, onCanonical)
	if err != nil {
		log.WithFields(log.Fields{
			"func":    "PoD.BlockLoop",
			"channel": "New Block",
			"block":   block,
			"err":     err,
		}).Error("cannot create new state machine")
		panic("cannot create new state machine")
	}
	createdStateMachine.SetInitialState(NewPrepareState(createdStateMachine, context))
	log.WithFields(log.Fields{
		"func":   "PoD.newCreatedStateMachine",
		"parent": block,
	}).Info("create new created state machine")
	createdStateMachine.Start()
	return createdStateMachine
}

// Start start pod service.
func (p *PoD) Start() {
	for _, key := range p.creatingStateMachines.Keys() {
		if stateMachine, ok := p.creatingStateMachines.Get(key); ok {
			stateMachine.(*consensus.StateMachine).Start()
		}
	}
	go p.blockLoop()
}

// Stop stop pow service.
func (p *PoD) Stop() {
	p.quitCh <- true
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
			log.Info("PoD Receive Block. ", block)
			// unlock coinbase
			if err := p.neblet.AccountManager().Unlock(p.coinbase, []byte("passphrase")); err != nil {
				log.WithFields(log.Fields{
					"func":    "PoD.BlockLoop",
					"channel": "New Blocks",
					"err":     err,
				}).Error("unlock address failed.")
				continue
			}
			// extract all votes
			for _, tx := range block.Transactions() {
				if tx.DataType() == core.TxPayloadVoteType {
					p.votesTransactionCh <- tx
				}
			}
			// fork choice
			log.Info("PoD Fork Choice.")
			p.ForkChoice(block)
			// new block event
			event := consensus.NewBaseEvent(NewBlockEvent, block)
			if stateMachine, exist := p.creatingStateMachines.Get(block.ParentHash().Hex()); exist {
				stateMachine.(*consensus.StateMachine).Event(event)
			} else {
				createdStateMachine := p.newCreatedStateMachine(block)
				p.createdStateMachines.Add(block.Hash().Hex(), createdStateMachine)
			}
			// create a creating state machine waiting for next block
			creatingStateMachine := p.newCreatingStateMachine(block)
			p.creatingStateMachines.Add(block.Hash().Hex(), creatingStateMachine)
			creatingStateMachine.Start()
		case voteTx := <-p.votesTransactionCh:
			// parse vote tx
			votePayload, err := core.LoadVotePayload(voteTx.Data())
			if err != nil {
				log.WithFields(log.Fields{
					"func":    "PoD.BlockLoop",
					"channel": "Votes Transaction",
					"err":     err,
				}).Error("invalid vote payload")
				continue
			}
			log.Info("PoD Receive Tx. ", votePayload)
			// unlock coinbase
			if err := p.neblet.AccountManager().Unlock(p.coinbase, []byte("passphrase")); err != nil {
				log.WithFields(log.Fields{
					"func":    "PoD.BlockLoop",
					"channel": "New Blocks",
					"err":     err,
				}).Error("unlock address failed.")
				continue
			}
			var action consensus.EventType
			var stateMachines *lru.Cache
			switch votePayload.Action {
			case core.PrepareAction:
				action = NewPrepareVoteEvent
				stateMachines = p.createdStateMachines
			case core.CommitAction:
				action = NewCommitVoteEvent
				stateMachines = p.createdStateMachines
			case core.ChangeAction:
				action = NewChangeVoteEvent
				stateMachines = p.creatingStateMachines
			case core.AbdicateAction:
				action = NewAbdicateVoteEvent
				stateMachines = p.creatingStateMachines
			}
			event := consensus.NewBaseEvent(action, voteTx.From().Bytes())
			if stateMachine, exist := stateMachines.Get(votePayload.Hash.Hex()); exist {
				stateMachine.(*consensus.StateMachine).Event(event)
				continue
			}
			log.WithFields(log.Fields{
				"func":    "PoD.BlockLoop",
				"channel": "Votes Transaction",
				"action":  votePayload.Action,
				"hash":    votePayload.Hash,
			}).Error("cannot find the related state machine")
		case <-p.quitCh:
			// unlock coinbase
			p.neblet.AccountManager().Lock(p.coinbase)
			log.Info("PoD.blockLoop: quit.")
			return
		}
	}
}
