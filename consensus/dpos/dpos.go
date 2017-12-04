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
	"github.com/nebulasio/go-nebulas/neblet/pb"
	"github.com/nebulasio/go-nebulas/net"

	"github.com/nebulasio/go-nebulas/net/p2p"
	log "github.com/sirupsen/logrus"
)

// Errors in PoW Consensus
var (
	ErrInvalidBlockInterval = errors.New("invalid block interval")
	ErrInvalidBlockCoinbase = errors.New("invalid block proposer")
	ErrMissingConfigForDpos = errors.New("missing configuration for Dpos")
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
	dynastyInterval int64
	txsPerBlock     int

	canMining bool
}

// NewDpos create Dpos instance.
func NewDpos(neblet Neblet) (*Dpos, error) {
	p := &Dpos{
		quitCh: make(chan bool, 5),

		chain: neblet.BlockChain(),
		nm:    neblet.NetService(),

		blockInterval:   core.BlockInterval,
		dynastyInterval: core.DynastyInterval,
		txsPerBlock:     500,

		canMining: false,
	}

	cfg := neblet.Config().Dpos
	if cfg == nil {
		return nil, ErrMissingConfigForDpos
	}
	coinbase, err := core.AddressParse(cfg.GetCoinbase())
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("Dpos.NewDpos: coinbase parse err.")
		return nil, err
	}
	p.coinbase = coinbase
	return p, nil
}

// Start start pow service.
func (p *Dpos) Start() {
	go p.blockLoop()
}

// Stop stop pow service.
func (p *Dpos) Stop() {
	p.quitCh <- true
}

// do fork choice
func (p *Dpos) forkChoice() {
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

// VerifyBlock return nil if timestamp & proposer are right, otherwise return error.
func (p *Dpos) VerifyBlock(block *core.Block) error {
	parent, err := block.ParentBlock()
	if err != nil {
		return err
	}
	// check timestamp
	elapsedSecond := block.Timestamp() - parent.Timestamp()
	if elapsedSecond%p.blockInterval != 0 {
		return ErrInvalidBlockInterval
	}
	// check proposer
	context, err := parent.NextDynastyContext(elapsedSecond)
	if err != nil {
		return err
	}
	if !context.Proposer.Equals(block.Coinbase().Bytes()) {
		return ErrInvalidBlockCoinbase
	}
	return nil
}

func (p *Dpos) mintBlock() {
	// check can do mining
	if !p.canMining {
		return
	}
	// check proposer
	tail := p.chain.TailBlock()
	elapsedSecond := time.Now().Unix() - tail.Timestamp()
	context, err := tail.NextDynastyContext(elapsedSecond)
	if err != nil {
		return
	}
	if !context.Proposer.Equals(p.coinbase.Bytes()) {
		log.WithFields(log.Fields{
			"func":     "Dpos.mintBlock",
			"tail":     tail,
			"elapsed":  elapsedSecond,
			"expected": context.Proposer.Hex(),
			"actual":   p.coinbase.ToHex(),
		}).Info("not my turn, waiting...")
		return
	}
	log.WithFields(log.Fields{
		"func":     "Dpos.mintBlock",
		"tail":     tail,
		"elapsed":  elapsedSecond,
		"offset":   context.Offset,
		"expected": context.Proposer.Hex(),
		"actual":   p.coinbase.ToHex(),
	}).Info("my turn")
	// mint new block
	block := core.NewBlock(p.chain.ChainID(), p.coinbase, tail)
	block.LoadDynastyContext(context)
	block.CollectTransactions(p.txsPerBlock)
	block.Seal()
	// broadcast it
	p.chain.BlockPool().PushAndBroadcast(block)
}

func (p *Dpos) blockLoop() {
	timeChan := time.NewTicker(time.Second).C
	for {
		select {
		case <-timeChan:
			p.mintBlock()
		case <-p.chain.BlockPool().ReceivedBlockCh():
			p.forkChoice()
		case <-p.quitCh:
			log.Info("Dpos.blockLoop: quit.")
			return
		}
	}
}
