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

package dpos

import (
	"errors"
	"time"

	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/net"

	"github.com/nebulasio/go-nebulas/neblet/pb"
	"github.com/nebulasio/go-nebulas/net/p2p"
	log "github.com/sirupsen/logrus"
)

// Errors in PoW Consensus
var (
	ErrInvalidBlockInterval = errors.New("invalid block interval")
)

// Neblet interface breaks cycle import dependency and hides unused services.
type Neblet interface {
	Config() nebletpb.Config
	BlockChain() *core.BlockChain
	NetService() *p2p.NetService
}

// Dpos Delegate Proof-of-Stake
type Dpos struct {
	quitCh chan bool

	chain    *core.BlockChain
	nm       net.Manager
	coinbase *core.Address

	blockInterval   int64
	dynastyInterval int
	txsPerBlock     int

	canMining bool
}

// NewDpos create Dpos instance.
func NewDpos(neblet Neblet) *Dpos {
	p := &Dpos{
		quitCh: make(chan bool, 5),

		chain: neblet.BlockChain(),
		nm:    neblet.NetService(),

		blockInterval:   3,
		dynastyInterval: 3600,
		txsPerBlock:     500,

		canMining: false,
	}

	cfg := neblet.Config().Dpos
	if cfg != nil {
		coinbase, err := core.AddressParse(cfg.GetCoinbase())
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Info("Dpos.NewDpos: coinbase parse err.")
			panic("coinbase should be correctly configed for Dpos, expect hex string without 0x prefix")
		}
		p.coinbase = coinbase
	} else {
		panic("coinbase should be configed for Dpos")
	}
	return p
}

// Start start pow service.
func (p *Dpos) Start() {
	go p.blockLoop()
}

// Stop stop pow service.
func (p *Dpos) Stop() {
	p.quitCh <- true
}

// ForkChoice do fork choice
func (p *Dpos) ForkChoice() {
	bc := p.chain
	tailBlock := bc.TailBlock()
	detachedTailBlocks := bc.DetachedTailBlocks()

	// find the max depth.
	log.WithFields(log.Fields{
		"func": "Dpos.ForkChoice",
	}).Debug("find the highest tail.")

	newTailBlock := tailBlock
	maxHeight := tailBlock.Height()

	for _, v := range detachedTailBlocks {
		h := v.Height()
		if h > maxHeight {
			maxHeight = h
			newTailBlock = v
		}
	}

	if newTailBlock == bc.TailBlock() {
		log.WithFields(log.Fields{
			"func": "Dpos.ForkChoice",
		}).Info("current tail is the highest, no change.")
	} else {
		log.WithFields(log.Fields{
			"func":      "Dpos.ForkChoice",
			"maxHeight": maxHeight,
			"tailBlock": newTailBlock,
		}).Info("change to new tail.")
		bc.SetTailBlock(newTailBlock)
	}
}

// CanMining return if consensus can do mining now
func (p *Dpos) CanMining() bool {
	return p.canMining
}

// SetCanMining set if consensus can do mining now
func (p *Dpos) SetCanMining(canMining bool) {
	log.WithFields(log.Fields{
		"func":  "Dpos.SetCanMining",
		"start": canMining,
	}).Info("control mining.")
	p.canMining = canMining
}

// VerifyBlock return nil if nonce is right, otherwise return error.
func (p *Dpos) VerifyBlock(block *core.Block) error {
	parent, err := block.ParentBlock()
	if err != nil {
		return err
	}
	elapsedSecond := parent.Timestamp() - block.Timestamp()
	if elapsedSecond%p.blockInterval != 0 {
		return ErrInvalidBlockInterval
	}
	return nil
}

func (p *Dpos) mintBlock() {
	// check can do mining
	if !p.canMining {
		return
	}
	tail := p.chain.TailBlock()
	// calculate time elapsed
	elapsedSecond := time.Now().Unix() - tail.Timestamp()
	// check whether it's my turn
	proposer := tail.NextProposer(elapsedSecond)
	if !proposer.Equals(p.coinbase) {
		return
	}
	// mint new block
	block := core.NewBlock(p.chain.ChainID(), p.coinbase, tail)
	block.CollectTransactions(p.txsPerBlock)
	block.Seal()
	// broadcast it
	p.chain.BlockPool().PushAndBroadcast(block)
}

func (p *Dpos) blockLoop() {
	timeChan := time.NewTimer(time.Second).C
	for {
		select {
		case <-timeChan:
			go p.mintBlock()
		case <-p.chain.BlockPool().ReceivedBlockCh():
			go p.ForkChoice()
		case <-p.quitCh:
			log.Info("Dpos.blockLoop: quit.")
			return
		}
	}
}
