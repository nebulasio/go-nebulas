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

	"github.com/nebulasio/go-nebulas/crypto"
	"github.com/nebulasio/go-nebulas/crypto/keystore"

	"github.com/nebulasio/go-nebulas/account"

	"github.com/nebulasio/go-nebulas/common/trie"
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/neblet/pb"
	"github.com/nebulasio/go-nebulas/net/p2p"

	"github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

// Errors in PoW Consensus
var (
	ErrInvalidBlockInterval = errors.New("invalid block interval")
	ErrMissingConfigForDpos = errors.New("missing configuration for Dpos")
	ErrInvalidBlockProposer = errors.New("invalid block proposer")
	ErrCannotMintBlockNow   = errors.New("cannot mint block now, waiting for sync over")
)

// Neblet interface breaks cycle import dependency and hides unused services.
type Neblet interface {
	Config() nebletpb.Config
	BlockChain() *core.BlockChain
	NetManager() p2p.Manager
	AccountManager() *account.Manager
}

// Dpos Delegate Proof-of-Stake
type Dpos struct {
	quitCh chan bool

	chain *core.BlockChain
	nm    p2p.Manager
	am    *account.Manager

	coinbase   *core.Address
	miner      *core.Address
	passphrase string

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
		nm:    neblet.NetManager(),
		am:    neblet.AccountManager(),

		blockInterval:   core.BlockInterval,
		dynastyInterval: core.DynastyInterval,
		txsPerBlock:     2000,

		canMining: false,
	}

	config := neblet.Config().Chain
	coinbase, err := core.AddressParse(config.Coinbase)
	if err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"address": config.Coinbase,
			"err":     err,
		}).Error("Failed to parse coinbase address.")
		return nil, err
	}
	miner, err := core.AddressParse(config.Miner)
	if err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"address": config.Miner,
			"err":     err,
		}).Error("Failed to parse miner address.")
		return nil, err
	}
	p.coinbase = coinbase
	p.miner = miner
	p.passphrase = config.Passphrase
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

func less(a *core.Block, b *core.Block) bool {
	if a.Height() != b.Height() {
		return a.Height() < b.Height()
	}
	return core.Less(a, b)
}

// do fork choice
func (p *Dpos) forkChoice() {
	bc := p.chain
	tailBlock := bc.TailBlock()
	detachedTailBlocks := bc.DetachedTailBlocks()

	// find the max depth.
	newTailBlock := tailBlock

	for _, v := range detachedTailBlocks {
		if less(newTailBlock, v) {
			newTailBlock = v
		}
	}

	if newTailBlock.Hash().Equals(tailBlock.Hash()) {
		logging.CLog().WithFields(logrus.Fields{
			"old tail": tailBlock,
			"new tail": newTailBlock,
		}).Info("Same blocks, no need to change.")
	} else {
		err := bc.SetTailBlock(newTailBlock)
		if err != nil {
			logging.CLog().WithFields(logrus.Fields{
				"new tail": newTailBlock,
				"old tail": tailBlock,
				"err":      err,
			}).Error("Failed to set new tail block.")
		} else {
			logging.CLog().WithFields(logrus.Fields{
				"new tail": newTailBlock,
				"old tail": tailBlock,
			}).Info("change to new tail.")
		}
	}
}

// CanMining return if consensus can do mining now
func (p *Dpos) CanMining() bool {
	return p.canMining
}

// SetCanMining set if consensus can do mining now
func (p *Dpos) SetCanMining(canMining bool) {
	if canMining {
		logging.CLog().Info("Start Dpos Mining.")
	} else {
		logging.CLog().Info("Stop Dpos Mining.")
	}
	p.canMining = canMining
}

func verifyBlockSign(miner *core.Address, block *core.Block) error {
	signature, err := crypto.NewSignature(keystore.Algorithm(block.Alg()))
	if err != nil {
		return err
	}
	pub, err := signature.RecoverPublic(block.Hash(), block.Signature())
	if err != nil {
		return err
	}
	pubdata, err := pub.Encoded()
	if err != nil {
		return err
	}
	addr, err := core.NewAddressFromPublicKey(pubdata)
	if err != nil {
		return err
	}
	if !miner.Equals(addr) {
		logging.VLog().WithFields(logrus.Fields{
			"recover address": addr.String(),
			"block":           block,
		}).Error("Failed to verify block's sign.")
		return ErrInvalidBlockProposer
	}
	block.SetMiner(miner)
	return nil
}

// FastVerifyBlock verify the block before its parent found
// can be verified if the block's dynasty == tail's dynasty
// can be verified if the block's dynasty == tails's next dynasty
func (p *Dpos) FastVerifyBlock(block *core.Block) error {
	tail := p.chain.TailBlock()
	// check timestamp
	elapsedSecond := block.Timestamp() - tail.Timestamp()
	if elapsedSecond%p.blockInterval != 0 {
		return ErrInvalidBlockInterval
	}
	// check proposer
	currentHour := block.Timestamp() / core.DynastyInterval
	tailHour := tail.Timestamp() / core.DynastyInterval
	var dynastyRoot byteutils.Hash
	if currentHour == tailHour {
		dynastyRoot = tail.DposContext().DynastyRoot
	} else if currentHour == tailHour+1 {
		dynastyRoot = tail.DposContext().NextDynastyRoot
	} else {
		return nil
	}
	dynasty, err := trie.NewBatchTrie(dynastyRoot, p.chain.Storage())
	if err != nil {
		return err
	}
	proposer, err := core.FindProposer(block.Timestamp(), dynasty)
	if err != nil {
		return err
	}
	miner, err := core.AddressParseFromBytes(proposer)
	if err != nil {
		return err
	}
	return verifyBlockSign(miner, block)
}

// VerifyBlock verify the block with its parent found
func (p *Dpos) VerifyBlock(block *core.Block, parent *core.Block) error {
	// check proposer
	dynasty, err := trie.NewBatchTrie(block.DposContext().DynastyRoot, block.Storage())
	if err != nil {
		return err
	}
	proposer, err := core.FindProposer(block.Timestamp(), dynasty)
	if err != nil {
		return err
	}
	miner, err := core.AddressParseFromBytes(proposer)
	if err != nil {
		return err
	}
	err = verifyBlockSign(miner, block)
	if err != nil {
		return err
	}
	return nil
}

func (p *Dpos) mintBlock(now int64) error {
	// check can do mining
	if !p.canMining {
		logging.VLog().WithFields(logrus.Fields{
			"now": now,
		}).Warn("Sync is not over yet.")
		return ErrCannotMintBlockNow
	}

	// check proposer
	tail := p.chain.TailBlock()
	elapsedSecond := now - tail.Timestamp()
	context, err := tail.NextDynastyContext(elapsedSecond)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"tail":    tail,
			"elapsed": elapsedSecond,
			"err":     err,
		}).Error("Failed to generate next dynasty context.")
		return core.ErrGenerateNextDynastyContext
	}
	if context.Proposer == nil || !context.Proposer.Equals(p.miner.Bytes()) {
		proposer := "nil"
		if context.Proposer != nil {
			proposer = string(context.Proposer.Hex())
		}
		logging.VLog().WithFields(logrus.Fields{
			"tail":     tail,
			"elapsed":  elapsedSecond,
			"expected": proposer,
			"actual":   p.miner.String(),
		}).Info("Not my turn, waiting...")
		return ErrInvalidBlockProposer
	}
	logging.VLog().WithFields(logrus.Fields{
		"tail":     tail,
		"elapsed":  elapsedSecond,
		"expected": context.Proposer.Hex(),
		"actual":   p.coinbase.String(),
	}).Info("My turn to mint block")

	// mint new block
	block, err := core.NewBlock(p.chain.ChainID(), p.coinbase, tail)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"tail":     tail,
			"coinbase": p.coinbase,
			"chainid":  p.chain.ChainID(),
			"err":      err,
		}).Error("Failed to create new block")
		return err
	}
	block.LoadDynastyContext(context)
	block.CollectTransactions(p.txsPerBlock)
	block.SetMiner(p.miner)
	if err = block.Seal(); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"block": block,
			"err":   err,
		}).Error("Failed to seal new block")
		return err
	}
	// TODO: move passphrase from config to console
	if err = p.am.Unlock(p.miner, []byte(p.passphrase)); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"miner": p.miner.String(),
			"err":   err,
		}).Error("Failed to unlock the miner")
		return err
	}
	if err = p.am.SignBlock(p.miner, block); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"miner": p.miner.String(),
			"block": block,
			"err":   err,
		}).Error("Failed to sign new block")
		return err
	}
	// broadcast it
	err = p.chain.BlockPool().PushAndBroadcast(block)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"tail":  tail,
			"block": block,
			"err":   err,
		}).Error("Failed to broadcast new block")
		return err
	}

	logging.VLog().WithFields(logrus.Fields{
		"tail":  tail,
		"block": block,
	}).Info("Minted new block")
	return nil
}

func (p *Dpos) blockLoop() {
	logging.CLog().Info("Launched Dpos Mining.")

	timeChan := time.NewTicker(time.Second).C
	for {
		select {
		case now := <-timeChan:
			p.mintBlock(now.Unix())
		case <-p.chain.BlockPool().ReceivedLinkedBlockCh():
			p.forkChoice()
		case <-p.quitCh:
			logging.CLog().Info("Shutdowned Dpos Mining.")
			return
		}
	}
}
